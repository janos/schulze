// Copyright (c) 2021, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schulze

import "sort"

// Voting holds voting state in memory for a list of choices and provides
// methods to vote, to export current voting state and to calculate the winner
// using the Schulze method.
type Voting struct {
	choices []string
	matrix  [][]voteCount
}

// NewVoting initializes a new voting with provided choices.
func NewVoting(choices ...string) *Voting {
	return &Voting{
		choices: choices,
		matrix:  makeVoteCountMatrix(len(choices)),
	}
}

// Ballot represents a single vote with ranked choices. Lowest number represents
// the highest rank. Not all choices have to be ranked and multiple choices can
// have the same rank. Ranks do not have to be in consecutive order.
type Ballot map[string]int

func (e *Voting) Vote(b Ballot) error {
	ranks, err := ballotRanks(b, e.choices)
	if err != nil {
		return err
	}

	for rank, choices1 := range ranks {
		rest := ranks[rank+1:]
		for _, i := range choices1 {
			for _, choices1 := range rest {
				for _, j := range choices1 {
					e.matrix[i][j]++
				}
			}
		}
	}

	return nil
}

// VoteMatrix returns the state of the voting in a form of VoteMatrix with
// pairwise number of votes.
func (e *Voting) VoteMatrix() VoteMatrix {
	l := len(e.matrix)
	matrix := make(VoteMatrix, l)

	for i := 0; i < l; i++ {
		for j := 0; j < l; j++ {
			if _, ok := matrix[e.choices[i]]; !ok {
				matrix[e.choices[i]] = make(map[string]int, l)
			}
			matrix[e.choices[i]][e.choices[j]] = int(e.matrix[i][j])
		}
	}

	return matrix
}

// Compute calculates a sorted list of choices with the total number of wins for
// each of them. If there are multiple winners, tie boolean parameter is true.
func (e *Voting) Compute() (scores []Score, tie bool) {
	return compute(e.matrix, e.choices)
}

func ballotRanks(b Ballot, choices []string) ([][]choiceIndex, error) {
	ballotRanks := make(map[int][]choiceIndex)
	rankedChoices := make(map[choiceIndex]struct{})

	for o, rank := range b {
		index := getChoiceIndex(o, choices)
		if index < 0 {
			return nil, &UnknownChoiceError{o}
		}
		ballotRanks[rank] = append(ballotRanks[rank], index)
		rankedChoices[index] = struct{}{}
	}

	rankNumbers := make([]int, 0, len(ballotRanks))
	for rank := range ballotRanks {
		rankNumbers = append(rankNumbers, rank)
	}

	sort.Slice(rankNumbers, func(i, j int) bool {
		return rankNumbers[i] < rankNumbers[j]
	})

	ranks := make([][]choiceIndex, 0)
	for _, rankNumber := range rankNumbers {
		ranks = append(ranks, ballotRanks[rankNumber])
	}

	unranked := make([]choiceIndex, 0)
	for i, l := choiceIndex(0), len(choices); int(i) < l; i++ {
		if _, ok := rankedChoices[i]; !ok {
			unranked = append(unranked, i)
		}
	}
	if len(unranked) > 0 {
		ranks = append(ranks, unranked)
	}

	return ranks, nil
}

type choiceIndex int

func getChoiceIndex(choice string, choices []string) choiceIndex {
	for i, o := range choices {
		if o == choice {
			return choiceIndex(i)
		}
	}
	return -1
}
