// Copyright (c) 2021, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package schulze implements the Schulze method for single winner voting.
package schulze

import (
	"sort"
)

// Score represents a total number of wins for a single choice.
type Score[C comparable] struct {
	Choice C
	Wins   int
}

// VoteMatrix holds number of votes for every pair of choices.
type VoteMatrix[C comparable] map[C]map[C]int

// Compute calculates a sorted list of choices with the total number of wins for
// each of them. If there are multiple winners, tie boolean parameter is true.
func Compute[C comparable](v VoteMatrix[C]) (scores []Score[C], tie bool) {
	choicesMap := make(map[C]struct{})
	for c1, row := range v {
		for c2 := range row {
			choicesMap[c1] = struct{}{}
			choicesMap[c2] = struct{}{}
		}
	}
	size := len(choicesMap)

	choices := make([]C, 0, size)
	for c := range choicesMap {
		choices = append(choices, c)
	}

	choiceIndexes := make(map[C]int)
	for i, c := range choices {
		choiceIndexes[c] = i
	}

	matrix := makeVoteCountMatrix(size)
	for c1, row := range v {
		for c2, count := range row {
			matrix[choiceIndexes[c1]][choiceIndexes[c2]] = voteCount(count)
		}
	}
	return compute(matrix, choices)
}

func compute[C comparable](matrix [][]voteCount, choices []C) (scores []Score[C], tie bool) {
	strengths := calculatePairwiseStrengths(matrix)
	return calculateScores(strengths, choices)
}

func calculatePairwiseStrengths(m [][]voteCount) [][]strength {
	size := len(m)
	strengths := makeStrenghtMatrix(size)

	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			if i != j {
				c := m[i][j]
				if c > m[j][i] {
					strengths[i][j] = strength(c)
				}
			}
		}
	}

	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			if i != j {
				for k := 0; k < size; k++ {
					if (i != k) && (j != k) {
						strengths[j][k] = max(
							strengths[j][k],
							min(
								strengths[j][i],
								strengths[i][k],
							),
						)
					}
				}
			}
		}
	}

	return strengths
}

func calculateScores[C comparable](strengths [][]strength, choices []C) (scores []Score[C], tie bool) {
	size := len(strengths)
	wins := make(map[int][]int)

	for i := 0; i < size; i++ {
		var count int

		for j := 0; j < size; j++ {
			if i != j {
				if strengths[i][j] > strengths[j][i] {
					count++
				}
			}
		}

		wins[count] = append(wins[count], i)
	}

	scores = make([]Score[C], 0, len(wins))

	for count, choicesIndex := range wins {
		for _, index := range choicesIndex {
			scores = append(scores, Score[C]{Choice: choices[index], Wins: count})
		}
	}

	sort.SliceStable(scores, func(i, j int) bool {
		return scores[i].Wins > scores[j].Wins
	})

	if len(scores) >= 2 {
		tie = scores[0].Wins == scores[1].Wins
	}

	return scores, tie
}

type voteCount int

type strength int

func makeVoteCountMatrix(size int) [][]voteCount {
	matrix := make([][]voteCount, size)
	for i := 0; i < size; i++ {
		matrix[i] = make([]voteCount, size)
	}
	return matrix
}

func makeStrenghtMatrix(size int) [][]strength {
	matrix := make([][]strength, size)
	for i := 0; i < size; i++ {
		matrix[i] = make([]strength, size)
	}
	return matrix
}

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
