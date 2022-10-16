// Copyright (c) 2021, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schulze

import (
	"fmt"
)

// UnknownChoiceError represent an error in case that a choice that is not in
// the voting is used.
type UnknownChoiceError[C comparable] struct {
	Choice C
}

func (e *UnknownChoiceError[C]) Error() string {
	return fmt.Sprintf("unknown choice %v", e.Choice)
}
