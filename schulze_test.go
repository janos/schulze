// Copyright (c) 2021, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schulze_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"reflect"
	"slices"
	"strconv"
	"testing"
	"time"
	"unsafe"

	"resenje.org/schulze"
)

func TestVoting(t *testing.T) {
	type ballot[C comparable] struct {
		vote   schulze.Ballot[C]
		unvote schulze.Record[C]
	}
	for _, tc := range []struct {
		name    string
		choices []string
		ballots []ballot[string]
		result  []schulze.Result[string]
		tie     bool
		duels   []schulze.Duel[string]
	}{
		{
			name:   "empty",
			result: []schulze.Result[string]{},
		},
		{
			name:    "single option no votes",
			choices: []string{"A"},
			result: []schulze.Result[string]{
				{Choice: "A", Index: 0, Wins: 0, Strength: 0, Advantage: 0},
			},
		},
		{
			name:    "single option one vote",
			choices: []string{"A"},
			ballots: []ballot[string]{
				{vote: schulze.Ballot[string]{"A": 1}},
			},
			result: []schulze.Result[string]{
				{Choice: "A", Index: 0, Wins: 0, Strength: 0, Advantage: 0},
			},
		},
		{
			name:    "two options one vote",
			choices: []string{"A", "B"},
			ballots: []ballot[string]{
				{vote: schulze.Ballot[string]{"A": 1}},
			},
			result: []schulze.Result[string]{
				{Choice: "A", Index: 0, Wins: 1, Strength: 1, Advantage: 1},
				{Choice: "B", Index: 1, Wins: 0, Strength: 0, Advantage: 0},
			},
		},
		{
			name:    "two options two votes",
			choices: []string{"A", "B"},
			ballots: []ballot[string]{
				{vote: schulze.Ballot[string]{"A": 1}},
				{vote: schulze.Ballot[string]{"A": 1, "B": 2}},
			},
			result: []schulze.Result[string]{
				{Choice: "A", Index: 0, Wins: 1, Strength: 2, Advantage: 2},
				{Choice: "B", Index: 1, Wins: 0, Strength: 0, Advantage: 0},
			},
		},
		{
			name:    "three options three votes",
			choices: []string{"A", "B", "C"},
			ballots: []ballot[string]{
				{vote: schulze.Ballot[string]{"A": 1}},
				{vote: schulze.Ballot[string]{"A": 1, "B": 2}},
				{vote: schulze.Ballot[string]{"A": 1, "B": 2, "C": 3}},
			},
			result: []schulze.Result[string]{
				{Choice: "A", Index: 0, Wins: 2, Strength: 6, Advantage: 6},
				{Choice: "B", Index: 1, Wins: 1, Strength: 2, Advantage: 2},
				{Choice: "C", Index: 2, Wins: 0, Strength: 0, Advantage: 0},
			},
		},
		{
			name:    "tie",
			choices: []string{"A", "B", "C"},
			ballots: []ballot[string]{
				{vote: schulze.Ballot[string]{"A": 1}},
				{vote: schulze.Ballot[string]{"B": 1}},
			},
			result: []schulze.Result[string]{
				{Choice: "A", Index: 0, Wins: 1, Strength: 1, Advantage: 1},
				{Choice: "B", Index: 1, Wins: 1, Strength: 1, Advantage: 1},
				{Choice: "C", Index: 2, Wins: 0, Strength: 0, Advantage: 0},
			},
			tie: true,
		},
		{
			name:    "complex",
			choices: []string{"A", "B", "C", "D", "E"},
			ballots: []ballot[string]{
				{vote: schulze.Ballot[string]{"A": 1, "B": 1}},
				{vote: schulze.Ballot[string]{"B": 1, "C": 1, "A": 2}},
				{vote: schulze.Ballot[string]{"A": 1, "B": 2, "C": 2}},
				{vote: schulze.Ballot[string]{"A": 1, "B": 200, "C": 10}},
			},
			result: []schulze.Result[string]{
				{Choice: "A", Index: 0, Wins: 4, Strength: 13, Advantage: 13},
				{Choice: "B", Index: 1, Wins: 2, Strength: 8, Advantage: 8},
				{Choice: "C", Index: 2, Wins: 2, Strength: 6, Advantage: 6},
				{Choice: "D", Index: 3, Wins: 0, Strength: 0, Advantage: 0},
				{Choice: "E", Index: 4, Wins: 0, Strength: 0, Advantage: 0},
			},
		},
		{
			name:    "duplicate choice", // only the first of the duplicate choices should receive votes
			choices: []string{"A", "B", "C", "C", "C"},
			ballots: []ballot[string]{
				{vote: schulze.Ballot[string]{"A": 1, "B": 1}},
				{vote: schulze.Ballot[string]{"B": 1, "C": 1, "A": 2}},
				{vote: schulze.Ballot[string]{"A": 1, "B": 2, "C": 2}},
				{vote: schulze.Ballot[string]{"A": 1, "B": 200, "C": 10}},
			},
			result: []schulze.Result[string]{
				{Choice: "A", Index: 0, Wins: 4, Strength: 13, Advantage: 13},
				{Choice: "B", Index: 1, Wins: 2, Strength: 8, Advantage: 8},
				{Choice: "C", Index: 2, Wins: 2, Strength: 6, Advantage: 6},
				{Choice: "C", Index: 3, Wins: 0, Strength: 0, Advantage: 0},
				{Choice: "C", Index: 4, Wins: 0, Strength: 0, Advantage: 0},
			},
		},
		{
			name:    "example from wiki page",
			choices: []string{"A", "B", "C", "D", "E"},
			ballots: []ballot[string]{
				{vote: schulze.Ballot[string]{"A": 1, "C": 2, "B": 3, "E": 4, "D": 5}},
				{vote: schulze.Ballot[string]{"A": 1, "C": 2, "B": 3, "E": 4, "D": 5}},
				{vote: schulze.Ballot[string]{"A": 1, "C": 2, "B": 3, "E": 4, "D": 5}},
				{vote: schulze.Ballot[string]{"A": 1, "C": 2, "B": 3, "E": 4, "D": 5}},
				{vote: schulze.Ballot[string]{"A": 1, "C": 2, "B": 3, "E": 4, "D": 5}},

				{vote: schulze.Ballot[string]{"A": 1, "D": 2, "E": 3, "C": 4, "B": 5}},
				{vote: schulze.Ballot[string]{"A": 1, "D": 2, "E": 3, "C": 4, "B": 5}},
				{vote: schulze.Ballot[string]{"A": 1, "D": 2, "E": 3, "C": 4, "B": 5}},
				{vote: schulze.Ballot[string]{"A": 1, "D": 2, "E": 3, "C": 4, "B": 5}},
				{vote: schulze.Ballot[string]{"A": 1, "D": 2, "E": 3, "C": 4, "B": 5}},

				{vote: schulze.Ballot[string]{"B": 1, "E": 2, "D": 3, "A": 4, "C": 5}},
				{vote: schulze.Ballot[string]{"B": 1, "E": 2, "D": 3, "A": 4, "C": 5}},
				{vote: schulze.Ballot[string]{"B": 1, "E": 2, "D": 3, "A": 4, "C": 5}},
				{vote: schulze.Ballot[string]{"B": 1, "E": 2, "D": 3, "A": 4, "C": 5}},
				{vote: schulze.Ballot[string]{"B": 1, "E": 2, "D": 3, "A": 4, "C": 5}},
				{vote: schulze.Ballot[string]{"B": 1, "E": 2, "D": 3, "A": 4, "C": 5}},
				{vote: schulze.Ballot[string]{"B": 1, "E": 2, "D": 3, "A": 4, "C": 5}},
				{vote: schulze.Ballot[string]{"B": 1, "E": 2, "D": 3, "A": 4, "C": 5}},

				{vote: schulze.Ballot[string]{"C": 1, "A": 2, "B": 3, "E": 4, "D": 5}},
				{vote: schulze.Ballot[string]{"C": 1, "A": 2, "B": 3, "E": 4, "D": 5}},
				{vote: schulze.Ballot[string]{"C": 1, "A": 2, "B": 3, "E": 4, "D": 5}},

				{vote: schulze.Ballot[string]{"C": 1, "A": 2, "E": 3, "B": 4, "D": 5}},
				{vote: schulze.Ballot[string]{"C": 1, "A": 2, "E": 3, "B": 4, "D": 5}},
				{vote: schulze.Ballot[string]{"C": 1, "A": 2, "E": 3, "B": 4, "D": 5}},
				{vote: schulze.Ballot[string]{"C": 1, "A": 2, "E": 3, "B": 4, "D": 5}},
				{vote: schulze.Ballot[string]{"C": 1, "A": 2, "E": 3, "B": 4, "D": 5}},
				{vote: schulze.Ballot[string]{"C": 1, "A": 2, "E": 3, "B": 4, "D": 5}},
				{vote: schulze.Ballot[string]{"C": 1, "A": 2, "E": 3, "B": 4, "D": 5}},

				{vote: schulze.Ballot[string]{"C": 1, "B": 2, "A": 3, "D": 4, "E": 5}},
				{vote: schulze.Ballot[string]{"C": 1, "B": 2, "A": 3, "D": 4, "E": 5}},

				{vote: schulze.Ballot[string]{"D": 1, "C": 2, "E": 3, "B": 4, "A": 5}},
				{vote: schulze.Ballot[string]{"D": 1, "C": 2, "E": 3, "B": 4, "A": 5}},
				{vote: schulze.Ballot[string]{"D": 1, "C": 2, "E": 3, "B": 4, "A": 5}},
				{vote: schulze.Ballot[string]{"D": 1, "C": 2, "E": 3, "B": 4, "A": 5}},
				{vote: schulze.Ballot[string]{"D": 1, "C": 2, "E": 3, "B": 4, "A": 5}},
				{vote: schulze.Ballot[string]{"D": 1, "C": 2, "E": 3, "B": 4, "A": 5}},
				{vote: schulze.Ballot[string]{"D": 1, "C": 2, "E": 3, "B": 4, "A": 5}},

				{vote: schulze.Ballot[string]{"E": 1, "B": 2, "A": 3, "D": 4, "C": 5}},
				{vote: schulze.Ballot[string]{"E": 1, "B": 2, "A": 3, "D": 4, "C": 5}},
				{vote: schulze.Ballot[string]{"E": 1, "B": 2, "A": 3, "D": 4, "C": 5}},
				{vote: schulze.Ballot[string]{"E": 1, "B": 2, "A": 3, "D": 4, "C": 5}},
				{vote: schulze.Ballot[string]{"E": 1, "B": 2, "A": 3, "D": 4, "C": 5}},
				{vote: schulze.Ballot[string]{"E": 1, "B": 2, "A": 3, "D": 4, "C": 5}},
				{vote: schulze.Ballot[string]{"E": 1, "B": 2, "A": 3, "D": 4, "C": 5}},
				{vote: schulze.Ballot[string]{"E": 1, "B": 2, "A": 3, "D": 4, "C": 5}},
			},
			result: []schulze.Result[string]{
				{Choice: "E", Index: 4, Wins: 4, Strength: 112, Advantage: 16},
				{Choice: "A", Index: 0, Wins: 3, Strength: 86, Advantage: 11},
				{Choice: "C", Index: 2, Wins: 2, Strength: 58, Advantage: 2},
				{Choice: "B", Index: 1, Wins: 1, Strength: 33, Advantage: 5},
				{Choice: "D", Index: 3, Wins: 0, Strength: 0, Advantage: 0},
			},
			duels: []schulze.Duel[string]{
				{
					Left:  schulze.ChoiceStrength[string]{Choice: "A", Index: 0, Strength: 28},
					Right: schulze.ChoiceStrength[string]{Choice: "B", Index: 1, Strength: 25}},
				{
					Left:  schulze.ChoiceStrength[string]{Choice: "A", Index: 0, Strength: 28},
					Right: schulze.ChoiceStrength[string]{Choice: "C", Index: 2, Strength: 25}},
				{
					Left:  schulze.ChoiceStrength[string]{Choice: "A", Index: 0, Strength: 30},
					Right: schulze.ChoiceStrength[string]{Choice: "D", Index: 3, Strength: 25}},
				{
					Left:  schulze.ChoiceStrength[string]{Choice: "A", Index: 0, Strength: 24},
					Right: schulze.ChoiceStrength[string]{Choice: "E", Index: 4, Strength: 25}},
				{
					Left:  schulze.ChoiceStrength[string]{Choice: "B", Index: 1, Strength: 28},
					Right: schulze.ChoiceStrength[string]{Choice: "C", Index: 2, Strength: 29}},
				{
					Left:  schulze.ChoiceStrength[string]{Choice: "B", Index: 1, Strength: 33},
					Right: schulze.ChoiceStrength[string]{Choice: "D", Index: 3, Strength: 28}},
				{
					Left:  schulze.ChoiceStrength[string]{Choice: "B", Index: 1, Strength: 24},
					Right: schulze.ChoiceStrength[string]{Choice: "E", Index: 4, Strength: 28}},
				{
					Left:  schulze.ChoiceStrength[string]{Choice: "C", Index: 2, Strength: 29},
					Right: schulze.ChoiceStrength[string]{Choice: "D", Index: 3, Strength: 28}},
				{
					Left:  schulze.ChoiceStrength[string]{Choice: "C", Index: 2, Strength: 24},
					Right: schulze.ChoiceStrength[string]{Choice: "E", Index: 4, Strength: 28}},
				{
					Left:  schulze.ChoiceStrength[string]{Choice: "D", Index: 3, Strength: 24},
					Right: schulze.ChoiceStrength[string]{Choice: "E", Index: 4, Strength: 31},
				},
			},
		},
		{
			name:    "unvote single option one vote",
			choices: []string{"A"},
			ballots: []ballot[string]{
				{vote: schulze.Ballot[string]{"A": 1}},
				{unvote: schulze.Record[string]{{"A"}}},
			},
			result: []schulze.Result[string]{
				{Choice: "A", Index: 0, Wins: 0, Strength: 0, Advantage: 0},
			},
		},
		{
			name:    "unvote two options one vote",
			choices: []string{"A", "B"},
			ballots: []ballot[string]{
				{vote: schulze.Ballot[string]{"A": 1}},
				{unvote: schulze.Record[string]{{"A"}, {"B"}}},
			},
			result: []schulze.Result[string]{
				{Choice: "A", Index: 0, Wins: 0, Strength: 0, Advantage: 0},
				{Choice: "B", Index: 1, Wins: 0, Strength: 0, Advantage: 0},
			},
			tie: true,
		},
		{
			name:    "empty unvote two options one vote",
			choices: []string{"A", "B"},
			ballots: []ballot[string]{
				{vote: schulze.Ballot[string]{"A": 1}},
				{unvote: schulze.Record[string]{}},
			},
			result: []schulze.Result[string]{
				{Choice: "A", Index: 0, Wins: 1, Strength: 1, Advantage: 1},
				{Choice: "B", Index: 1, Wins: 0, Strength: 0, Advantage: 0},
			},
		},
		{
			name:    "unvote complex",
			choices: []string{"A", "B", "C", "D", "E"},
			ballots: []ballot[string]{
				{vote: schulze.Ballot[string]{"A": 1, "B": 1}},
				{vote: schulze.Ballot[string]{"B": 1, "C": 1, "A": 2}},
				{vote: schulze.Ballot[string]{"A": 1, "B": 2, "C": 2}},
				{vote: schulze.Ballot[string]{"A": 1, "B": 200, "C": 10}},
				{unvote: schulze.Record[string]{{"A"}, {"B", "C"}, {"D", "E"}}},
			},
			result: []schulze.Result[string]{
				{Choice: "A", Index: 0, Wins: 3, Strength: 8, Advantage: 8},
				{Choice: "B", Index: 1, Wins: 2, Strength: 6, Advantage: 6},
				{Choice: "C", Index: 2, Wins: 2, Strength: 4, Advantage: 4},
				{Choice: "D", Index: 3, Wins: 0, Strength: 0, Advantage: 0},
				{Choice: "E", Index: 4, Wins: 0, Strength: 0, Advantage: 0},
			},
		},
		{
			name:    "multiple unvote complex",
			choices: []string{"A", "B", "C", "D", "E"},
			ballots: []ballot[string]{
				{vote: schulze.Ballot[string]{"A": 1, "B": 1}},
				{vote: schulze.Ballot[string]{"B": 1, "C": 1, "A": 2}},
				{vote: schulze.Ballot[string]{"A": 1, "B": 2, "C": 2}},
				{unvote: schulze.Record[string]{{"A", "B"}, {"C", "D", "E"}}},
				{vote: schulze.Ballot[string]{"A": 1, "B": 200, "C": 10}},
				{unvote: schulze.Record[string]{{"A"}, {"B", "C"}, {"D", "E"}}},
				{unvote: schulze.Record[string]{{"B", "C"}, {"A"}, {"D", "E"}}},
			},
			result: []schulze.Result[string]{
				{Choice: "A", Index: 0, Wins: 4, Strength: 4, Advantage: 4},
				{Choice: "C", Index: 2, Wins: 3, Strength: 3, Advantage: 3},
				{Choice: "B", Index: 1, Wins: 2, Strength: 2, Advantage: 2},
				{Choice: "D", Index: 3, Wins: 0, Strength: 0, Advantage: 0},
				{Choice: "E", Index: 4, Wins: 0, Strength: 0, Advantage: 0},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("functional", func(t *testing.T) {
				preferences := schulze.NewPreferences[int](len(tc.choices))

				for _, b := range tc.ballots {
					if b.unvote != nil {
						if err := schulze.Unvote(preferences, tc.choices, b.unvote); err != nil {
							t.Fatal(err)
						}
					} else {
						if _, err := schulze.Vote(preferences, tc.choices, b.vote); err != nil {
							t.Fatal(err)
						}
					}
				}

				result, duels, tie := schulze.Compute(preferences, tc.choices)
				if tie != tc.tie {
					t.Errorf("got tie %v, want %v", tie, tc.tie)
				}
				if !reflect.DeepEqual(result, tc.result) {
					t.Errorf("got result %+v, want %+v", result, tc.result)
				}
				if tc.duels != nil {
					var got []schulze.Duel[string]
					for d := duels(); d != nil; d = duels() {
						got = append(got, *d)
					}
					if !reflect.DeepEqual(got, tc.duels) {
						t.Errorf("got duels %+v, want %+v", got, tc.duels)
					}
				}
			})
			t.Run("Voting", func(t *testing.T) {
				v := schulze.NewVoting(tc.choices)

				for _, b := range tc.ballots {
					if b.unvote != nil {
						if err := v.Unvote(b.unvote); err != nil {
							t.Fatal(err)
						}
					} else {
						if _, err := v.Vote(b.vote); err != nil {
							t.Fatal(err)
						}
					}
				}

				result, duels, tie := v.Compute()
				if tie != tc.tie {
					t.Errorf("got tie %v, want %v", tie, tc.tie)
				}
				if !reflect.DeepEqual(result, tc.result) {
					t.Errorf("got result %+v, want %+v", result, tc.result)
				}
				if tc.duels != nil {
					var got []schulze.Duel[string]
					for d := duels(); d != nil; d = duels() {
						got = append(got, *d)
					}
					if !reflect.DeepEqual(got, tc.duels) {
						t.Errorf("got duels %+v, want %+v", got, tc.duels)
					}
				}
			})
		})
	}
}

