// Copyright (c) 2021, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package schulze implements the Schulze preferential voting method.
package schulze

import (
	"cmp"
	"fmt"
	"iter"
	"runtime"
	"slices"
	"unsafe"
)

// Number defines a constraint for numeric types that can be used to store
// pairwise preference counters.
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// NewPreferences initializes a fixed size slice of the specified Number type
// that stores all pairwise preferences for voting. The resulting slice is
// supposed to be updated by the Vote function with Ballot preferences and read
// by the Compute or Winner function to order choices or find the winner.
func NewPreferences[N Number](choicesLength int) []N {
	return make([]N, choicesLength*choicesLength)
}

// ValidateChoices checks that the choices slice is not empty and contains no duplicate choices.
func ValidateChoices[C comparable](choices []C) error {
	if len(choices) == 0 {
		return ErrEmptyChoices
	}
	seen := make(map[C]struct{}, len(choices))
	for _, c := range choices {
		if _, ok := seen[c]; ok {
			return &DuplicateChoiceError[C]{Choice: c}
		}
		seen[c] = struct{}{}
	}
	return nil
}

// Ballot represents a single vote with ranked choices. Lowest number represents
// the highest rank. Not all choices have to be ranked and multiple choices can
// have the same rank. Ranks do not have to be in consecutive order.
type Ballot[C comparable] map[C]int

// Record represents a single vote with ranked choices. It is a list of Ballot
// values. The first ballot is the list with the first choices, the second
// ballot is the list with the second choices, and so on. The last ballot is the
// list of choices that are not ranked, which can be an empty list.
type Record[C comparable] [][]C

// Map transforms the choices in the Record using the provided mapping function.
func (r Record[C]) Map[NewC comparable](fn func(C) NewC) Record[NewC] {
	if r == nil {
		return nil
	}
	res := make(Record[NewC], len(r))
	for i, group := range r {
		newGroup := make([]NewC, len(group))
		for j, c := range group {
			newGroup[j] = fn(c)
		}
		res[i] = newGroup
	}
	return res
}

// Vote updates the preferences passed as the first argument with the Ballot
// values. A record of a complete and normalized preferences is returned that
// can be used to unvote.
func Vote[C comparable, N Number](preferences []N, choices []C, b Ballot[C]) (Record[C], error) {
	choicesCount := len(choices)
	if len(preferences) < choicesCount*choicesCount {
		return nil, &InvalidPreferencesError{
			Length: len(preferences),
			Wanted: choicesCount * choicesCount,
		}
	}

	ranks, hasUnrankedChoices, err := ballotRanks(choices, b)
	if err != nil {
		return nil, fmt.Errorf("ballot ranks: %w", err)
	}

	for rank, choices1 := range ranks {
		rest := ranks[rank+1:]
		for _, i := range choices1 {
			icc := int(i) * choicesCount
			for _, choices2 := range rest {
				for _, j := range choices2 {
					preferences[icc+int(j)] += 1
				}
			}
		}
	}

	ranksLen := len(ranks)

	// set diagonal values as the values of the column of the least ranked
	// choice to be able to have the correct preferences matrix when adding new
	// choices
	if hasUnrankedChoices {
		// treat the diagonal values as one of the unranked choices,
		// deprioritizing all choices except unranked as they are of the same
		if ranksLen > 0 {
			for _, choices1 := range ranks[:ranksLen-1] {
				for _, i := range choices1 {
					preferences[int(i)*choicesCount+int(i)] += 1
				}
			}
		}
	} else {
		// all choices are ranked, treat diagonal values as a single not ranked
		// choice, deprioritizing them for all existing choices
		for i := range choicesCount {
			preferences[i*choicesCount+i] += 1
		}
	}

	// prepare results capacity to avoid allocation on appending the potential
	// unranked choices
	resultsCap := ranksLen
	if !hasUnrankedChoices {
		resultsCap++
	}

	r := make([][]C, ranksLen, resultsCap)
	for rank, indexes := range ranks {
		r[rank] = make([]C, len(indexes))
		for k, index := range indexes {
			r[rank][k] = choices[index]
		}
	}

	if !hasUnrankedChoices {
		r = append(r, make([]C, 0))
	}

	return r, nil
}

