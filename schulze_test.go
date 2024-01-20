// Copyright (c) 2021, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schulze_test

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"reflect"
	"strconv"
	"testing"
	"time"

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
				preferences := schulze.NewPreferences(len(tc.choices))

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

				result, tie := schulze.Compute(preferences, tc.choices)
				if tie != tc.tie {
					t.Errorf("got tie %v, want %v", tie, tc.tie)
				}
				if !reflect.DeepEqual(result, tc.result) {
					t.Errorf("got result %+v, want %+v", result, tc.result)
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

				result, tie := v.Compute()
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

func TestUnvote_afterSetChoices(t *testing.T) {
	t.Run("add", func(t *testing.T) {
		choices := []string{"A", "B", "C"}
		preferences := schulze.NewPreferences(len(choices))

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

	t.Run("remove", func(t *testing.T) {
		choices := []string{"A", "B", "C"}
		preferences := schulze.NewPreferences(len(choices))

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

	t.Run("rearrange", func(t *testing.T) {
		choices := []string{"A", "B", "C"}
		preferences := schulze.NewPreferences(len(choices))

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

	t.Run("complex", func(t *testing.T) {
		choices := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"}
		preferences := schulze.NewPreferences(len(choices))

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
				currentPreferences := schulze.NewPreferences(len(tc.current))
				for _, b := range tc.ballots {
					if _, err := schulze.Vote(currentPreferences, tc.current, b); err != nil {
						t.Fatal(err)
					}
				}
				updatedChoicesCount := len(tc.updated)
				validationPreferences := schulze.NewPreferences(updatedChoicesCount)
				for _, b := range tc.ballots {
					b := removeChoices(b, removedChoices(tc.current, tc.updated))
					if _, err := schulze.Vote(validationPreferences, tc.updated, b); err != nil {
						t.Fatal(err)
					}
				}

				// annulate wins for the unknown choices in validation preferences
				// for i := 0; i < updatedChoicesCount; i++ {
				// 	for _, j := range indexesOfNewChoices(tc.current, tc.updated) {
				// 		if i == 0 && j == 0 {
				// 			continue
				// 		}
				// 		validationPreferences[i*updatedChoicesCount+j] = 0
				// 	}
				// }

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
				// annulate wins for the unknown choices in validation preferences
				// updatedChoicesCount := len(tc.updated)
				// for i := 0; i < updatedChoicesCount; i++ {
				// 	for _, j := range indexesOfNewChoices(tc.current, tc.updated) {
				// 		if i == 0 && j == 0 {
				// 			continue
				// 		}
				// 		validationPreferences[i*updatedChoicesCount+j] = 0
				// 	}
				// }

				currentVoting.SetChoices(tc.updated)
				updatedPreferences := currentVoting.Preferences()

				validatePreferences(t, updatedPreferences, validationPreferences, currentPreferences, tc.current, tc.updated)
			})
		})
	}
}

func BenchmarkNewVoting(b *testing.B) {
	choices := newChoices(1000)

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = schulze.NewVoting(choices)
	}
}

func BenchmarkVoting_Vote(b *testing.B) {
	v := schulze.NewVoting(newChoices(1000))

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
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
	preferences := schulze.NewPreferences(choicesCount)

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		if _, err := schulze.Vote(preferences, choices, schulze.Ballot[string]{
			"a": 1,
		}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVoting_Results(b *testing.B) {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))

	const choicesCount = 100

	choices := newChoices(choicesCount)

	v := schulze.NewVoting(choices)

	for i := 0; i < 1000; i++ {
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

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_, _ = v.Compute()
	}
}

func BenchmarkResults(b *testing.B) {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))

	const choicesCount = 100

	choices := newChoices(choicesCount)
	preferences := schulze.NewPreferences(choicesCount)

	for i := 0; i < 1000; i++ {
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

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_, _ = schulze.Compute(preferences, choices)
	}
}

func newChoices(count int) []string {
	choices := make([]string, 0, count)
	for i := 0; i < count; i++ {
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
	for i := 0; i < count; i++ {
		b := make(schulze.Ballot[C])
		for i := 0; i < choicesLength; i++ {
			b[choices[random.Intn(choicesLength)]] = random.Intn(choicesLength)
		}
		ballots = append(ballots, b)
	}

	return ballots
}

func removedChoices[C comparable](old, new []C) (removed []C) {
	for _, c := range old {
		if !contains(new, c) {
			removed = append(removed, c)
		}
	}
	return removed
}

func removeChoices[C comparable](b schulze.Ballot[C], choices []C) schulze.Ballot[C] {
	r := make(map[C]int)
	for c, v := range b {
		if contains(choices, c) {
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

	for i := 0; i < choicesCount; i++ {
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

func contains[C comparable](s []C, e C) bool {
	for _, x := range s {
		if x == e {
			return true
		}
	}
	return false
}
