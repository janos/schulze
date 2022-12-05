// Copyright (c) 2021, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package schulze implements the Schulze preferential voting method.
package schulze

import (
	"fmt"
	"sort"
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

// Vote updates the preferences passed as the first argument with the Ballot
// values.
func Vote[C comparable](preferences []int, choices []C, b Ballot[C]) error {
	return vote(preferences, choices, b, 1) // add one to increment every pairwise preference
}

// Unvote removes the Ballot values from the preferences.
func Unvote[C comparable](preferences []int, choices []C, b Ballot[C]) error {
	return vote(preferences, choices, b, -1) // subtract one to decrement every pairwise preference
}

// vote updates the preferences with ballot values according to the passed
// choices. The weight is the value which is added to the preferences slice
// values for pairwise wins. If the weight is 1, the ballot is added, and if it
// is -1 the ballot is removed.
func vote[C comparable](preferences []int, choices []C, b Ballot[C], weight int) error {
	ranks, choicesCount, err := ballotRanks(choices, b)
	if err != nil {
		return fmt.Errorf("ballot ranks: %w", err)
	}

	for rank, choices1 := range ranks {
		rest := ranks[rank+1:]
		for _, i := range choices1 {
			for _, choices1 := range rest {
				for _, j := range choices1 {
					preferences[int(i)*choicesCount+int(j)] += weight
				}
			}
		}
	}

	return nil
}

// Result represents a total number of wins for a single choice.
type Result[C comparable] struct {
	// The choice value.
	Choice C
	// 0-based ordinal number of the choice in the choice slice.
	Index int
	// Number of wins in pairwise comparisons to other choices votings.
	Wins int
}

// Compute calculates a sorted list of choices with the total number of wins for
// each of them by reading preferences data previously populated by the Vote
// function. If there are multiple winners, tie boolean parameter is true.
func Compute[C comparable](preferences []int, choices []C) (results []Result[C], tie bool) {
	strengths := calculatePairwiseStrengths(choices, preferences)
	return calculateResults(choices, strengths)
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

func ballotRanks[C comparable](choices []C, b Ballot[C]) (ranks [][]choiceIndex, choicesLen int, err error) {
	choicesLen = len(choices)
	ballotLen := len(b)
	hasUnrankedChoices := ballotLen != choicesLen

	ballotRanks := make(map[int][]choiceIndex, ballotLen)
	var rankedChoices bitSet
	if hasUnrankedChoices {
		rankedChoices = newBitset(uint(choicesLen))
	}

	choicesLen = len(choices)

	for choice, rank := range b {
		index := getChoiceIndex(choices, choice)
		if index < 0 {
			return nil, 0, &UnknownChoiceError[C]{Choice: choice}
		}
		ballotRanks[rank] = append(ballotRanks[rank], index)

		if hasUnrankedChoices {
			rankedChoices.set(uint(index))
		}
	}

	rankNumbers := make([]int, 0, len(ballotRanks))
	for rank := range ballotRanks {
		rankNumbers = append(rankNumbers, rank)
	}

	sort.Slice(rankNumbers, func(i, j int) bool {
		return rankNumbers[i] < rankNumbers[j]
	})

	ranks = make([][]choiceIndex, 0, len(rankNumbers))
	for _, rankNumber := range rankNumbers {
		ranks = append(ranks, ballotRanks[rankNumber])
	}

	if hasUnrankedChoices {
		unranked := make([]choiceIndex, 0, choicesLen-ballotLen)
		for i := uint(0); int(i) < choicesLen; i++ {
			if !rankedChoices.isSet(i) {
				unranked = append(unranked, choiceIndex(i))
			}
		}
		if len(unranked) > 0 {
			ranks = append(ranks, unranked)
		}
	}

	return ranks, choicesLen, nil
}

func calculatePairwiseStrengths[C comparable](choices []C, preferences []int) []strength {
	choicesCount := len(choices)

	strengths := make([]strength, choicesCount*choicesCount)

	for i := 0; i < choicesCount; i++ {
		for j := 0; j < choicesCount; j++ {
			if i != j {
				c := preferences[i*choicesCount+j]
				if c > preferences[j*choicesCount+i] {
					strengths[i*choicesCount+j] = strength(c)
				}
			}
		}
	}

	for i := 0; i < choicesCount; i++ {
		for j := 0; j < choicesCount; j++ {
			if i != j {
				for k := 0; k < choicesCount; k++ {
					if (i != k) && (j != k) {
						jk := j*choicesCount + k
						strengths[jk] = max(
							strengths[jk],
							min(
								strengths[j*choicesCount+i],
								strengths[i*choicesCount+k],
							),
						)
					}
				}
			}
		}
	}

	return strengths
}

func calculateResults[C comparable](choices []C, strengths []strength) (results []Result[C], tie bool) {
	choicesCount := len(choices)
	results = make([]Result[C], 0, choicesCount)

	for i := 0; i < choicesCount; i++ {
		var count int

		for j := 0; j < choicesCount; j++ {
			if i != j {
				if strengths[i*choicesCount+j] > strengths[j*choicesCount+i] {
					count++
				}
			}
		}
		results = append(results, Result[C]{Choice: choices[i], Index: i, Wins: count})

	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Wins == results[j].Wins {
			return results[i].Index < results[j].Index
		}
		return results[i].Wins > results[j].Wins
	})

	if len(results) >= 2 {
		tie = results[0].Wins == results[1].Wins
	}

	return results, tie
}

type strength int

func min(a, b strength) strength {
	if a < b {
		return a
	}
	return b
}

func max(a, b strength) strength {
	if a > b {
		return a
	}
	return b
}