func TestUnvote_afterSetChoices(t *testing.T) {
	t.Run("add with unranked", func(t *testing.T) {
		choices := []string{"A", "B", "C"}
		preferences := schulze.NewPreferences[int](len(choices))

		ballot := schulze.Ballot[string]{"A": 1, "B": 2}
		record, err := schulze.Vote(preferences, choices, ballot)
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("initial\n%v", sprintPreferences(choices, preferences))

		updatedChoices := []string{"A", "D", "B", "C"}

		updatedPreferences := schulze.SetChoices(preferences, choices, updatedChoices)

		t.Logf("updated\n%v", sprintPreferences(updatedChoices, updatedPreferences))

		if err := schulze.Unvote(updatedPreferences, updatedChoices, record); err != nil {
			t.Fatal(err)
		}

		t.Logf("unvoted\n%v\n%v", sprintPreferences(updatedChoices, updatedPreferences), record)

		wantPreferences := make([]int, len(updatedPreferences))

		if !reflect.DeepEqual(updatedPreferences, wantPreferences) {
			t.Errorf("got preferences %v, want %v", updatedPreferences, wantPreferences)
		}
	})

	t.Run("add without unranked", func(t *testing.T) {
		choices := []string{"A", "B", "C"}
		preferences := schulze.NewPreferences[int](len(choices))

		ballot := schulze.Ballot[string]{"A": 1, "B": 2, "C": 1}
		record, err := schulze.Vote(preferences, choices, ballot)
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("initial\n%v", sprintPreferences(choices, preferences))

		updatedChoices := []string{"A", "B", "C", "D"}

		updatedPreferences := schulze.SetChoices(preferences, choices, updatedChoices)

		t.Logf("updated\n%v", sprintPreferences(updatedChoices, updatedPreferences))

		if err := schulze.Unvote(updatedPreferences, updatedChoices, record); err != nil {
			t.Fatal(err)
		}

		t.Logf("unvoted\n%v\n%v", sprintPreferences(updatedChoices, updatedPreferences), record)

		wantPreferences := make([]int, len(updatedPreferences))

		if !reflect.DeepEqual(updatedPreferences, wantPreferences) {
			t.Errorf("got preferences %v, want %v", updatedPreferences, wantPreferences)
		}
	})

	t.Run("remove with unranked", func(t *testing.T) {
		choices := []string{"A", "B", "C"}
		preferences := schulze.NewPreferences[int](len(choices))

		ballot := schulze.Ballot[string]{"A": 1, "B": 2}
		record, err := schulze.Vote(preferences, choices, ballot)
		if err != nil {
			t.Fatal(err)
		}

		updatedChoices := []string{"A", "C"}

		updatedPreferences := schulze.SetChoices(preferences, choices, updatedChoices)

		if err := schulze.Unvote(updatedPreferences, updatedChoices, record); err != nil {
			t.Fatal(err)
		}

		wantPreferences := make([]int, len(updatedPreferences))

		if !reflect.DeepEqual(updatedPreferences, wantPreferences) {
			t.Errorf("got preferences %v, want %v", updatedPreferences, wantPreferences)
		}
	})

	t.Run("remove without unranked", func(t *testing.T) {
		choices := []string{"A", "B", "C"}
		preferences := schulze.NewPreferences[int](len(choices))

		ballot := schulze.Ballot[string]{"A": 1, "B": 2, "C": 1}
		record, err := schulze.Vote(preferences, choices, ballot)
		if err != nil {
			t.Fatal(err)
		}

		updatedChoices := []string{"A", "C"}

		updatedPreferences := schulze.SetChoices(preferences, choices, updatedChoices)

		if err := schulze.Unvote(updatedPreferences, updatedChoices, record); err != nil {
			t.Fatal(err)
		}

		wantPreferences := make([]int, len(updatedPreferences))

		if !reflect.DeepEqual(updatedPreferences, wantPreferences) {
			t.Errorf("got preferences %v, want %v", updatedPreferences, wantPreferences)
		}
	})

	t.Run("rearrange", func(t *testing.T) {
		choices := []string{"A", "B", "C"}
		preferences := schulze.NewPreferences[int](len(choices))

		ballot := schulze.Ballot[string]{"A": 1, "B": 2}
		record, err := schulze.Vote(preferences, choices, ballot)
		if err != nil {
			t.Fatal(err)
		}

		updatedChoices := []string{"B", "A", "C"}

		updatedPreferences := schulze.SetChoices(preferences, choices, updatedChoices)

		if err := schulze.Unvote(updatedPreferences, updatedChoices, record); err != nil {
			t.Fatal(err)
		}

		wantPreferences := make([]int, len(updatedPreferences))

		if !reflect.DeepEqual(updatedPreferences, wantPreferences) {
			t.Errorf("got preferences %v, want %v", updatedPreferences, wantPreferences)
		}
	})

	t.Run("complex with unranked", func(t *testing.T) {
		choices := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"}
		preferences := schulze.NewPreferences[int](len(choices))

		ballot := schulze.Ballot[string]{"A": 1, "B": 2, "C": 2, "D": 3, "E": 3, "F": 3, "G": 3}
		record, err := schulze.Vote(preferences, choices, ballot)
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("initial\n%v", sprintPreferences(choices, preferences))

		updatedChoices := []string{"A", "K", "C", "E", "D", "G", "H", "J"}

		updatedPreferences := schulze.SetChoices(preferences, choices, updatedChoices)

		t.Logf("updated\n%v", sprintPreferences(updatedChoices, updatedPreferences))

		if err := schulze.Unvote(updatedPreferences, updatedChoices, record); err != nil {
			t.Fatal(err)
		}

		t.Logf("unvoted\n%v\n%v", sprintPreferences(updatedChoices, updatedPreferences), record)

		wantPreferences := make([]int, len(updatedPreferences))

		if !reflect.DeepEqual(updatedPreferences, wantPreferences) {
			t.Errorf("got preferences %v, want %v", updatedPreferences, wantPreferences)
		}
	})

	t.Run("complex without unranked", func(t *testing.T) {
		choices := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"}
		preferences := schulze.NewPreferences[int](len(choices))

		ballot := schulze.Ballot[string]{"A": 1, "B": 2, "C": 2, "D": 3, "E": 3, "F": 3, "G": 3, "H": 1, "I": 2, "J": 3}
		record, err := schulze.Vote(preferences, choices, ballot)
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("initial\n%v", sprintPreferences(choices, preferences))

		updatedChoices := []string{"A", "K", "C", "E", "D", "G", "H", "J"}

		updatedPreferences := schulze.SetChoices(preferences, choices, updatedChoices)

		t.Logf("updated\n%v", sprintPreferences(updatedChoices, updatedPreferences))

		if err := schulze.Unvote(updatedPreferences, updatedChoices, record); err != nil {
			t.Fatal(err)
		}

		t.Logf("unvoted\n%v\n%v", sprintPreferences(updatedChoices, updatedPreferences), record)

		wantPreferences := make([]int, len(updatedPreferences))

		if !reflect.DeepEqual(updatedPreferences, wantPreferences) {
			t.Errorf("got preferences %v, want %v", updatedPreferences, wantPreferences)
		}
	})
}

