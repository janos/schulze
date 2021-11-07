// Copyright (c) 2021, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schulze_test

import (
	"reflect"
	"testing"

	"resenje.org/schulze"
)

func TestVoting(t *testing.T) {
	for _, tc := range []struct {
		name    string
		choices []string
		ballots []schulze.Ballot
		result  []schulze.Score
		tie     bool
	}{
		{
			name:   "empty",
			result: []schulze.Score{},
		},
		{
			name:    "single option no votes",
			choices: []string{"A"},
			result: []schulze.Score{
				{Choice: "A", Wins: 0},
			},
		},
		{
			name:    "single option one vote",
			choices: []string{"A"},
			ballots: []schulze.Ballot{
				{"A": 1},
			},
			result: []schulze.Score{
				{Choice: "A", Wins: 0},
			},
		},
		{
			name:    "two options one vote",
			choices: []string{"A", "B"},
			ballots: []schulze.Ballot{
				{"A": 1},
			},
			result: []schulze.Score{
				{Choice: "A", Wins: 1},
				{Choice: "B", Wins: 0},
			},
		},
		{
			name:    "two options two votes",
			choices: []string{"A", "B"},
			ballots: []schulze.Ballot{
				{"A": 1},
				{"A": 1, "B": 2},
			},
			result: []schulze.Score{
				{Choice: "A", Wins: 1},
				{Choice: "B", Wins: 0},
			},
		},
		{
			name:    "three options three votes",
			choices: []string{"A", "B", "C"},
			ballots: []schulze.Ballot{
				{"A": 1},
				{"A": 1, "B": 2},
				{"A": 1, "B": 2, "C": 3},
			},
			result: []schulze.Score{
				{Choice: "A", Wins: 2},
				{Choice: "B", Wins: 1},
				{Choice: "C", Wins: 0},
			},
		},
		{
			name:    "tie",
			choices: []string{"A", "B", "C"},
			ballots: []schulze.Ballot{
				{"A": 1},
				{"B": 1},
			},
			result: []schulze.Score{
				{Choice: "A", Wins: 1},
				{Choice: "B", Wins: 1},
				{Choice: "C", Wins: 0},
			},
			tie: true,
		},
		{
			name:    "complex",
			choices: []string{"A", "B", "C", "D"},
			ballots: []schulze.Ballot{
				{"A": 1, "B": 1},
				{"B": 1, "C": 1, "A": 2},
				{"A": 1, "B": 2, "C": 2},
				{"A": 1, "B": 200, "C": 10},
			},
			result: []schulze.Score{
				{Choice: "A", Wins: 3},
				{Choice: "B", Wins: 1},
				{Choice: "C", Wins: 1},
				{Choice: "D", Wins: 0},
			},
		},
		{
			name:    "example from wiki page",
			choices: []string{"A", "B", "C", "D", "E"},
			ballots: []schulze.Ballot{
				{"A": 1, "C": 2, "B": 3, "E": 4, "D": 5},
				{"A": 1, "C": 2, "B": 3, "E": 4, "D": 5},
				{"A": 1, "C": 2, "B": 3, "E": 4, "D": 5},
				{"A": 1, "C": 2, "B": 3, "E": 4, "D": 5},
				{"A": 1, "C": 2, "B": 3, "E": 4, "D": 5},

				{"A": 1, "D": 2, "E": 3, "C": 4, "B": 5},
				{"A": 1, "D": 2, "E": 3, "C": 4, "B": 5},
				{"A": 1, "D": 2, "E": 3, "C": 4, "B": 5},
				{"A": 1, "D": 2, "E": 3, "C": 4, "B": 5},
				{"A": 1, "D": 2, "E": 3, "C": 4, "B": 5},

				{"B": 1, "E": 2, "D": 3, "A": 4, "C": 5},
				{"B": 1, "E": 2, "D": 3, "A": 4, "C": 5},
				{"B": 1, "E": 2, "D": 3, "A": 4, "C": 5},
				{"B": 1, "E": 2, "D": 3, "A": 4, "C": 5},
				{"B": 1, "E": 2, "D": 3, "A": 4, "C": 5},
				{"B": 1, "E": 2, "D": 3, "A": 4, "C": 5},
				{"B": 1, "E": 2, "D": 3, "A": 4, "C": 5},
				{"B": 1, "E": 2, "D": 3, "A": 4, "C": 5},

				{"C": 1, "A": 2, "B": 3, "E": 4, "D": 5},
				{"C": 1, "A": 2, "B": 3, "E": 4, "D": 5},
				{"C": 1, "A": 2, "B": 3, "E": 4, "D": 5},

				{"C": 1, "A": 2, "E": 3, "B": 4, "D": 5},
				{"C": 1, "A": 2, "E": 3, "B": 4, "D": 5},
				{"C": 1, "A": 2, "E": 3, "B": 4, "D": 5},
				{"C": 1, "A": 2, "E": 3, "B": 4, "D": 5},
				{"C": 1, "A": 2, "E": 3, "B": 4, "D": 5},
				{"C": 1, "A": 2, "E": 3, "B": 4, "D": 5},
				{"C": 1, "A": 2, "E": 3, "B": 4, "D": 5},

				{"C": 1, "B": 2, "A": 3, "D": 4, "E": 5},
				{"C": 1, "B": 2, "A": 3, "D": 4, "E": 5},

				{"D": 1, "C": 2, "E": 3, "B": 4, "A": 5},
				{"D": 1, "C": 2, "E": 3, "B": 4, "A": 5},
				{"D": 1, "C": 2, "E": 3, "B": 4, "A": 5},
				{"D": 1, "C": 2, "E": 3, "B": 4, "A": 5},
				{"D": 1, "C": 2, "E": 3, "B": 4, "A": 5},
				{"D": 1, "C": 2, "E": 3, "B": 4, "A": 5},
				{"D": 1, "C": 2, "E": 3, "B": 4, "A": 5},

				{"E": 1, "B": 2, "A": 3, "D": 4, "C": 5},
				{"E": 1, "B": 2, "A": 3, "D": 4, "C": 5},
				{"E": 1, "B": 2, "A": 3, "D": 4, "C": 5},
				{"E": 1, "B": 2, "A": 3, "D": 4, "C": 5},
				{"E": 1, "B": 2, "A": 3, "D": 4, "C": 5},
				{"E": 1, "B": 2, "A": 3, "D": 4, "C": 5},
				{"E": 1, "B": 2, "A": 3, "D": 4, "C": 5},
				{"E": 1, "B": 2, "A": 3, "D": 4, "C": 5},
			},
			result: []schulze.Score{
				{Choice: "E", Wins: 4},
				{Choice: "A", Wins: 3},
				{Choice: "C", Wins: 2},
				{Choice: "B", Wins: 1},
				{Choice: "D", Wins: 0},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := schulze.NewVoting(tc.choices...)

			for _, b := range tc.ballots {
				if err := e.Vote(b); err != nil {
					t.Fatal(err)
				}
			}

			t.Run("direct", func(t *testing.T) {
				result, tie := e.Compute()
				if tie != tc.tie {
					t.Errorf("got tie %v, want %v", tie, tc.tie)
				}
				if !reflect.DeepEqual(result, tc.result) {
					t.Errorf("got result %+v, want %+v", result, tc.result)
				}
			})

			t.Run("indirect", func(t *testing.T) {
				result, tie := schulze.Compute(e.VoteMatrix())
				if tie != tc.tie {
					t.Errorf("got tie %v, want %v", tie, tc.tie)
				}
				if !reflect.DeepEqual(result, tc.result) {
					t.Errorf("got result %+v, want %+v", result, tc.result)
				}
			})
		})
	}
}
