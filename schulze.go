// Copyright (c) 2021, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package schulze implements the Schulze preferential voting method.
package schulze

import (
	"fmt"
	"sort"
	"unsafe"
)

// NewPreferences initializes a fixed size slice that stores all pairwise
// preferences for voting. The resulting slice supposed to be updated by the
// Vote function with Ballot preferences and read by the Results function to
// order choices by their wins.
func NewPreferences(choicesLength int) []int {
	return make([]int, choicesLength*choicesLength)
}

// Ballot represents a single vote with ranked choices. Lowest number represents
// the highest rank. Not all choices have to be ranked and multiple choices can
// have the same rank. Ranks do not have to be in consecutive order.
type Ballot[C comparable] map[C]int

// Record represents a single vote with ranked choices. It is a list of Ballot
// values. The first ballot is the list with the first choices, the second
// ballot is the list with the second choices, and so on.
type Record[C comparable] [][]C

// Vote updates the preferences passed as the first argument with the Ballot
// values. A record of a complete and normalized preferences is returned that
// can be used to unvote.
func Vote[C comparable](preferences []int, choices []C, b Ballot[C]) (Record[C], error) {
	ranks, choicesCount, hasUnrankedChoices, err := ballotRanks(choices, b)
	if err != nil {
		return nil, fmt.Errorf("ballot ranks: %w", err)
	}

	for rank, choices1 := range ranks {
		rest := ranks[rank+1:]
		for _, i := range choices1 {
			for _, choices1 := range rest {
				for _, j := range choices1 {
					preferences[int(i)*choicesCount+int(j)] += 1
				}
			}
		}
	}

	// set diagonal values as the values of the column of the least ranked
	// choice to be able to have the correct preferences matrix when adding new
	// choices
	if hasUnrankedChoices {
		// treat the diagonal values as one of the unranked choices,
		// deprioritizing all choices except unranked as they are of the same
		if l := len(ranks); l > 0 {
			for _, choices1 := range ranks[:l-1] {
				for _, i := range choices1 {
					preferences[int(i)*choicesCount+int(i)] += 1
				}
			}
		}
	} else {
		// all choices are ranked, tread diagonal values as a single not ranked
		// choice, deprioritizing them for all existing choices
		for i := 0; i < choicesCount; i++ {
			preferences[int(i)*choicesCount+int(i)] += 1
		}
	}

	r := make([][]C, len(ranks))
	for rank, indexes := range ranks {
		if r[rank] == nil {
			r[rank] = make([]C, 0, len(indexes))
		}
		for _, index := range indexes {
			r[rank] = append(r[rank], choices[index])
		}
	}

	return r, nil
}