func TestDuel_Outcome(t *testing.T) {
	t.Run("tie", func(t *testing.T) {
		winner, defeated := schulze.Duel[string]{
			Left: schulze.ChoiceStrength[string]{
				Choice:   "A",
				Index:    1,
				Strength: 5,
			},
			Right: schulze.ChoiceStrength[string]{
				Choice:   "B",
				Index:    2,
				Strength: 5,
			},
		}.Outcome()

		if winner != nil {
			t.Error("expected for winner to be nil")
		}
		if defeated != nil {
			t.Error("expected for defeated to be nil")
		}
	})

	t.Run("winner", func(t *testing.T) {
		winner, defeated := schulze.Duel[string]{
			Left: schulze.ChoiceStrength[string]{
				Choice:   "A",
				Index:    1,
				Strength: 6,
			},
			Right: schulze.ChoiceStrength[string]{
				Choice:   "B",
				Index:    2,
				Strength: 5,
			},
		}.Outcome()

		if winner.Choice != "A" {
			t.Errorf("expected for winner to be A, got %v", winner.Choice)
		}
		if defeated.Choice != "B" {
			t.Errorf("expected for defeated to be B, got %v", defeated.Choice)
		}
		if winner.Index != 1 {
			t.Errorf("expected for winner to be 1, got %v", winner.Index)
		}
		if defeated.Index != 2 {
			t.Errorf("expected for defeated to be 2, got %v", defeated.Index)
		}
		if winner.Strength != 6 {
			t.Errorf("expected for winner to be 6, got %v", winner.Strength)
		}
		if defeated.Strength != 5 {
			t.Errorf("expected for defeated to be 5, got %v", defeated.Strength)
		}
	})

	t.Run("defeated", func(t *testing.T) {
		winner, defeated := schulze.Duel[string]{
			Left: schulze.ChoiceStrength[string]{
				Choice:   "A",
				Index:    1,
				Strength: 4,
			},
			Right: schulze.ChoiceStrength[string]{
				Choice:   "B",
				Index:    2,
				Strength: 5,
			},
		}.Outcome()

		if winner.Choice != "B" {
			t.Errorf("expected for winner to be B, got %v", winner.Choice)
		}
		if defeated.Choice != "A" {
			t.Errorf("expected for defeated to be A, got %v", defeated.Choice)
		}
		if winner.Index != 2 {
			t.Errorf("expected for winner to be 2, got %v", winner.Index)
		}
		if defeated.Index != 1 {
			t.Errorf("expected for defeated to be 1, got %v", defeated.Index)
		}
		if winner.Strength != 5 {
			t.Errorf("expected for winner to be 5, got %v", winner.Strength)
		}
		if defeated.Strength != 4 {
			t.Errorf("expected for defeated to be 4, got %v", defeated.Strength)
		}
	})
}

