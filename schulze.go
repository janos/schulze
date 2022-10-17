// Copyright (c) 2021, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package schulze implements the Schulze method for single winner voting.
package schulze

import (
	"fmt"
	"sort"
)

type VoteCount int

type Matrix [][]VoteCount

// Score represents a total number of wins for a single choice.
type Score[C comparable] struct {
	Choice C
	Index  int
	Wins   int
}

// Voting holds number of votes for every pair of choices.
type Voting[C comparable] struct {
	choices      []C
	choicesIndex map[C]int
	choicesCount int
	matrix       []VoteCount
}

// NewVoting initializes a new voting matrix for a fixed number of choices.
func NewVoting[C comparable](choices ...C) (Voting[C], error) {
	choicesCount := len(choices)
	choicesIndex := make(map[C]int, choicesCount)
	for i, c := range choices {
		for _, cc := range choices[i+1:] {
			if cc == c {
				var v Voting[C]
				return v, fmt.Errorf("duplicate choice: %v", c)
			}
		}
		choicesIndex[c] = i
	}
	return Voting[C]{
		choices:      choices,
		choicesIndex: choicesIndex,
		choicesCount: choicesCount,
		matrix:       make([]VoteCount, choicesCount*choicesCount),
	}, nil
}

func (v Voting[C]) Export() ([]C, Matrix) {
	matrix := make(Matrix, 0, v.choicesCount)
	for i := 0; i < v.choicesCount; i++ {
		row := make([]VoteCount, v.choicesCount)
		copy(row, v.matrix[i:(i+1)*v.choicesCount])
		matrix = append(matrix, row)
	}
	choices := make([]C, len(v.choices))
	copy(choices, v.choices)
	return choices, matrix
}

func (v Voting[C]) Import(matrix Matrix) error {
	choicesCount := v.choicesCount

	if l := len(matrix); l != choicesCount {
		return fmt.Errorf("incorrect matrix length %v", l)
	}
	for i := 0; i < choicesCount; i++ {
		if l := len(matrix[i]); l != choicesCount {
			return fmt.Errorf("incorrect length %v of row %v", l, i)
		}
	}
	for i := 0; i < choicesCount; i++ {
		n := copy(v.matrix[i:(i+1)*choicesCount], matrix[i])
		if n != choicesCount {
			return fmt.Errorf("row %v short read %v", i, n)
		}
	}
	return nil
}

// Ballot represents a single vote with ranked choices. Lowest number represents
// the highest rank. Not all choices have to be ranked and multiple choices can
// have the same rank. Ranks do not have to be in consecutive order.
type Ballot[C comparable] map[C]int

func (v Voting[C]) Vote(b Ballot[C]) error {
	ranks, err := v.ballotRanks(b)
	if err != nil {
		return err
	}

	for rank, choices1 := range ranks {
		rest := ranks[rank+1:]
		for _, i := range choices1 {
			for _, choices1 := range rest {
				for _, j := range choices1 {
					v.matrix[int(i)*v.choicesCount+int(j)]++
				}
			}
		}
	}

	return nil
}

// Results calculates a sorted list of choices with the total number of wins for
// each of them. If there are multiple winners, tie boolean parameter is true.
func (v Voting[C]) Results() (results []Score[C], tie bool) {
	strengths := v.calculatePairwiseStrengths()
	return v.calculateScores(strengths)
}

type choiceIndex int

func (v Voting[C]) ballotRanks(b Ballot[C]) ([][]choiceIndex, error) {
	ballotRanks := make(map[int][]choiceIndex)
	rankedChoices := make(map[choiceIndex]struct{})

	choicesCount := v.choicesCount

	for choice, rank := range b {
		index, ok := v.choicesIndex[choice]
		if !ok {
			return nil, &UnknownChoiceError[C]{Choice: choice}
		}
		ballotRanks[rank] = append(ballotRanks[rank], choiceIndex(index))
		rankedChoices[choiceIndex(index)] = struct{}{}
	}

	rankNumbers := make([]int, 0, len(ballotRanks))
	for rank := range ballotRanks {
		rankNumbers = append(rankNumbers, rank)
	}

	sort.Slice(rankNumbers, func(i, j int) bool {
		return rankNumbers[i] < rankNumbers[j]
	})

	ranks := make([][]choiceIndex, 0, len(rankNumbers))
	for _, rankNumber := range rankNumbers {
		ranks = append(ranks, ballotRanks[rankNumber])
	}

	unranked := make([]choiceIndex, 0, choicesCount-len(rankedChoices))
	for i := choiceIndex(0); int(i) < choicesCount; i++ {
		if _, ok := rankedChoices[i]; !ok {
			unranked = append(unranked, i)
		}
	}
	if len(unranked) > 0 {
		ranks = append(ranks, unranked)
	}

	return ranks, nil
}

func (v Voting[C]) calculatePairwiseStrengths() []strength {
	choicesCount := v.choicesCount

	strengths := make([]strength, choicesCount*choicesCount)

	for i := 0; i < choicesCount; i++ {
		for j := 0; j < choicesCount; j++ {
			if i != j {
				c := v.matrix[i*choicesCount+j]
				if c > v.matrix[j*choicesCount+i] {
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

func (v *Voting[C]) calculateScores(strengths []strength) (scores []Score[C], tie bool) {
	wins := make(map[int][]int)

	for i := 0; i < v.choicesCount; i++ {
		var count int

		for j := 0; j < v.choicesCount; j++ {
			if i != j {
				if strengths[i*v.choicesCount+j] > strengths[j*v.choicesCount+i] {
					count++
				}
			}
		}

		wins[count] = append(wins[count], i)
	}

	scores = make([]Score[C], 0, len(wins))

	for count, choicesIndex := range wins {
		for _, index := range choicesIndex {
			scores = append(scores, Score[C]{Choice: v.choices[index], Index: index, Wins: count})
		}
	}

	sort.Slice(scores, func(i, j int) bool {
		if scores[i].Wins == scores[j].Wins {
			return scores[i].Index < scores[j].Index
		}
		return scores[i].Wins > scores[j].Wins
	})

	if len(scores) >= 2 {
		tie = scores[0].Wins == scores[1].Wins
	}

	return scores, tie
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
