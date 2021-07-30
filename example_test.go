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

func ExampleElection() {
	// Create a new election.
	e := schulze.NewElection("A", "B", "C", "D", "E")

	// First vote.
	if err := e.Vote(schulze.Ballot{
		"A": 1,
	}); err != nil {
		log.Fatal(err)
	}

	// Second vote.
	if err := e.Vote(schulze.Ballot{
		"A": 1,
		"B": 1,
		"D": 2,
	}); err != nil {
		log.Fatal(err)
	}

	// Calculate the result.
	result, tie := e.Compute()
	if tie {
		log.Fatal("tie")
	}
	fmt.Println("winner:", result[0].Choice)

	// Output: winner: A
}