func TestSetChoices(t *testing.T) {
	validatePreferences := func(t *testing.T, updatedPreferences, validationPreferences, currentPreferences []int, currentChoices, updatedChoices []string) {
		t.Helper()

		if fmt.Sprint(updatedPreferences) != fmt.Sprint(validationPreferences) {
			t.Errorf("\ngot preferences\n%v\nwant\n%v\nbased on\n%v\n", sprintPreferences(updatedChoices, updatedPreferences), sprintPreferences(updatedChoices, validationPreferences), sprintPreferences(currentChoices, currentPreferences))
		} else {
			t.Logf("\nupdated preferences\n%v\nvalidation preferences\n%v\nbased on\n%v\n", sprintPreferences(updatedChoices, updatedPreferences), sprintPreferences(updatedChoices, validationPreferences), sprintPreferences(currentChoices, currentPreferences))
		}
	}

	for _, tc := range []struct {
		name    string
		ballots []schulze.Ballot[string]
		current []string
		updated []string
	}{
		{
			name:    "no votes, no choices",
			current: []string{},
			updated: []string{},
		},
		{
			name:    "no votes, no change",
			current: []string{"A", "B", "C", "D", "E"},
			updated: []string{"A", "B", "C", "D", "E"},
		},
		{
			name: "single vote, single choice",
			ballots: []schulze.Ballot[string]{
				{"A": 1},
			},
			current: []string{"A"},
			updated: []string{"A"},
		},
		{
			name: "single vote, no change",
			ballots: []schulze.Ballot[string]{
				{"B": 1},
			},
			current: []string{"A", "B", "C", "D", "E"},
			updated: []string{"A", "B", "C", "D", "E"},
		},
		{
			name: "multiple votes, no change",
			ballots: []schulze.Ballot[string]{
				{"B": 1},
				{"A": 1, "C": 2, "D": 2},
				{"B": 1, "D": 2, "E": 3},
			},
			current: []string{"A", "B", "C", "D", "E"},
			updated: []string{"A", "B", "C", "D", "E"},
		},
		{
			name: "single vote, swap two choices",
			ballots: []schulze.Ballot[string]{
				{"C": 1},
			},
			current: []string{"A", "B", "C", "D", "E"},
			updated: []string{"A", "B", "D", "C", "E"},
		},
		{
			name: "multiple votes, swap two choices",
			ballots: []schulze.Ballot[string]{
				{"A": 1},
				{"A": 1},
				{"B": 1},
				{"B": 1},
				{"B": 1, "A": 2},
				{"C": 1},
				{"C": 1},
				{"C": 1, "B": 2},
				{"C": 2, "B": 2, "A": 3},
				{"D": 1},
				{"D": 1},
				{"D": 1, "C": 2},
				{"D": 2, "C": 2, "B": 3},
				{"D": 1, "C": 3, "B": 3, "A": 4},
				{"E": 1},
				{"E": 1},
				{"E": 2, "D": 2},
				{"E": 1, "D": 2},
				{"E": 2, "D": 2, "C": 3},
				{"E": 1, "D": 2, "C": 3, "B": 3},
				{"E": 2, "D": 2, "C": 3, "B": 4, "A": 5},
			},
			current: []string{"A", "B", "C", "D", "E"},
			updated: []string{"A", "B", "D", "C", "E"},
		},
		{
			name: "multiple votes, remove first choice",
			ballots: []schulze.Ballot[string]{
				{"A": 1},
				{"A": 1},
				{"B": 1},
				{"B": 1},
				{"B": 1, "A": 2},
				{"C": 1},
				{"C": 1},
				{"C": 1, "B": 2},
				{"C": 2, "B": 2, "A": 3},
				{"D": 1},
				{"D": 1},
				{"D": 1, "C": 2},
				{"D": 2, "C": 2, "B": 3},
				{"D": 1, "C": 3, "B": 3, "A": 4},
				{"E": 1},
				{"E": 1},
				{"E": 2, "D": 2},
				{"E": 1, "D": 2},
				{"E": 2, "D": 2, "C": 3},
				{"E": 1, "D": 2, "C": 3, "B": 3},
				{"E": 2, "D": 2, "C": 3, "B": 4, "A": 5},
			},
			current: []string{"A", "B", "C", "D", "E"},
			updated: []string{"B", "C", "D", "E"},
		},
		{
			name: "multiple votes, remove last choice",
			ballots: []schulze.Ballot[string]{
				{"A": 1},
				{"A": 1},
				{"B": 1},
				{"B": 1},
				{"B": 1, "A": 2},
				{"C": 1},
				{"C": 1},
				{"C": 1, "B": 2},
				{"C": 2, "B": 2, "A": 3},
				{"D": 1},
				{"D": 1},
				{"D": 1, "C": 2},
				{"D": 2, "C": 2, "B": 3},
				{"D": 1, "C": 3, "B": 3, "A": 4},
				{"E": 1},
				{"E": 1},
				{"E": 2, "D": 2},
				{"E": 1, "D": 2},
				{"E": 2, "D": 2, "C": 3},
				{"E": 1, "D": 2, "C": 3, "B": 3},
				{"E": 2, "D": 2, "C": 3, "B": 4, "A": 5},
			},
			current: []string{"A", "B", "C", "D", "E"},
			updated: []string{"A", "B", "C", "D"},
		},
		{
			name: "multiple votes, remove middle choice",
			ballots: []schulze.Ballot[string]{
				{"A": 1},
				{"A": 1},
				{"B": 1},
				{"B": 1},
				{"B": 1, "A": 2},
				{"C": 1},
				{"C": 1},
				{"C": 1, "B": 2},
				{"C": 2, "B": 2, "A": 3},
				{"D": 1},
				{"D": 1},
				{"D": 1, "C": 2},
				{"D": 2, "C": 2, "B": 3},
				{"D": 1, "C": 3, "B": 3, "A": 4},
				{"E": 1},
				{"E": 1},
				{"E": 2, "D": 2},
				{"E": 1, "D": 2},
				{"E": 2, "D": 2, "C": 3},
				{"E": 1, "D": 2, "C": 3, "B": 3},
				{"E": 2, "D": 2, "C": 3, "B": 4, "A": 5},
			},
			current: []string{"A", "B", "C", "D", "E"},
			updated: []string{"A", "B", "D", "E"},
		},
		{
			name: "multiple votes, remove multiple choices",
			ballots: []schulze.Ballot[string]{
				{"A": 1},
				{"A": 1},
				{"B": 1},
				{"B": 1},
				{"B": 1, "A": 2},
				{"C": 1},
				{"C": 1},
				{"C": 1, "B": 2},
				{"C": 2, "B": 2, "A": 3},
				{"D": 1},
				{"D": 1},
				{"D": 1, "C": 2},
				{"D": 2, "C": 2, "B": 3},
				{"D": 1, "C": 3, "B": 3, "A": 4},
				{"E": 1},
				{"E": 1},
				{"E": 2, "D": 2},
				{"E": 1, "D": 2},
				{"E": 2, "D": 2, "C": 3},
				{"E": 1, "D": 2, "C": 3, "B": 3},
				{"E": 2, "D": 2, "C": 3, "B": 4, "A": 5},
			},
			current: []string{"A", "B", "C", "D", "E"},
			updated: []string{"B", "C"},
		},
		{
			name: "multiple votes, remove and swap multiple choices",
			ballots: []schulze.Ballot[string]{
				{"A": 1},
				{"A": 1},
				{"B": 1},
				{"B": 1},
				{"B": 1, "A": 2},
				{"C": 1},
				{"C": 1},
				{"C": 1, "B": 2},
				{"C": 2, "B": 2, "A": 3},
				{"D": 1},
				{"D": 1},
				{"D": 1, "C": 2},
				{"D": 2, "C": 2, "B": 3},
				{"D": 1, "C": 3, "B": 3, "A": 4},
				{"E": 1},
				{"E": 1},
				{"E": 2, "D": 2},
				{"E": 1, "D": 2},
				{"E": 2, "D": 2, "C": 3},
				{"E": 1, "D": 2, "C": 3, "B": 3},
				{"E": 2, "D": 2, "C": 3, "B": 4, "A": 5},
			},
			current: []string{"A", "B", "C", "D", "E"},
			updated: []string{"B", "D", "C"},
		},
		{
			name: "single vote append choice",
			ballots: []schulze.Ballot[string]{
				{"A": 1},
			},
			current: []string{"A", "B"},
			updated: []string{"A", "B", "C"},
		},
		{
			name: "multiple votes, append choice",
			ballots: []schulze.Ballot[string]{
				{"A": 1},
				{"A": 1},
				{"A": 1},
				{"A": 1},
				{"A": 1},
				{"A": 1},
				{"A": 1},
				{"A": 1},
				{"A": 1},
				{"A": 1},
				{"B": 1, "A": 2},
				{"B": 1},
				{"B": 1, "A": 2},
				{"B": 1, "A": 2},
				{"B": 1, "A": 2},
				{"B": 1, "A": 2},
				{"B": 1, "A": 2},
				{"C": 1},
				{"C": 1},
				{"C": 1},
				{"C": 1, "B": 2},
				{"C": 2, "B": 2, "A": 3},
				{"D": 1},
				{"D": 1},
				{"D": 1, "C": 2},
				{"D": 2, "C": 2, "B": 3},
				{"D": 1, "C": 3, "B": 3, "A": 4},
				{"E": 1},
				{"E": 1},
				{"E": 2, "D": 2},
				{"E": 1, "D": 2},
				{"E": 2, "D": 2, "C": 3},
				{"E": 1, "D": 2, "C": 3, "B": 3},
				{"E": 2, "D": 2, "C": 3, "B": 4, "A": 5},
				{"A": 1, "B": 2, "C": 3, "D": 4, "E": 5},
				{"A": 1, "B": 2, "C": 3, "D": 4},
				{"F": 1},
			},
			current: []string{"A", "B", "C", "D", "E", "F"},
			updated: []string{"A", "B", "C", "D", "E", "F", "G"},
		},
		{
			name: "multiple votes, new choices",
			ballots: []schulze.Ballot[string]{
				{"A": 1},
				{"A": 1},
				{"A": 1},
				{"A": 1},
				{"A": 1},
				{"A": 1},
				{"A": 1},
				{"A": 1},
				{"A": 1},
				{"A": 1},
				{"B": 1, "A": 2},
				{"B": 1},
				{"B": 1, "A": 2},
				{"B": 1, "A": 2},
				{"B": 1, "A": 2},
				{"B": 1, "A": 2},
				{"B": 1, "A": 2},
				{"C": 1},
				{"C": 1},
				{"C": 1},
				{"C": 1, "B": 2},
				{"C": 2, "B": 2, "A": 3},
				{"D": 1},
				{"D": 1},
				{"D": 1, "C": 2},
				{"D": 2, "C": 2, "B": 3},
				{"D": 1, "C": 3, "B": 3, "A": 4},
				{"E": 1},
				{"E": 1},
				{"E": 2, "D": 2},
				{"E": 1, "D": 2},
				{"E": 2, "D": 2, "C": 3},
				{"E": 1, "D": 2, "C": 3, "B": 3},
				{"E": 2, "D": 2, "C": 3, "B": 4, "A": 5},
				{"A": 1, "B": 2, "C": 3, "D": 4, "E": 5},
				{"A": 1, "B": 2, "C": 3, "D": 4},
				{"F": 1},
			},
			current: []string{"A", "B", "C", "D", "E", "F"},
			updated: []string{"G", "A", "B", "H", "C", "D", "E", "F", "I", "J"},
		},
		{
			name: "multiple votes, new, remove and swap choices",
			ballots: []schulze.Ballot[string]{
				{"A": 1},
				{"A": 1},
				{"A": 1},
				{"A": 1},
				{"A": 1},
				{"A": 1},
				{"A": 1},
				{"A": 1},
				{"A": 1},
				{"A": 1},
				{"B": 1, "A": 2},
				{"B": 1},
				{"B": 1, "A": 2},
				{"B": 1, "A": 2},
				{"B": 1, "A": 2},
				{"B": 1, "A": 2},
				{"B": 1, "A": 2},
				{"C": 1},
				{"C": 1},
				{"C": 1},
				{"C": 1, "B": 2},
				{"C": 2, "B": 2, "A": 3},
				{"D": 1},
				{"D": 1},
				{"D": 1, "C": 2},
				{"D": 2, "C": 2, "B": 3},
				{"D": 1, "C": 3, "B": 3, "A": 4},
				{"E": 1},
				{"E": 1},
				{"E": 2, "D": 2},
				{"E": 1, "D": 2},
				{"E": 2, "D": 2, "C": 3},
				{"E": 1, "D": 2, "C": 3, "B": 3},
				{"E": 2, "D": 2, "C": 3, "B": 4, "A": 5},
				{"A": 1, "B": 2, "C": 3, "D": 4, "E": 5},
				{"A": 1, "B": 2, "C": 3, "D": 4},
				{"F": 1},
			},
			current: []string{"A", "B", "C", "D", "E", "F", "G", "H"},
			updated: []string{"I", "A", "C", "H", "J", "K", "D", "B", "F", "L", "M"},
		},
		{
			name:    "thousand random votes, new, remove and swap choices",
			ballots: randomBallots(t, []string{"A", "B", "C", "D", "E", "F"}, 1000),
			current: []string{"A", "B", "C", "D", "E", "F"},
			updated: []string{"G", "A", "C", "H", "D", "B", "F", "I", "J"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("functional", func(t *testing.T) {
				currentPreferences := schulze.NewPreferences[int](len(tc.current))
				for _, b := range tc.ballots {
					if _, err := schulze.Vote(currentPreferences, tc.current, b); err != nil {
						t.Fatal(err)
					}
				}
				updatedChoicesCount := len(tc.updated)
				validationPreferences := schulze.NewPreferences[int](updatedChoicesCount)
				for _, b := range tc.ballots {
					b := removeChoices(b, removedChoices(tc.current, tc.updated))
					if _, err := schulze.Vote(validationPreferences, tc.updated, b); err != nil {
						t.Fatal(err)
					}
				}

				updatedPreferences := schulze.SetChoices(currentPreferences, tc.current, tc.updated)

				validatePreferences(t, updatedPreferences, validationPreferences, currentPreferences, tc.current, tc.updated)
			})
			t.Run("Voting", func(t *testing.T) {
				currentVoting := schulze.NewVoting(tc.current)
				for _, b := range tc.ballots {
					if _, err := currentVoting.Vote(b); err != nil {
						t.Fatal(err)
					}
				}
				currentPreferences := currentVoting.Preferences()
				validationVoting := schulze.NewVoting(tc.updated)
				for _, b := range tc.ballots {
					b := removeChoices(b, removedChoices(tc.current, tc.updated))
					if _, err := validationVoting.Vote(b); err != nil {
						t.Fatal(err)
					}
				}

				validationPreferences := validationVoting.Preferences()

				currentVoting.SetChoices(tc.updated)
				updatedPreferences := currentVoting.Preferences()

				validatePreferences(t, updatedPreferences, validationPreferences, currentPreferences, tc.current, tc.updated)
			})
		})
	}
}

