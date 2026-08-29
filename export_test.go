// Copyright (c) 2022, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schulze

import "unsafe"

func RelaxRow(rowI, rowJ unsafe.Pointer, jip int, n int) {
	relaxRow(rowI, rowJ, jip, n)
}