// VoteRanked updates the preferences with a pre-ranked list of choices. Each
// element in rankedChoices is a slice of choices tied at that rank. Choices
// omitted from rankedChoices are considered unranked.
func VoteRanked[C comparable, N Number](preferences []N, choices []C, rankedChoices [][]C) (Record[C], error) {
	choicesCount := len(choices)
	if len(preferences) < choicesCount*choicesCount {
		return nil, &InvalidPreferencesError{
			Length: len(preferences),
			Wanted: choicesCount * choicesCount,
		}
	}

	ranks, hasUnrankedChoices, err := rankedChoicesRanks(choices, rankedChoices)
	if err != nil {
		return nil, fmt.Errorf("ranked choices ranks: %w", err)
	}

	for rank, choices1 := range ranks {
		rest := ranks[rank+1:]
		for _, i := range choices1 {
			icc := int(i) * choicesCount
			for _, choices2 := range rest {
				for _, j := range choices2 {
					preferences[icc+int(j)] += 1
				}
			}
		}
	}

	ranksLen := len(ranks)

	if hasUnrankedChoices {
		if ranksLen > 0 {
			for _, choices1 := range ranks[:ranksLen-1] {
				for _, i := range choices1 {
					preferences[int(i)*choicesCount+int(i)] += 1
				}
			}
		}
	} else {
		for i := range choicesCount {
			preferences[i*choicesCount+i] += 1
		}
	}

	resultsCap := ranksLen
	if !hasUnrankedChoices {
		resultsCap++
	}

	r := make([][]C, ranksLen, resultsCap)
	for rank, indexes := range ranks {
		r[rank] = make([]C, len(indexes))
		for k, index := range indexes {
			r[rank][k] = choices[index]
		}
	}

	if !hasUnrankedChoices {
		r = append(r, make([]C, 0))
	}

	return r, nil
}

// Unvote removes the Ballot values from the preferences.
func Unvote[C comparable, N Number](preferences []N, choices []C, r Record[C]) error {
	choicesCount := len(choices)
	if len(preferences) < choicesCount*choicesCount {
		return &InvalidPreferencesError{
			Length: len(preferences),
			Wanted: choicesCount * choicesCount,
		}
	}

	recordLength := len(r)
	if recordLength == 0 {
		return nil
	}

	for rank, choices1 := range r {
		rest := r[rank+1:]
		for _, choice1 := range choices1 {
			i := getChoiceIndex(choices, choice1)
			if i < 0 {
				continue
			}
			for _, choices2 := range rest {
				for _, choice2 := range choices2 {
					j := getChoiceIndex(choices, choice2)
					if j < 0 {
						continue
					}
					preferences[int(i)*choicesCount+int(j)] -= 1
				}
			}
		}
	}

	knownChoices := newBitset(uint64(choicesCount))
	rankedChoices := newBitset(uint64(choicesCount))
	// remove voting from the ranked choices of the Record
	//
	// it is essential to have the last record ballot as
	// unranked choices even if it is empty
	for _, choices1 := range r[:recordLength-1] {
		for _, choice1 := range choices1 {
			i := getChoiceIndex(choices, choice1)
			if i < 0 {
				continue
			}
			preferences[int(i)*choicesCount+int(i)] -= 1
			knownChoices.set(uint64(i))
			rankedChoices.set(uint64(i))
		}
	}
	// mark the rest of the known choices in the Record
	for _, choice1 := range r[recordLength-1] {
		i := getChoiceIndex(choices, choice1)
		if i < 0 {
			continue
		}
		knownChoices.set(uint64(i))
	}

	// remove votes of the choices that were added after the Record
	for i := range uint64(choicesCount) {
		if rankedChoices.isSet(i) {
			for j := range uint64(choicesCount) {
				if !knownChoices.isSet(j) {
					preferences[int(i)*choicesCount+int(j)] -= 1
				}
			}
		}
	}

	return nil
}

