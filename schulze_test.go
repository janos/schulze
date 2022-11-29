// Copyright (c) 2021, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schulze_test

import (
	"math/rand"
	"reflect"
	"strconv"
	"testing"
	"time"

	"resenje.org/schulze"
)

func TestVoting(t *testing.T) {
	for _, tc := range []struct {
		name    string
		choices []string
		ballots []schulze.Ballot[string]
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
				{Choice: "A", Index: 0, Wins: 0},
			},
		},
		{
			name:    "single option one vote",
			choices: []string{"A"},
			ballots: []schulze.Ballot[string]{
				{"A": 1},
			},
			result: []schulze.Result[string]{
				{Choice: "A", Index: 0, Wins: 0},
			},
		},
		{
			name:    "two options one vote",
			choices: []string{"A", "B"},
			ballots: []schulze.Ballot[string]{
				{"A": 1},
			},
			result: []schulze.Result[string]{
				{Choice: "A", Index: 0, Wins: 1},
				{Choice: "B", Index: 1, Wins: 0},
			},
		},
		{
			name:    "two options two votes",
			choices: []string{"A", "B"},
			ballots: []schulze.Ballot[string]{
				{"A": 1},
				{"A": 1, "B": 2},
			},
			result: []schulze.Result[string]{
				{Choice: "A", Index: 0, Wins: 1},
				{Choice: "B", Index: 1, Wins: 0},
			},
		},
		{
			name:    "three options three votes",
			choices: []string{"A", "B", "C"},
			ballots: []schulze.Ballot[string]{
				{"A": 1},
				{"A": 1, "B": 2},
				{"A": 1, "B": 2, "C": 3},
			},
			result: []schulze.Result[string]{
				{Choice: "A", Index: 0, Wins: 2},
				{Choice: "B", Index: 1, Wins: 1},
				{Choice: "C", Index: 2, Wins: 0},
			},
		},
		{
			name:    "tie",
			choices: []string{"A", "B", "C"},
			ballots: []schulze.Ballot[string]{
				{"A": 1},
				{"B": 1},
			},
			result: []schulze.Result[string]{
				{Choice: "A", Index: 0, Wins: 1},
				{Choice: "B", Index: 1, Wins: 1},
				{Choice: "C", Index: 2, Wins: 0},
			},
			tie: true,
		},
		{
			name:    "complex",
			choices: []string{"A", "B", "C", "D", "E"},
			ballots: []schulze.Ballot[string]{
				{"A": 1, "B": 1},
				{"B": 1, "C": 1, "A": 2},
				{"A": 1, "B": 2, "C": 2},
				{"A": 1, "B": 200, "C": 10},
			},
			result: []schulze.Result[string]{
				{Choice: "A", Index: 0, Wins: 4},
				{Choice: "B", Index: 1, Wins: 2},
				{Choice: "C", Index: 2, Wins: 2},
				{Choice: "D", Index: 3, Wins: 0},
				{Choice: "E", Index: 4, Wins: 0},
			},
		},
		{
			name:    "duplicate choice", // only the first of the duplicate choices should receive votes
			choices: []string{"A", "B", "C", "C", "C"},
			ballots: []schulze.Ballot[string]{
				{"A": 1, "B": 1},
				{"B": 1, "C": 1, "A": 2},
				{"A": 1, "B": 2, "C": 2},
				{"A": 1, "B": 200, "C": 10},
			},
			result: []schulze.Result[string]{
				{Choice: "A", Index: 0, Wins: 4},
				{Choice: "B", Index: 1, Wins: 2},
				{Choice: "C", Index: 2, Wins: 2},
				{Choice: "C", Index: 3, Wins: 0},
				{Choice: "C", Index: 4, Wins: 0},
			},
		},
		{
			name:    "example from wiki page",
			choices: []string{"A", "B", "C", "D", "E"},
			ballots: []schulze.Ballot[string]{
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
			result: []schulze.Result[string]{
				{Choice: "E", Index: 4, Wins: 4},
				{Choice: "A", Index: 0, Wins: 3},
				{Choice: "C", Index: 2, Wins: 2},
				{Choice: "B", Index: 1, Wins: 1},
				{Choice: "D", Index: 3, Wins: 0},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("functional", func(t *testing.T) {
				preferences := schulze.NewPreferences(len(tc.choices))

				for _, b := range tc.ballots {
					if err := schulze.Vote(preferences, tc.choices, b); err != nil {
						t.Fatal(err)
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
					if err := v.Vote(b); err != nil {
						t.Fatal(err)
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
		if err := v.Vote(schulze.Ballot[string]{
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
		if err := schulze.Vote(preferences, choices, schulze.Ballot[string]{
			"a": 1,
		}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVoting_Results(b *testing.B) {
	rand.Seed(time.Now().UnixNano())

	const choicesCount = 100

	choices := newChoices(choicesCount)

	v := schulze.NewVoting(choices)

	for i := 0; i < 1000; i++ {
		ballot := make(schulze.Ballot[string])
		ballot[choices[rand.Intn(choicesCount)]] = 1
		ballot[choices[rand.Intn(choicesCount)]] = 1
		ballot[choices[rand.Intn(choicesCount)]] = 2
		ballot[choices[rand.Intn(choicesCount)]] = 3
		ballot[choices[rand.Intn(choicesCount)]] = 20
		ballot[choices[rand.Intn(choicesCount)]] = 20
		if err := v.Vote(ballot); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_, _ = v.Compute()
	}
}

func BenchmarkResults(b *testing.B) {
	rand.Seed(time.Now().UnixNano())

	const choicesCount = 100

	choices := newChoices(choicesCount)
	preferences := schulze.NewPreferences(choicesCount)

	for i := 0; i < 1000; i++ {
		ballot := make(schulze.Ballot[string])
		ballot[choices[rand.Intn(choicesCount)]] = 1
		ballot[choices[rand.Intn(choicesCount)]] = 1
		ballot[choices[rand.Intn(choicesCount)]] = 2
		ballot[choices[rand.Intn(choicesCount)]] = 3
		ballot[choices[rand.Intn(choicesCount)]] = 20
		ballot[choices[rand.Intn(choicesCount)]] = 20
		if err := schulze.Vote(preferences, choices, ballot); err != nil {
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
