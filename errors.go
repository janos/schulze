// Copyright (c) 2021, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schulze

import (
	"errors"
	"fmt"
)

// ErrEmptyChoices indicates that an operation was attempted with an empty choices slice.
var ErrEmptyChoices = errors.New("schulze: choices cannot be empty")

// ErrInvalidWeight indicates that a voting operation was attempted with non-positive weight.
var ErrInvalidWeight = errors.New("schulze: weight must be greater than zero")

// UnknownChoiceError indicates that a choice in a ballot is not among known choices.
type UnknownChoiceError[C comparable] struct {
	Choice C
}

func (e *UnknownChoiceError[C]) Error() string {
	return fmt.Sprintf("schulze: unknown choice %v", e.Choice)
}

// DuplicateChoiceError indicates that choices list contains duplicates.
type DuplicateChoiceError[C comparable] struct {
	Choice C
}

func (e *DuplicateChoiceError[C]) Error() string {
	return fmt.Sprintf("schulze: duplicate choice %v", e.Choice)
}

// InvalidPreferencesError indicates that the preferences slice length does not match choicesCount*choicesCount.
type InvalidPreferencesError struct {
	Length int
	Wanted int
}

func (e *InvalidPreferencesError) Error() string {
	return fmt.Sprintf("schulze: invalid preferences length %d, wanted %d", e.Length, e.Wanted)
}