// Unvote removes the Ballot values from the preferences.
func Unvote[C comparable](preferences []int, choices []C, r Record[C]) error {
	choicesCount := len(choices)

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
			for _, choices1 := range rest {
				for _, choice2 := range choices1 {
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
	for i := uint64(0); int(i) < choicesCount; i++ {
		if rankedChoices.isSet(i) {
			for j := uint64(0); int(j) < choicesCount; j++ {
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
func SetChoices[C comparable](preferences []int, current, updated []C) []int {
	currentLength := len(current)
	updatedLength := len(updated)
	updatedPreferences := NewPreferences(updatedLength)
	for iUpdated := 0; iUpdated < updatedLength; iUpdated++ {
		iCurrent := int(getChoiceIndex(current, updated[iUpdated]))
		for j := 0; j < updatedLength; j++ {
			if iUpdated < currentLength && updated[iUpdated] == current[iUpdated] && j < currentLength && updated[j] == current[j] {
				updatedPreferences[iUpdated*updatedLength+j] = preferences[iUpdated*currentLength+j]
			} else {
				jCurrent := int(getChoiceIndex(current, updated[j]))
				if iCurrent >= 0 {
					if jCurrent >= 0 {
						updatedPreferences[iUpdated*updatedLength+j] = preferences[iCurrent*currentLength+jCurrent]
					} else {
						// set the column of the new choice to the values of the
						// preferences' diagonal values, just as nobody voted for the
						// new choice to ensure consistency
						updatedPreferences[iUpdated*updatedLength+j] = preferences[iCurrent*currentLength+iCurrent]
					}
				}
			}
		}
	}
	return updatedPreferences
}

type Choice[C comparable] struct {
	// The choice value.
	Value C
	// 0-based ordinal number of the choice in the choice slice.
	Index int
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

// Compute calculates a sorted list of choices with the total number of wins for
// each of them by reading preferences data previously populated by the Vote
// function. If there are multiple winners, tie boolean parameter is true.
func Compute[C comparable](preferences []int, choices []C) (results []Result[C], duels DuelsIterator[C], tie bool) {
	strengths := calculatePairwiseStrengths(choices, preferences)
	results, tie = calculateResults(choices, strengths)
	return results, newDuelsIterator(choices, strengths), tie
}

// DuelsIterator is a function that returns the next Duel ordered by the choice indexes.
type DuelsIterator[C comparable] func() *Duel[C]

func newDuelsIterator[C comparable](choices []C, strengths []int) (duels DuelsIterator[C]) {
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
				Strength: strengths[choiceIndexRow*choicesCount+choiceIndexColumn],
			},
			Right: ChoiceStrength[C]{
				Choice:   choices[choiceIndexColumn],
				Index:    choiceIndexColumn,
				Strength: strengths[choiceIndexColumn*choicesCount+choiceIndexRow],
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

// Outcome returns the the winner and the defeated choice in the pairwise
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

type choiceIndex int

func getChoiceIndex[C comparable](choices []C, choice C) choiceIndex {
	for i, c := range choices {
		if c == choice {
			return choiceIndex(i)
		}
	}
	return -1
}

func ballotRanks[C comparable](choices []C, b Ballot[C]) (ranks [][]choiceIndex, choicesLen int, hasUnrankedChoices bool, err error) {
	choicesLen = len(choices)
	ballotLen := len(b)
	hasUnrankedChoices = ballotLen != choicesLen

	ballotRanks := make(map[int][]choiceIndex, ballotLen)
	var rankedChoices bitSet
	if hasUnrankedChoices {
		rankedChoices = newBitset(uint64(choicesLen))
	}

	choicesLen = len(choices)

	for choice, rank := range b {
		index := getChoiceIndex(choices, choice)
		if index < 0 {
			return nil, 0, false, &UnknownChoiceError[C]{Choice: choice}
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

	sort.Slice(rankNumbers, func(i, j int) bool {
		return rankNumbers[i] < rankNumbers[j]
	})

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

	return ranks, choicesLen, hasUnrankedChoices, nil
}

const intSize = unsafe.Sizeof(int(0))

func calculatePairwiseStrengths[C comparable](choices []C, preferences []int) []int {
	choicesCount := uintptr(len(choices))

	if choicesCount == 0 {
		return nil
	}

	strengths := make([]int, choicesCount*choicesCount)

	strengthsPtr := unsafe.Pointer(&strengths[0])

	for i := uintptr(0); i < choicesCount; i++ {
		for j := uintptr(0); j < choicesCount; j++ {
			// removed unnecessary check for optimization: if i == j { continue }
			ij := i*choicesCount + j
			ji := j*choicesCount + i
			c := preferences[ij]
			if c > preferences[ji] {
				*(*int)(unsafe.Add(strengthsPtr, ij*intSize)) = c
			}
		}
	}

	for i := uintptr(0); i < choicesCount; i++ {
		for j := uintptr(0); j < choicesCount; j++ {
			// removed unnecessary check for optimization: if i == j { continue }
			ji := j*choicesCount + i
			jip := *(*int)(unsafe.Add(strengthsPtr, ji*intSize))
			for k := uintptr(0); k < choicesCount; k++ {
				// removed unnecessary check for optimization: if i == k || j == k { continue }
				jk := j*choicesCount + k
				ik := i*choicesCount + k
				jkp := (*int)(unsafe.Add(strengthsPtr, jk*intSize))
				m := max(
					*jkp,
					min(
						jip,
						*(*int)(unsafe.Add(strengthsPtr, ik*intSize)),
					),
				)
				if *jkp != m {
					*jkp = m
				}
			}
		}
	}

	return strengths
}

func calculateResults[C comparable](choices []C, strengths []int) (results []Result[C], tie bool) {
	choicesCount := len(choices)
	results = make([]Result[C], 0, choicesCount)

	for i := 0; i < choicesCount; i++ {
		var wins int
		var strength int
		var advantage int

		for j := 0; j < choicesCount; j++ {
			if i != j {
				sij := strengths[i*choicesCount+j]
				sji := strengths[j*choicesCount+i]
				if sij > sji {
					wins++
					strength += sij
					advantage += sij - sji
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

	sort.Slice(results, func(i, j int) bool {
		if results[i].Wins != results[j].Wins {
			return results[i].Wins > results[j].Wins
		}
		if results[i].Strength != results[j].Strength {
			return results[i].Strength > results[j].Strength
		}
		return results[i].Index < results[j].Index
	})

	if len(results) >= 2 {
		tie = results[0].Wins == results[1].Wins
	}

	return results, tie
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