func BenchmarkNewVoting(b *testing.B) {
	choices := newChoices(1000)

	for b.Loop() {
		_ = schulze.NewVoting(choices)
	}
}

func BenchmarkVoting_Vote(b *testing.B) {
	v := schulze.NewVoting(newChoices(1000))

	for b.Loop() {
		if _, err := v.Vote(schulze.Ballot[string]{
			"a": 1,
		}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVote(b *testing.B) {
	const choicesCount = 1000

	choices := newChoices(choicesCount)
	preferences := schulze.NewPreferences[int](choicesCount)

	for b.Loop() {
		if _, err := schulze.Vote(preferences, choices, schulze.Ballot[string]{
			"a": 1,
		}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVoting_Results(b *testing.B) {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))

	const choicesCount = 1000

	choices := newChoices(choicesCount)

	v := schulze.NewVoting(choices)

	for range 1000 {
		ballot := make(schulze.Ballot[string])
		ballot[choices[random.Intn(choicesCount)]] = 1
		ballot[choices[random.Intn(choicesCount)]] = 1
		ballot[choices[random.Intn(choicesCount)]] = 2
		ballot[choices[random.Intn(choicesCount)]] = 3
		ballot[choices[random.Intn(choicesCount)]] = 20
		ballot[choices[random.Intn(choicesCount)]] = 20
		if _, err := v.Vote(ballot); err != nil {
			b.Fatal(err)
		}
	}

	for b.Loop() {
		_, _, _ = v.Compute()
	}
}

func BenchmarkResults(b *testing.B) {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))

	const choicesCount = 1000

	choices := newChoices(choicesCount)
	preferences := schulze.NewPreferences[int](choicesCount)

	for range 1000 {
		ballot := make(schulze.Ballot[string])
		ballot[choices[random.Intn(choicesCount)]] = 1
		ballot[choices[random.Intn(choicesCount)]] = 1
		ballot[choices[random.Intn(choicesCount)]] = 2
		ballot[choices[random.Intn(choicesCount)]] = 3
		ballot[choices[random.Intn(choicesCount)]] = 20
		ballot[choices[random.Intn(choicesCount)]] = 20
		if _, err := schulze.Vote(preferences, choices, ballot); err != nil {
			b.Fatal(err)
		}
	}

	for b.Loop() {
		_, _, _ = schulze.Compute(preferences, choices)
	}
}

func newChoices(count int) []string {
	choices := make([]string, 0, count)
	for i := range count {
		choices = append(choices, strconv.FormatInt(int64(i), 36))
	}
	return choices
}

func randomBallots[C comparable](t *testing.T, choices []C, count int) []schulze.Ballot[C] {
	t.Helper()

	seed := time.Now().UnixNano()
	t.Logf("random ballots seed: %v", seed)

	random := rand.New(rand.NewSource(seed))

	ballots := make([]schulze.Ballot[C], 0, count)

	choicesLength := len(choices)
	for range count {
		b := make(schulze.Ballot[C])
		for range choicesLength {
			b[choices[random.Intn(choicesLength)]] = random.Intn(choicesLength)
		}
		ballots = append(ballots, b)
	}

	return ballots
}

func removedChoices[C comparable](old, new []C) (removed []C) {
	for _, c := range old {
		if !slices.Contains(new, c) {
			removed = append(removed, c)
		}
	}
	return removed
}

func removeChoices[C comparable](b schulze.Ballot[C], choices []C) schulze.Ballot[C] {
	r := make(map[C]int)
	for c, v := range b {
		if slices.Contains(choices, c) {
			continue
		}
		r[c] = v
	}
	return r
}

func fprintPreferences[C comparable](w io.Writer, choices []C, preferences []int) (int, error) {
	var width int
	for _, c := range choices {
		l := len(fmt.Sprint(c))
		if l > width {
			width = l
		}
	}
	for _, p := range preferences {
		l := len(strconv.Itoa(p))
		if l > width {
			width = l
		}
	}
	format := fmt.Sprintf("%%%vv ", width)
	var count int
	write := func(v string) error {
		n, err := fmt.Fprint(w, v)
		if err != nil {
			return err
		}
		count += n
		return nil
	}

	if err := write(fmt.Sprintf(format, "")); err != nil {
		return count, err
	}
	for _, c := range choices {
		if err := write(fmt.Sprintf(format, c)); err != nil {
			return count, err
		}
	}
	if err := write("\n"); err != nil {
		return count, err
	}

	m := matrix(preferences)

	for i, col := range m {
		if err := write(fmt.Sprintf(format, choices[i])); err != nil {
			return count, err
		}
		for _, p := range col {
			if err := write(fmt.Sprintf(format, p)); err != nil {
				return count, err
			}
		}
		if err := write("\n"); err != nil {
			return count, err
		}
	}

	return count, nil
}

func sprintPreferences[C comparable](choices []C, preferences []int) string {
	var buf bytes.Buffer
	_, _ = fprintPreferences(&buf, choices, preferences)
	return buf.String()
}

func matrix(preferences []int) [][]int {
	l := len(preferences)
	choicesCount := floorSqrt(l)
	if choicesCount*choicesCount != l {
		return nil
	}

	matrix := make([][]int, 0, choicesCount)

	for i := range choicesCount {
		matrix = append(matrix, preferences[i*choicesCount:(i+1)*choicesCount])
	}
	return matrix
}

func floorSqrt(x int) int {
	if x == 0 || x == 1 {
		return x
	}
	start := 1
	end := x / 2
	ans := 0
	for start <= end {
		mid := (start + end) / 2
		if mid*mid == x {
			return mid
		}
		if mid*mid < x {
			start = mid + 1
			ans = mid
		} else {
			end = mid - 1
		}
	}
	return ans
}

func TestVoteRanked(t *testing.T) {
	choices := []string{"A", "B", "C", "D"}
	v1 := schulze.NewVoting(choices)
	v2 := schulze.NewVoting(choices)

	// Vote via map
	_, err := v1.Vote(schulze.Ballot[string]{
		"A": 1,
		"B": 2,
		"C": 2,
		"D": 3,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Vote via ranked slice
	rec2, err := v2.VoteRanked([][]string{
		{"A"},
		{"B", "C"},
		{"D"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(v1.Preferences(), v2.Preferences()) {
		t.Fatalf("preferences mismatch:\nmap:    %v\nranked: %v", v1.Preferences(), v2.Preferences())
	}

	// Test unvote with ranked record
	if err := v2.Unvote(rec2); err != nil {
		t.Fatal(err)
	}
	for _, p := range v2.Preferences() {
		if p != 0 {
			t.Fatalf("expected 0 after unvote, got %v", p)
		}
	}

	// Test invalid ranked vote with duplicates
	_, err = v2.VoteRanked([][]string{
		{"A", "B"},
		{"A"},
	})
	if err == nil {
		t.Fatal("expected error on duplicate choice in VoteRanked")
	}

	// Test unknown choice
	_, err = v2.VoteRanked([][]string{
		{"X"},
	})
	if err == nil {
		t.Fatal("expected error on unknown choice in VoteRanked")
	}
}

func TestVoteFrom(t *testing.T) {
	type Survey struct {
		Scores map[string]int
	}

	choices := []string{"A", "B", "C"}
	v := schulze.NewVoting(choices)

	survey := Survey{
		Scores: map[string]int{
			"A": 10,
			"B": 5,
		},
	}

	rec, err := v.VoteFrom(survey, func(s Survey, c string) (int, bool) {
		score, ok := s.Scores[c]
		if !ok {
			return 0, false
		}
		// Invert score so higher score = lower rank number
		return 100 - score, true
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rec) != 3 { // 2 ranked + 1 unranked
		t.Fatalf("unexpected record length %d", len(rec))
	}
}

func TestMap(t *testing.T) {
	choices := []string{"apple", "banana", "cherry"}
	v := schulze.NewVoting(choices)

	rec, err := v.Vote(schulze.Ballot[string]{
		"apple":  1,
		"banana": 2,
		"cherry": 3,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Map Record
	mappedRec := rec.Map(func(s string) int {
		return len(s)
	})
	if mappedRec[0][0] != 5 { // "apple" len
		t.Fatalf("expected 5, got %v", mappedRec[0][0])
	}

	// Map Voting
	mappedV := v.Map(func(s string) string {
		return "choice:" + s
	})
	if mappedV.Choices()[0] != "choice:apple" {
		t.Fatalf("expected choice:apple, got %v", mappedV.Choices()[0])
	}

	// Map Compute Results and Duels
	results, duelsIt, _ := v.Compute()
	mappedRes := results[0].Map(func(s string) string {
		return "winner:" + s
	})
	if mappedRes.Choice != "winner:apple" {
		t.Fatalf("expected winner:apple, got %v", mappedRes.Choice)
	}

	d := duelsIt()
	if d != nil {
		mappedDuel := d.Map(func(s string) string {
			return "c:" + s
		})
		if mappedDuel.Left.Choice != "c:apple" {
			t.Fatalf("expected c:apple, got %v", mappedDuel.Left.Choice)
		}
	}
}

func TestDuelsIterator_Seq(t *testing.T) {
	choices := []string{"A", "B", "C"}
	v := schulze.NewVoting(choices)

	_, _ = v.Vote(schulze.Ballot[string]{
		"A": 1,
		"B": 2,
		"C": 3,
	})

	count := 0
	for duel := range v.Duels() {
		if duel == nil {
			t.Fatal("unexpected nil duel")
		}
		count++
	}
	// 3 choices -> 3 pairs (A-B, A-C, B-C)
	if count != 3 {
		t.Fatalf("expected 3 duels, got %d", count)
	}

	// Test Duels standalone function
	count2 := 0
	for duel := range schulze.Duels(v.Preferences(), v.Choices()) {
		if duel == nil {
			t.Fatal("unexpected nil duel")
		}
		count2++
	}
	if count2 != 3 {
		t.Fatalf("expected 3 duels from Duels(), got %d", count2)
	}
}

func TestGenericNumberPreferences(t *testing.T) {
	choices := []string{"A", "B", "C"}

	// uint32 preferences
	prefsU32 := schulze.NewPreferences[uint32](len(choices))
	rec, err := schulze.Vote(prefsU32, choices, schulze.Ballot[string]{
		"A": 1,
		"B": 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	resU32, _, tie := schulze.Compute(prefsU32, choices)
	if tie || resU32[0].Choice != "A" {
		t.Fatalf("expected A winner, got %v (tie=%v)", resU32[0].Choice, tie)
	}

	// uint16 preferences
	prefsU16 := schulze.NewPreferences[uint16](len(choices))
	_, err = schulze.Vote(prefsU16, choices, schulze.Ballot[string]{
		"A": 1,
		"B": 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	resU16, _, _ := schulze.Compute(prefsU16, choices)
	if resU16[0].Choice != "A" {
		t.Fatalf("expected A winner, got %v", resU16[0].Choice)
	}

	// Unvote on uint32
	if err := schulze.Unvote(prefsU32, choices, rec); err != nil {
		t.Fatal(err)
	}
	for _, p := range prefsU32 {
		if p != 0 {
			t.Fatalf("expected 0, got %d", p)
		}
	}
}

func TestExportImportPreferences(t *testing.T) {
	choices := []string{"A", "B"}
	v := schulze.NewVoting(choices)
	_, _ = v.Vote(schulze.Ballot[string]{"A": 1, "B": 2})

	exported := v.ExportPreferences[uint32]()
	if len(exported) != 4 {
		t.Fatalf("expected 4 exported prefs, got %d", len(exported))
	}

	v2 := schulze.NewVoting(choices)
	if err := v2.ImportPreferences(exported); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(v.Preferences(), v2.Preferences()) {
		t.Fatalf("preferences mismatch after import: %v vs %v", v.Preferences(), v2.Preferences())
	}
}

func FuzzVote(f *testing.F) {
	f.Add("A", "B", "C", 1, 2, 3)
	f.Add("X", "Y", "Z", 10, 5, 5)

	f.Fuzz(func(t *testing.T, c1, c2, c3 string, r1, r2, r3 int) {
		choices := []string{c1, c2, c3}
		if schulze.ValidateChoices(choices) != nil {
			return
		}

		v := schulze.NewVoting(choices)
		ballot := schulze.Ballot[string]{
			c1: r1,
			c2: r2,
			c3: r3,
		}

		rec, err := v.Vote(ballot)
		if err != nil {
			t.Fatalf("unexpected error voting: %v", err)
		}

		results, _, _ := v.Compute()
		if len(results) != len(choices) {
			t.Fatalf("expected %d results, got %d", len(choices), len(results))
		}

		if err := v.Unvote(rec); err != nil {
			t.Fatalf("unexpected error unvoting: %v", err)
		}
	})
}

func FuzzCompute(f *testing.F) {
	f.Add(1, 2, 0, 3)
	f.Fuzz(func(t *testing.T, p0, p1, p2, p3 int) {
		if p0 < 0 || p1 < 0 || p2 < 0 || p3 < 0 {
			return
		}
		choices := []string{"A", "B"}
		prefs := []int{p0, p1, p2, p3}
		results, duels, _ := schulze.Compute(prefs, choices)
		if len(results) != 2 {
			t.Fatalf("expected 2 results, got %d", len(results))
		}
		d := duels()
		if d == nil {
			t.Fatal("expected duel")
		}
	})
}

func BenchmarkVoteRanked(b *testing.B) {
	const choicesCount = 1000
	choices := newChoices(choicesCount)
	preferences := schulze.NewPreferences[int](choicesCount)
	ranked := [][]string{{"0"}, {"1", "2"}, {"3"}}

	for b.Loop() {
		if _, err := schulze.VoteRanked(preferences, choices, ranked); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVote_Sizes(b *testing.B) {
	for _, nChoices := range []int{5, 20, 100, 1000} {
		b.Run(strconv.Itoa(nChoices), func(b *testing.B) {
			choices := newChoices(nChoices)
			preferences := schulze.NewPreferences[int](nChoices)
			ballot := schulze.Ballot[string]{
				choices[0]: 1,
				choices[1]: 2,
			}

			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				if _, err := schulze.Vote(preferences, choices, ballot); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkVoteRanked_Sizes(b *testing.B) {
	for _, nChoices := range []int{5, 20, 100, 1000} {
		b.Run(strconv.Itoa(nChoices), func(b *testing.B) {
			choices := newChoices(nChoices)
			preferences := schulze.NewPreferences[int](nChoices)
			ranked := [][]string{{choices[0]}, {choices[1]}}

			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				if _, err := schulze.VoteRanked(preferences, choices, ranked); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func TestRelaxRowDirect(t *testing.T) {
	rowI := []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	rowJ := []int{5, 25, 15, 35, 60, 10, 20, 95, 30, 50}
	expected := append([]int(nil), rowJ...)
	jip := 45

	for k := range rowI {
		m := min(jip, rowI[k])
		if m > expected[k] {
			expected[k] = m
		}
	}

	schulze.RelaxRow(unsafe.Pointer(&rowI[0]), unsafe.Pointer(&rowJ[0]), jip, len(rowI))

	t.Logf("Expected: %v", expected)
	t.Logf("Actual:   %v", rowJ)
	if !reflect.DeepEqual(expected, rowJ) {
		t.Fatalf("mismatch in relaxRow: expected %v, got %v", expected, rowJ)
	}
}

func TestWinner(t *testing.T) {
	choices := []string{"A", "B", "C", "D"}
	v := schulze.NewVoting(choices)

	// A beats B, C, D in every pairwise matchup (Condorcet winner)
	_, _ = v.Vote(schulze.Ballot[string]{"A": 1, "B": 2, "C": 3, "D": 4})
	_, _ = v.Vote(schulze.Ballot[string]{"A": 1, "C": 2, "B": 3, "D": 4})
	_, _ = v.Vote(schulze.Ballot[string]{"A": 1, "D": 2, "B": 3, "C": 4})

	winner, tie := v.Winner()
	if tie {
		t.Fatal("expected no tie for Condorcet winner")
	}
	if winner.Choice != "A" {
		t.Fatalf("expected A, got %v", winner.Choice)
	}

	// Test election without Condorcet winner (cycle with unequal margin)
	vCycle := schulze.NewVoting([]string{"A", "B", "C"})
	_, _ = vCycle.Vote(schulze.Ballot[string]{"A": 1, "B": 2, "C": 3})
	_, _ = vCycle.Vote(schulze.Ballot[string]{"A": 1, "B": 2, "C": 3})
	_, _ = vCycle.Vote(schulze.Ballot[string]{"B": 1, "C": 2, "A": 3})
	_, _ = vCycle.Vote(schulze.Ballot[string]{"C": 1, "A": 2, "B": 3})

	winnerCycle, tieCycle := vCycle.Winner()
	if tieCycle {
		t.Fatal("expected no tie in cycle with clear winner")
	}
	results, _, _ := vCycle.Compute()
	if winnerCycle.Choice != results[0].Choice {
		t.Fatalf("expected %v, got %v", results[0].Choice, winnerCycle.Choice)
	}
}

func TestCompute_Large(t *testing.T) {
	const nChoices = 100
	choices := newChoices(nChoices)
	v := schulze.NewVoting(choices)

	for i := range 50 {
		ballot := make(schulze.Ballot[string])
		for j := range 10 {
			ballot[choices[(i+j*7)%nChoices]] = j + 1
		}
		if _, err := v.Vote(ballot); err != nil {
			t.Fatal(err)
		}
	}

	results, duels, _ := v.Compute()
	if len(results) != nChoices {
		t.Fatalf("expected %d results, got %d", nChoices, len(results))
	}

	var count int
	for range v.Duels() {
		count++
	}
	expectedDuels := (nChoices * (nChoices - 1)) / 2
	if count != expectedDuels {
		t.Fatalf("expected %d duels, got %d", expectedDuels, count)
	}

	// Verify duels iterator returns the same count
	var iterCount int
	for {
		d := duels()
		if d == nil {
			break
		}
		iterCount++
	}
	if iterCount != expectedDuels {
		t.Fatalf("expected %d duels from iterator, got %d", expectedDuels, iterCount)
	}
}

func BenchmarkCompute_Sizes(b *testing.B) {
	for _, nChoices := range []int{5, 20, 100, 500, 1000} {
		b.Run(strconv.Itoa(nChoices), func(b *testing.B) {
			choices := newChoices(nChoices)
			preferences := schulze.NewPreferences[int](nChoices)
			for i := range nChoices {
				for j := range nChoices {
					if i != j {
						preferences[i*nChoices+j] = (i*7 + j*13) % 100
					}
				}
			}

			for b.Loop() {
				_, _, _ = schulze.Compute(preferences, choices)
			}
		})
	}
}

func BenchmarkWinner_Sizes(b *testing.B) {
	for _, nChoices := range []int{5, 20, 100, 500, 1000} {
		b.Run(strconv.Itoa(nChoices), func(b *testing.B) {
			choices := newChoices(nChoices)
			preferences := schulze.NewPreferences[int](nChoices)
			// Make candidate 0 the Condorcet winner
			for j := 1; j < nChoices; j++ {
				preferences[0*nChoices+j] = 100
				preferences[j*nChoices+0] = 10
			}

			for b.Loop() {
				_, _ = schulze.Winner(preferences, choices)
			}
		})
	}
}

func BenchmarkSetChoices_Sizes(b *testing.B) {
	for _, nChoices := range []int{5, 20, 100, 1000} {
		b.Run(strconv.Itoa(nChoices), func(b *testing.B) {
			choices := newChoices(nChoices)
			updated := newChoices(nChoices)
			if nChoices >= 2 {
				updated[0], updated[1] = updated[1], updated[0]
			}
			preferences := schulze.NewPreferences[int](nChoices)

			for b.Loop() {
				_ = schulze.SetChoices(preferences, choices, updated)
			}
		})
	}
}

func TestVoteWeighted(t *testing.T) {
	choices := []string{"A", "B", "C", "D"}
	ballot := schulze.Ballot[string]{
		"A": 1,
		"B": 2,
		"C": 3,
	}

	t.Run("invalid weight", func(t *testing.T) {
		prefs := schulze.NewPreferences[int](len(choices))
		if _, err := schulze.VoteWeighted(prefs, choices, ballot, 0); !errors.Is(err, schulze.ErrInvalidWeight) {
			t.Errorf("expected ErrInvalidWeight for weight 0, got %v", err)
		}
		if _, err := schulze.VoteWeighted(prefs, choices, ballot, -5); !errors.Is(err, schulze.ErrInvalidWeight) {
			t.Errorf("expected ErrInvalidWeight for negative weight, got %v", err)
		}
	})

	t.Run("equivalence to repeated Vote", func(t *testing.T) {
		for weight := 1; weight <= 10; weight++ {
			prefsWeighted := schulze.NewPreferences[int](len(choices))
			recWeighted, err := schulze.VoteWeighted(prefsWeighted, choices, ballot, weight)
			if err != nil {
				t.Fatalf("weight %d: VoteWeighted: %v", weight, err)
			}

			prefsLoop := schulze.NewPreferences[int](len(choices))
			var recLoop schulze.Record[string]
			for range weight {
				recLoop, err = schulze.Vote(prefsLoop, choices, ballot)
				if err != nil {
					t.Fatalf("weight %d: Vote: %v", weight, err)
				}
			}

			if !slices.Equal(prefsWeighted, prefsLoop) {
				t.Errorf("weight %d: preferences mismatch: weighted=%v, loop=%v", weight, prefsWeighted, prefsLoop)
			}
			if !reflect.DeepEqual(recWeighted, recLoop) {
				t.Errorf("weight %d: record mismatch: weighted=%v, loop=%v", weight, recWeighted, recLoop)
			}

			// Test Voting method equivalence
			v := schulze.NewVoting(choices)
			recV, err := v.VoteWeighted(ballot, weight)
			if err != nil {
				t.Fatalf("weight %d: Voting.VoteWeighted: %v", weight, err)
			}
			if !slices.Equal(v.Preferences(), prefsWeighted) {
				t.Errorf("weight %d: Voting preferences mismatch: %v != %v", weight, v.Preferences(), prefsWeighted)
			}
			if !reflect.DeepEqual(recV, recWeighted) {
				t.Errorf("weight %d: Voting record mismatch: %v != %v", weight, recV, recWeighted)
			}
		}
	})
}

func TestVoteRankedWeighted(t *testing.T) {
	choices := []string{"A", "B", "C", "D", "E"}
	ranked := [][]string{
		{"A", "B"},
		{"C"},
	}

	t.Run("invalid weight", func(t *testing.T) {
		prefs := schulze.NewPreferences[int](len(choices))
		if _, err := schulze.VoteRankedWeighted(prefs, choices, ranked, 0); !errors.Is(err, schulze.ErrInvalidWeight) {
			t.Errorf("expected ErrInvalidWeight for weight 0, got %v", err)
		}
		if _, err := schulze.VoteRankedWeighted(prefs, choices, ranked, -1); !errors.Is(err, schulze.ErrInvalidWeight) {
			t.Errorf("expected ErrInvalidWeight for negative weight, got %v", err)
		}
	})

	t.Run("equivalence to repeated VoteRanked", func(t *testing.T) {
		for weight := 1; weight <= 10; weight++ {
			prefsWeighted := schulze.NewPreferences[int](len(choices))
			recWeighted, err := schulze.VoteRankedWeighted(prefsWeighted, choices, ranked, weight)
			if err != nil {
				t.Fatalf("weight %d: VoteRankedWeighted: %v", weight, err)
			}

			prefsLoop := schulze.NewPreferences[int](len(choices))
			var recLoop schulze.Record[string]
			for range weight {
				recLoop, err = schulze.VoteRanked(prefsLoop, choices, ranked)
				if err != nil {
					t.Fatalf("weight %d: VoteRanked: %v", weight, err)
				}
			}

			if !slices.Equal(prefsWeighted, prefsLoop) {
				t.Errorf("weight %d: preferences mismatch: weighted=%v, loop=%v", weight, prefsWeighted, prefsLoop)
			}
			if !reflect.DeepEqual(recWeighted, recLoop) {
				t.Errorf("weight %d: record mismatch: weighted=%v, loop=%v", weight, recWeighted, recLoop)
			}

			// Test Voting method equivalence
			v := schulze.NewVoting(choices)
			recV, err := v.VoteRankedWeighted(ranked, weight)
			if err != nil {
				t.Fatalf("weight %d: Voting.VoteRankedWeighted: %v", weight, err)
			}
			if !slices.Equal(v.Preferences(), prefsWeighted) {
				t.Errorf("weight %d: Voting preferences mismatch: %v != %v", weight, v.Preferences(), prefsWeighted)
			}
			if !reflect.DeepEqual(recV, recWeighted) {
				t.Errorf("weight %d: Voting record mismatch: %v != %v", weight, recV, recWeighted)
			}
		}
	})
}

func TestUnvoteWeighted(t *testing.T) {
	choices := []string{"A", "B", "C", "D"}
	ranked := [][]string{
		{"A"},
		{"B", "C"},
	}

	t.Run("invalid weight", func(t *testing.T) {
		prefs := schulze.NewPreferences[int](len(choices))
		rec, _ := schulze.VoteRanked(prefs, choices, ranked)
		if err := schulze.UnvoteWeighted(prefs, choices, rec, 0); !errors.Is(err, schulze.ErrInvalidWeight) {
			t.Errorf("expected ErrInvalidWeight for weight 0, got %v", err)
		}
		if err := schulze.UnvoteWeighted(prefs, choices, rec, -3); !errors.Is(err, schulze.ErrInvalidWeight) {
			t.Errorf("expected ErrInvalidWeight for negative weight, got %v", err)
		}
	})

	t.Run("equivalence to repeated Unvote", func(t *testing.T) {
		for weight := 1; weight <= 10; weight++ {
			// 1. Vote weighted then unvote weighted
			prefsWeighted := schulze.NewPreferences[int](len(choices))
			rec, err := schulze.VoteRankedWeighted(prefsWeighted, choices, ranked, weight)
			if err != nil {
				t.Fatalf("VoteRankedWeighted: %v", err)
			}
			if err := schulze.UnvoteWeighted(prefsWeighted, choices, rec, weight); err != nil {
				t.Fatalf("UnvoteWeighted: %v", err)
			}

			// 2. Vote loop then unvote loop
			prefsLoop := schulze.NewPreferences[int](len(choices))
			var recLoop schulze.Record[string]
			for range weight {
				recLoop, err = schulze.VoteRanked(prefsLoop, choices, ranked)
				if err != nil {
					t.Fatalf("VoteRanked: %v", err)
				}
			}
			for range weight {
				if err := schulze.Unvote(prefsLoop, choices, recLoop); err != nil {
					t.Fatalf("Unvote: %v", err)
				}
			}

			// Both must be equal and back to zero
			emptyPrefs := schulze.NewPreferences[int](len(choices))
			if !slices.Equal(prefsWeighted, emptyPrefs) {
				t.Errorf("weight %d: expected zero preferences after UnvoteWeighted, got %v", weight, prefsWeighted)
			}
			if !slices.Equal(prefsWeighted, prefsLoop) {
				t.Errorf("weight %d: preferences mismatch: weighted=%v, loop=%v", weight, prefsWeighted, prefsLoop)
			}

			// Test Voting method
			v := schulze.NewVoting(choices)
			recV, err := v.VoteRankedWeighted(ranked, weight)
			if err != nil {
				t.Fatalf("Voting.VoteRankedWeighted: %v", err)
			}
			if err := v.UnvoteWeighted(recV, weight); err != nil {
				t.Fatalf("Voting.UnvoteWeighted: %v", err)
			}
			if !slices.Equal(v.Preferences(), emptyPrefs) {
				t.Errorf("weight %d: Voting preferences not empty: %v", weight, v.Preferences())
			}
		}
	})
}

func BenchmarkVoteWeighted(b *testing.B) {
	choices := make([]string, 100)
	for i := range choices {
		choices[i] = fmt.Sprintf("choice_%d", i)
	}
	ranked := make([][]string, 20)
	for i := range ranked {
		ranked[i] = choices[i*5 : (i+1)*5]
	}

	b.Run("single weighted vote W=10", func(b *testing.B) {
		prefs := schulze.NewPreferences[int](len(choices))
		b.ResetTimer()
		b.ReportAllocs()
		for b.Loop() {
			_, _ = schulze.VoteRankedWeighted(prefs, choices, ranked, 10)
		}
	})

	b.Run("loop of 10 votes W=1", func(b *testing.B) {
		prefs := schulze.NewPreferences[int](len(choices))
		b.ResetTimer()
		b.ReportAllocs()
		for b.Loop() {
			for range 10 {
				_, _ = schulze.VoteRanked(prefs, choices, ranked)
			}
		}
	})
}
