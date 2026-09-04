// Copyright (c) 2022, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schulze

import (
	"iter"
	"slices"
)

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
		choices:     slices.Clone(choices),
		preferences: NewPreferences[int](len(choices)),
	}
}

// Choices returns a copy of the current choices.
func (v *Voting[C]) Choices() []C {
	return slices.Clone(v.choices)
}

// Preferences returns a copy of the preferences slice.
func (v *Voting[C]) Preferences() []int {
	return slices.Clone(v.preferences)
}

// VoteWeighted adds voting preferences by a single voting ballot scaled by the specified weight.
func (v *Voting[C]) VoteWeighted(b Ballot[C], weight int) (Record[C], error) {
	return VoteWeighted(v.preferences, v.choices, b, weight)
}

// Vote adds voting preferences by a single voting ballot. A record of a
// complete and normalized preferences is returned that can be used to unvote.
func (v *Voting[C]) Vote(b Ballot[C]) (Record[C], error) {
	return Vote(v.preferences, v.choices, b)
}

// VoteRankedWeighted adds voting preferences from a pre-ranked slice of choices scaled by the specified weight.
func (v *Voting[C]) VoteRankedWeighted(rankedChoices [][]C, weight int) (Record[C], error) {
	return VoteRankedWeighted(v.preferences, v.choices, rankedChoices, weight)
}

// VoteRanked adds voting preferences from a pre-ranked slice of choices.
func (v *Voting[C]) VoteRanked(rankedChoices [][]C) (Record[C], error) {
	return VoteRanked(v.preferences, v.choices, rankedChoices)
}

// VoteFrom allows voting directly from an arbitrary data structure using a ranking function.
func (v *Voting[C]) VoteFrom[T any](item T, ranker func(item T, choice C) (rank int, ranked bool)) (Record[C], error) {
	b := make(Ballot[C], len(v.choices))
	for _, c := range v.choices {
		if rank, ok := ranker(item, c); ok {
			b[c] = rank
		}
	}
	return v.Vote(b)
}

// UnvoteWeighted removes voting preferences from a single voting ballot record scaled by the specified weight.
func (v *Voting[C]) UnvoteWeighted(r Record[C], weight int) error {
	return UnvoteWeighted(v.preferences, v.choices, r, weight)
}

// Unvote removes voting preferences from a single voting ballot record.
func (v *Voting[C]) Unvote(r Record[C]) error {
	return Unvote(v.preferences, v.choices, r)
}

// SetChoices updates the voting to accommodate changes to the choices. It is
// required to pass complete updated choices.
func (v *Voting[C]) SetChoices(updated []C) {
	v.preferences = SetChoices(v.preferences, v.choices, updated)
	v.choices = slices.Clone(updated)
}

// Winner calculates the election winner. It first checks for a Condorcet
// winner in O(N^2) time; if none exists, it computes the Schulze beatpaths.
// If multiple choices tie for the winning position or there are no choices,
// tie is true.
func (v *Voting[C]) Winner() (winner Result[C], tie bool) {
	return Winner(v.preferences, v.choices)
}

// Compute calculates a sorted list of choices with the total number of wins for
// each of them. If there are multiple winners, tie boolean parameter is true.
func (v *Voting[C]) Compute() (results []Result[C], duels DuelsIterator[C], tie bool) {
	return Compute(v.preferences, v.choices)
}

// Duels returns an iter.Seq[*Duel[C]] standard iterator over pairwise duel outcomes.
func (v *Voting[C]) Duels() iter.Seq[*Duel[C]] {
	return Duels(v.preferences, v.choices)
}

// Map creates a new Voting instance with transformed choices.
func (v *Voting[C]) Map[NewC comparable](fn func(C) NewC) *Voting[NewC] {
	newChoices := make([]NewC, len(v.choices))
	for i, c := range v.choices {
		newChoices[i] = fn(c)
	}
	newPreferences := slices.Clone(v.preferences)
	return &Voting[NewC]{
		choices:     newChoices,
		preferences: newPreferences,
	}
}

// ExportPreferences exports the preferences slice converted into the requested Number type.
func (v *Voting[C]) ExportPreferences[N Number]() []N {
	res := make([]N, len(v.preferences))
	for i, p := range v.preferences {
		res[i] = N(p)
	}
	return res
}

// ImportPreferences updates internal preferences by converting from the provided Number slice.
func (v *Voting[C]) ImportPreferences[N Number](prefs []N) error {
	if len(prefs) != len(v.preferences) {
		return &InvalidPreferencesError{
			Length: len(prefs),
			Wanted: len(v.preferences),
		}
	}
	for i, p := range prefs {
		v.preferences[i] = int(p)
	}
	return nil
}
