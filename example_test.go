// Copyright (c) 2021, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schulze_test

import (
	"fmt"
	"log"

	"resenje.org/schulze"
)

func ExampleVoting() {
	// Create a new voting.
	v := schulze.NewVoting([]string{"A", "B", "C"})

	// First vote.
	if _, err := v.Vote(schulze.Ballot[string]{
		"A": 1,
		"C": 2,
	}); err != nil {
		log.Fatal(err)
	}

	// Second vote.
	if _, err := v.Vote(schulze.Ballot[string]{
		"A": 1,
		"B": 1,
	}); err != nil {
		log.Fatal(err)
	}

	// Third vote.
	if _, err := v.Vote(schulze.Ballot[string]{
		"A": 1,
		"B": 2,
		"C": 2,
	}); err != nil {
		log.Fatal(err)
	}

	// Calculate the result.
	result, duels, tie := v.Compute()
	if tie {
		log.Fatal("Tie")
	}

	for duel := duels(); duel != nil; duel = duels() {
		winner, defeated := duel.Outcome()
		if winner == nil {
			fmt.Printf("Options %s and %s are in tie %v\n", duel.Left.Choice, duel.Right.Choice, duel.Left.Strength)
		} else {
			fmt.Printf("Options %s defeats %s by (%v - %v) = %v votes\n", winner.Choice, defeated.Choice, winner.Strength, defeated.Strength, duel.Left.Strength-defeated.Strength)
		}
	}

	fmt.Println("Winner:", result[0].Choice)

	// Output: Options A defeats B by (2 - 0) = 2 votes
	// Options A defeats C by (3 - 0) = 3 votes
	// Options B and C are in tie 0
	// Winner: A
}

func ExampleNewPreferences() {
	// Create a new voting.
	choices := []string{"A", "B", "C"}
	preferences := schulze.NewPreferences(len(choices))

	// First vote.
	if _, err := schulze.Vote(preferences, choices, schulze.Ballot[string]{
		"A": 1,
	}); err != nil {
		log.Fatal(err)
	}

	// Second vote.
	if _, err := schulze.Vote(preferences, choices, schulze.Ballot[string]{
		"A": 1,
		"B": 1,
		"C": 2,
	}); err != nil {
		log.Fatal(err)
	}

	// Calculate the result.
	result, _, tie := schulze.Compute(preferences, choices)
	if tie {
		log.Fatal("tie")
	}
	fmt.Println("winner:", result[0].Choice)

	// Output: winner: A
}
