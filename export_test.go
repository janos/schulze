// Copyright (c) 2022, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schulze

// Preferences reruns a copy of preferences for testing purposes.
func (v *Voting[C]) Preferences() []int {
	p := make([]int, len(v.preferences))
	copy(p, v.preferences)
	return p
}
