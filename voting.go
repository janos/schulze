// Copyright (c) 2022, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schulze

// Voting holds number of votes for every pair of choices. It is a convenient
// construct to use when the preferences slice does not have to be exposed, and
// should be kept safe from accidental mutation. Methods on the Voting type are
// not safe for concurrent calls.
type Voting[C comparable] struct {
	choices     []C
	preferences []int
}

// NewVoting initializes a new voting state for the provided choices.
func NewVoting[C comparable](choices []C) *Voting[C] {
	return &Voting[C]{
		choices:     choices,
		preferences: NewPreferences(len(choices)),
	}
}

// Vote adds a voting preferences by a single voting ballot. A record of a
// complete and normalized preferences is returned that can be used to unvote.
func (v *Voting[C]) Vote(b Ballot[C]) (Record[C], error) {
	return Vote(v.preferences, v.choices, b)
}

// Unvote removes a voting preferences from a single voting ballot.
func (v *Voting[C]) Unvote(r Record[C]) error {
	return Unvote(v.preferences, v.choices, r)
}

// SetChoices updates the voting accommodate the changes to the choices. It is
// required to pass a complete updated choices.
func (v *Voting[C]) SetChoices(updated []C) {
	v.preferences = SetChoices(v.preferences, v.choices, updated)
}

// Compute calculates a sorted list of choices with the total number of wins for
// each of them. If there are multiple winners, tie boolean parameter is true.
func (v *Voting[C]) Compute() (results []Result[C], duels DuelsIterator[C], tie bool) {
	return Compute(v.preferences, v.choices)
}