// SetChoices updates the preferences passed as the first argument by changing
// its values to accommodate the changes to the choices. It is required to
// pass the exact choices as the second parameter and complete updated choices
// as the third argument.
func SetChoices[C comparable, N Number](preferences []N, current, updated []C) []N {
	currentLength := len(current)
	updatedLength := len(updated)
	if len(preferences) < currentLength*currentLength {
		return nil
	}
	updatedPreferences := NewPreferences[N](updatedLength)

	var updatedInCurrent []int
	if currentLength <= 16 {
		updatedInCurrent = make([]int, updatedLength)
		for i, u := range updated {
			updatedInCurrent[i] = int(getChoiceIndex(current, u))
		}
	} else {
		currentIndex := make(map[C]int, currentLength)
		for i, c := range current {
			if _, ok := currentIndex[c]; !ok {
				currentIndex[c] = i
			}
		}
		updatedInCurrent = make([]int, updatedLength)
		for i, u := range updated {
			if idx, ok := currentIndex[u]; ok {
				updatedInCurrent[i] = idx
			} else {
				updatedInCurrent[i] = -1
			}
		}
	}

	for iUpdated := range updatedLength {
		iCurrent := updatedInCurrent[iUpdated]
		rowOffset := iUpdated * updatedLength
		for j := range updatedLength {
			if iUpdated < currentLength && j < currentLength && updated[iUpdated] == current[iUpdated] && updated[j] == current[j] {
				updatedPreferences[rowOffset+j] = preferences[iUpdated*currentLength+j]
			} else {
				if iCurrent >= 0 {
					jCurrent := updatedInCurrent[j]
					if jCurrent >= 0 {
						updatedPreferences[rowOffset+j] = preferences[iCurrent*currentLength+jCurrent]
					} else {
						// set the column of the new choice to the values of the
						// preferences' diagonal values, just as nobody voted for the
						// new choice to ensure consistency
						updatedPreferences[rowOffset+j] = preferences[iCurrent*currentLength+iCurrent]
					}
				}
			}
		}
	}
	return updatedPreferences
}

// Result represents a total number of wins for a single choice.
type Result[C comparable] struct {
	// The choice value.
	Choice C
	// 0-based ordinal number of the choice in the choice slice.
	Index int
	// Number of wins in pairwise comparisons to other choices votings.
	Wins int
	// Total number of votes in the weakest link of the strongest path in wins
	// in pairwise comparisons to other choices votings. Strength does not
	// effect the winner, and may be less then the Strength of the choice with
	// more wins.
	Strength int
	// Total number of preferred votes (difference between votes of the winner
	// choice and the opponent choice) in the weakest link of the strongest path
	// in wins in pairwise comparisons to other choices votings. Advantage does
	// not effect the winner, and may be less then the Advantage of the choice
	// with more wins. The code with less wins and greater Advantage had
	// stronger but fewer wins and that information can be taken into the
	// analysis of the results.
	Advantage int
}

// Map transforms the Choice in the Result using the provided mapping function.
func (res Result[C]) Map[NewC comparable](fn func(C) NewC) Result[NewC] {
	return Result[NewC]{
		Choice:    fn(res.Choice),
		Index:     res.Index,
		Wins:      res.Wins,
		Strength:  res.Strength,
		Advantage: res.Advantage,
	}
}

// Winner calculates the election winner. It first checks for a Condorcet
// winner in O(N^2) time; if none exists, it computes the Schulze beatpaths.
// If multiple choices tie for the winning position or there are no choices,
// tie is true.
func Winner[C comparable, N Number](preferences []N, choices []C) (winner Result[C], tie bool) {
	choicesCount := len(choices)
	if choicesCount == 0 || len(preferences) < choicesCount*choicesCount {
		return Result[C]{}, true
	}
	if choicesCount == 1 {
		return Result[C]{Choice: choices[0], Index: 0}, false
	}

	// Fast-path: O(N^2) Condorcet winner check
	for i := range choicesCount {
		isCondorcet := true
		var wins int
		var strength int
		var advantage int
		for j := range choicesCount {
			if i == j {
				continue
			}
			pij := preferences[i*choicesCount+j]
			pji := preferences[j*choicesCount+i]
			if pij > pji {
				wins++
				strength += int(pij)
				advantage += int(pij - pji)
			} else {
				isCondorcet = false
				break
			}
		}
		if isCondorcet && wins == choicesCount-1 {
			return Result[C]{
				Choice:    choices[i],
				Index:     i,
				Wins:      wins,
				Strength:  strength,
				Advantage: advantage,
			}, false
		}
	}

	// Fallback: Compute beatpaths for tournament cycles
	results, _, tie := Compute(preferences, choices)
	if len(results) == 0 {
		return Result[C]{}, true
	}
	return results[0], tie
}

