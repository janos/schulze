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
	v, err := schulze.NewVoting("A", "B", "C")
	if err != nil {
		log.Fatal(err)
	}

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
	result, tie := v.Results()
	if tie {
		log.Fatal("tie")
	}
	fmt.Println("winner:", result[0].Choice)

	// Output: winner: A
}
