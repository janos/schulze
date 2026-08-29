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

	_, err := v.Vote(schulze.Ballot[int]{20: 1})
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
	preferences := schulze.NewPreferences[int](len(choices))

	_, err := schulze.Vote(preferences, choices, schulze.Ballot[int]{20: 1})
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

func TestValidateChoices_Errors(t *testing.T) {
	if err := schulze.ValidateChoices([]string{}); !errors.Is(err, schulze.ErrEmptyChoices) {
		t.Fatalf("got error %v, want ErrEmptyChoices", err)
	}

	err := schulze.ValidateChoices([]string{"A", "B", "A"})
	var dupErr *schulze.DuplicateChoiceError[string]
	if !errors.As(err, &dupErr) {
		t.Fatalf("got error %v, want DuplicateChoiceError", err)
	}
	if dupErr.Choice != "A" {
		t.Fatalf("got choice %v, want A", dupErr.Choice)
	}
}

func TestInvalidPreferencesError(t *testing.T) {
	choices := []string{"A", "B"}
	badPrefs := make([]int, 2) // wanted 4

	_, err := schulze.Vote(badPrefs, choices, schulze.Ballot[string]{"A": 1})
	var invErr *schulze.InvalidPreferencesError
	if !errors.As(err, &invErr) {
		t.Fatalf("got error %v, want InvalidPreferencesError", err)
	}
	if invErr.Length != 2 || invErr.Wanted != 4 {
		t.Fatalf("got lengths %d/%d, want 2/4", invErr.Length, invErr.Wanted)
	}
}