// Compute calculates a sorted list of choices with the total number of wins for
// each of them by reading preferences data previously populated by the Vote
// function. If there are multiple winners, tie boolean parameter is true.
func Compute[C comparable, N Number](preferences []N, choices []C) (results []Result[C], duels DuelsIterator[C], tie bool) {
	strengths := calculatePairwiseStrengths(choices, preferences)
	results, tie = calculateResults(choices, strengths)
	return results, newDuelsIterator(choices, strengths), tie
}

// Duels returns an iter.Seq[*Duel[C]] standard iterator over pairwise duel outcomes.
func Duels[C comparable, N Number](preferences []N, choices []C) iter.Seq[*Duel[C]] {
	return func(yield func(*Duel[C]) bool) {
		strengths := calculatePairwiseStrengths(choices, preferences)
		it := newDuelsIterator(choices, strengths)
		for {
			d := it()
			if d == nil || !yield(d) {
				return
			}
		}
	}
}

// DuelsIterator is a function that returns the next Duel ordered by the choice indexes.
type DuelsIterator[C comparable] func() *Duel[C]

// Seq converts a DuelsIterator to a standard Go iter.Seq[*Duel[C]].
func (d DuelsIterator[C]) Seq() iter.Seq[*Duel[C]] {
	return func(yield func(*Duel[C]) bool) {
		for {
			duel := d()
			if duel == nil || !yield(duel) {
				return
			}
		}
	}
}

func newDuelsIterator[C comparable, N Number](choices []C, strengths []N) (duels DuelsIterator[C]) {
	choicesCount := len(choices)
	choiceIndexRow := 0
	choiceIndexColumn := 1

	return func() *Duel[C] {
		if choiceIndexRow >= choicesCount || choiceIndexColumn >= choicesCount {
			return nil
		}

		defer func() {
			choiceIndexColumn++
			if choiceIndexColumn >= choicesCount {
				choiceIndexRow++
				choiceIndexColumn = choiceIndexRow + 1
			}
		}()

		return &Duel[C]{
			Left: ChoiceStrength[C]{
				Choice:   choices[choiceIndexRow],
				Index:    choiceIndexRow,
				Strength: int(strengths[choiceIndexRow*choicesCount+choiceIndexColumn]),
			},
			Right: ChoiceStrength[C]{
				Choice:   choices[choiceIndexColumn],
				Index:    choiceIndexColumn,
				Strength: int(strengths[choiceIndexColumn*choicesCount+choiceIndexRow]),
			},
		}
	}
}

// Duel represents a pairwise comparison between two choices that are compared
// by their strongest paths strengths (number of votes in the weakest link of
// the strongest path).
type Duel[C comparable] struct {
	Left  ChoiceStrength[C]
	Right ChoiceStrength[C]
}

// Map transforms the Choices in the Duel using the provided mapping function.
func (d Duel[C]) Map[NewC comparable](fn func(C) NewC) Duel[NewC] {
	return Duel[NewC]{
		Left:  d.Left.Map(fn),
		Right: d.Right.Map(fn),
	}
}

// Outcome returns the winner and the defeated choice in the pairwise
// comparison of their strengths. If nils are returned, the outcome of the duel
// is a tie.
func (d Duel[C]) Outcome() (winner, defeated *ChoiceStrength[C]) {
	if d.Left.Strength > d.Right.Strength {
		return &d.Left, &d.Right
	}
	if d.Right.Strength > d.Left.Strength {
		return &d.Right, &d.Left
	}
	return nil, nil // tie
}

// ChoiceStrength stores the strength of a choice. The strength is the number of
// votes in the weakest link of the strongest path between votes for different
// choices.
type ChoiceStrength[C comparable] struct {
	// The choice value.
	Choice C
	// 0-based ordinal number of the choice in the choice slice.
	Index    int
	Strength int
}

// Map transforms the Choice in the ChoiceStrength using the provided mapping function.
func (cs ChoiceStrength[C]) Map[NewC comparable](fn func(C) NewC) ChoiceStrength[NewC] {
	return ChoiceStrength[NewC]{
		Choice:   fn(cs.Choice),
		Index:    cs.Index,
		Strength: cs.Strength,
	}
}

type choiceIndex int

func getChoiceIndex[C comparable](choices []C, choice C) choiceIndex {
	for i, c := range choices {
		if c == choice {
			return choiceIndex(i)
		}
	}
	return -1
}

