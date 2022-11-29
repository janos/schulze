// Copyright (c) 2022, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schulze_test

import (
	"errors"
	"strings"
	"testing"

	"resenje.org/schulze"
)

func TestVoting_Vote_UnknownChoiceError(t *testing.T) {
	v := schulze.NewVoting([]int{0, 2, 5, 7})

	err := v.Vote(schulze.Ballot[int]{20: 1})
	var verr *schulze.UnknownChoiceError[int]
	if !errors.As(err, &verr) {
		t.Fatalf("got error %v, want UnknownChoiceError", err)
	}
	if verr.Choice != 20 {
		t.Fatalf("got unknown choice error choice %v, want %v", verr.Choice, 20)
	}
	if !strings.Contains(verr.Error(), "20") {
		t.Fatal("choice index not found in error string")
	}
}

func TestVote_UnknownChoiceError(t *testing.T) {
	choices := []int{0, 2, 5, 7}
	preferences := schulze.NewPreferences(len(choices))

	err := schulze.Vote(choices, preferences, schulze.Ballot[int]{20: 1})
	var verr *schulze.UnknownChoiceError[int]
	if !errors.As(err, &verr) {
		t.Fatalf("got error %v, want UnknownChoiceError", err)
	}
	if verr.Choice != 20 {
		t.Fatalf("got unknown choice error choice %v, want %v", verr.Choice, 20)
	}
	if !strings.Contains(verr.Error(), "20") {
		t.Fatal("choice index not found in error string")
	}
}
