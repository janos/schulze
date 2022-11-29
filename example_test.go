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
	if err := v.Vote(schulze.Ballot[string]{
		"A": 1,
	}); err != nil {
		log.Fatal(err)
	}

	// Second vote.
	if err := v.Vote(schulze.Ballot[string]{
		"A": 1,
		"B": 1,
		"C": 2,
	}); err != nil {
		log.Fatal(err)
	}

	// Calculate the result.
	result, tie := v.Compute()
	if tie {
		log.Fatal("tie")
	}
	fmt.Println("winner:", result[0].Choice)

	// Output: winner: A
}

func ExampleNewPreferences() {
	// Create a new voting.
	choices := []string{"A", "B", "C"}
	preferences := schulze.NewPreferences(len(choices))

	// First vote.
	if err := schulze.Vote(preferences, choices, schulze.Ballot[string]{
		"A": 1,
	}); err != nil {
		log.Fatal(err)
	}

	// Second vote.
	if err := schulze.Vote(preferences, choices, schulze.Ballot[string]{
		"A": 1,
		"B": 1,
		"C": 2,
	}); err != nil {
		log.Fatal(err)
	}

	// Calculate the result.
	result, tie := schulze.Compute(preferences, choices)
	if tie {
		log.Fatal("tie")
	}
	fmt.Println("winner:", result[0].Choice)

	// Output: winner: A
}