func ballotRanks[C comparable](choices []C, b Ballot[C]) (ranks [][]choiceIndex, hasUnrankedChoices bool, err error) {
	choicesLen := len(choices)
	ballotLen := len(b)
	hasUnrankedChoices = ballotLen != choicesLen

	ballotRanks := make(map[int][]choiceIndex, ballotLen)
	var rankedChoices bitSet
	if hasUnrankedChoices {
		rankedChoices = newBitset(uint64(choicesLen))
	}

	for choice, rank := range b {
		index := getChoiceIndex(choices, choice)
		if index < 0 {
			return nil, false, &UnknownChoiceError[C]{Choice: choice}
		}
		ballotRanks[rank] = append(ballotRanks[rank], index)

		if hasUnrankedChoices {
			rankedChoices.set(uint64(index))
		}
	}

	rankNumbers := make([]int, 0, len(ballotRanks))
	for rank := range ballotRanks {
		rankNumbers = append(rankNumbers, rank)
	}

	slices.Sort(rankNumbers)

	if hasUnrankedChoices {
		ranks = make([][]choiceIndex, 0, len(rankNumbers)+1)
	} else {
		ranks = make([][]choiceIndex, 0, len(rankNumbers))
	}
	for _, rankNumber := range rankNumbers {
		ranks = append(ranks, ballotRanks[rankNumber])
	}

	if hasUnrankedChoices {
		unranked := make([]choiceIndex, 0, choicesLen-ballotLen)
		for i := uint64(0); int(i) < choicesLen; i++ {
			if !rankedChoices.isSet(i) {
				unranked = append(unranked, choiceIndex(i))
			}
		}
		if len(unranked) > 0 {
			ranks = append(ranks, unranked)
		}
	}

	return ranks, hasUnrankedChoices, nil
}

func rankedChoicesRanks[C comparable](choices []C, rankedChoices [][]C) (ranks [][]choiceIndex, hasUnrankedChoices bool, err error) {
	choicesLen := len(choices)
	rankedBitset := newBitset(uint64(choicesLen))

	totalRanked := 0
	ranks = make([][]choiceIndex, 0, len(rankedChoices)+1)

	for _, group := range rankedChoices {
		if len(group) == 0 {
			continue
		}
		groupIndexes := make([]choiceIndex, 0, len(group))
		for _, choice := range group {
			idx := getChoiceIndex(choices, choice)
			if idx < 0 {
				return nil, false, &UnknownChoiceError[C]{Choice: choice}
			}
			if rankedBitset.isSet(uint64(idx)) {
				return nil, false, &DuplicateChoiceError[C]{Choice: choice}
			}
			rankedBitset.set(uint64(idx))
			groupIndexes = append(groupIndexes, idx)
			totalRanked++
		}
		if len(groupIndexes) > 0 {
			ranks = append(ranks, groupIndexes)
		}
	}

	hasUnrankedChoices = totalRanked != choicesLen
	if hasUnrankedChoices {
		unranked := make([]choiceIndex, 0, choicesLen-totalRanked)
		for i := uint64(0); int(i) < choicesLen; i++ {
			if !rankedBitset.isSet(i) {
				unranked = append(unranked, choiceIndex(i))
			}
		}
		if len(unranked) > 0 {
			ranks = append(ranks, unranked)
		}
	}

	return ranks, hasUnrankedChoices, nil
}

func calculatePairwiseStrengths[C comparable, N Number](choices []C, preferences []N) []N {
	choicesCount := uintptr(len(choices))

	if choicesCount == 0 || len(preferences) < int(choicesCount*choicesCount) {
		return nil
	}

	strengths := make([]N, choicesCount*choicesCount)
	strengthsPtr := unsafe.Pointer(&strengths[0])
	elemSize := unsafe.Sizeof(N(0))

	for i := range choicesCount {
		icc := i * choicesCount

		for j := range choicesCount {
			ij := icc + j
			ji := j*choicesCount + i
			c := preferences[ij]

			if c > preferences[ji] {
				*(*N)(unsafe.Add(strengthsPtr, ij*elemSize)) = c
			}
		}
	}

	numWorkers := runtime.GOMAXPROCS(0)

	if numWorkers <= 1 || choicesCount < 64 {
		for i := range choicesCount {
			rowIPtr := unsafe.Add(strengthsPtr, i*choicesCount*elemSize)
			for j := range choicesCount {
				ji := j*choicesCount + i
				jip := *(*N)(unsafe.Add(strengthsPtr, ji*elemSize))
				if jip <= 0 {
					continue
				}
				rowJPtr := unsafe.Add(strengthsPtr, j*choicesCount*elemSize)
				relaxRow(rowIPtr, rowJPtr, jip, int(choicesCount))
			}
		}
		return strengths
	}

	chunkSize := (int(choicesCount) + numWorkers - 1) / numWorkers
	startChans := make([]chan uintptr, numWorkers)
	doneChans := make([]chan struct{}, numWorkers)

	for w := range numWorkers {
		startChans[w] = make(chan uintptr, 1)
		doneChans[w] = make(chan struct{}, 1)
		jStart := uintptr(w * chunkSize)
		jEnd := uintptr(min((w+1)*chunkSize, int(choicesCount)))

		go func(w int, jStart, jEnd uintptr) {
			for i := range startChans[w] {
				if i == ^uintptr(0) {
					return
				}
				icc := i * choicesCount
				for j := jStart; j < jEnd; j++ {
					ji := j*choicesCount + i
					jip := *(*N)(unsafe.Add(strengthsPtr, ji*elemSize))
					if jip <= 0 {
						continue
					}
					jcc := j * choicesCount
					const step = 8
					ccMod := choicesCount % step
					cc := choicesCount - ccMod
					end := cc + icc
					for ik, jk := icc, jcc; ik < end; ik, jk = ik+step, jk+step {
						setStrengthValueInlined(strengthsPtr, ik, jk, jip, elemSize)
						setStrengthValueInlined(strengthsPtr, ik+1, jk+1, jip, elemSize)
						setStrengthValueInlined(strengthsPtr, ik+2, jk+2, jip, elemSize)
						setStrengthValueInlined(strengthsPtr, ik+3, jk+3, jip, elemSize)
						setStrengthValueInlined(strengthsPtr, ik+4, jk+4, jip, elemSize)
						setStrengthValueInlined(strengthsPtr, ik+5, jk+5, jip, elemSize)
						setStrengthValueInlined(strengthsPtr, ik+6, jk+6, jip, elemSize)
						setStrengthValueInlined(strengthsPtr, ik+7, jk+7, jip, elemSize)
					}
					end = choicesCount + icc
					for ik, jk := cc+icc, cc+jcc; ik < end; ik, jk = ik+1, jk+1 {
						setStrengthValueInlined(strengthsPtr, ik, jk, jip, elemSize)
					}
				}
				doneChans[w] <- struct{}{}
			}
		}(w, jStart, jEnd)
	}

	for i := range choicesCount {
		for w := range numWorkers {
			startChans[w] <- i
		}
		for w := range numWorkers {
			<-doneChans[w]
		}
	}

	for w := range numWorkers {
		startChans[w] <- ^uintptr(0)
	}

	return strengths
}

func setStrengthValueInlined[N Number](strengthsPtr unsafe.Pointer, ik, jk uintptr, jip N, elemSize uintptr) {
	m := min(
		jip,
		*(*N)(unsafe.Add(strengthsPtr, ik*elemSize)),
	)

	jkp := (*N)(unsafe.Add(strengthsPtr, jk*elemSize))
	if m > *jkp {
		*jkp = m
	}
}

func calculateResults[C comparable, N Number](choices []C, strengths []N) (results []Result[C], tie bool) {
	choicesCount := len(choices)
	results = make([]Result[C], 0, choicesCount)

	for i := range choicesCount {
		var wins int
		var strength int
		var advantage int

		for j := range choicesCount {
			if i != j {
				sij := strengths[i*choicesCount+j]
				sji := strengths[j*choicesCount+i]
				if sij > sji {
					wins++
					strength += int(sij)
					advantage += int(sij - sji)
				}
			}
		}
		results = append(results, Result[C]{
			Choice:    choices[i],
			Index:     i,
			Wins:      wins,
			Strength:  strength,
			Advantage: advantage,
		})
	}

	slices.SortFunc(results, func(a, b Result[C]) int {
		if a.Wins != b.Wins {
			return cmp.Compare(b.Wins, a.Wins)
		}
		if a.Strength != b.Strength {
			return cmp.Compare(b.Strength, a.Strength)
		}
		return cmp.Compare(a.Index, b.Index)
	})

	if len(results) >= 2 {
		tie = results[0].Wins == results[1].Wins
	}

	return results, tie
}
