// Copyright (c) 2026, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build arm64

package schulze

import "unsafe"

//go:noescape
func relaxRowARM64(rowI, rowJ unsafe.Pointer, jip int, n int)

func relaxRowInt(rowI, rowJ unsafe.Pointer, jip int, n int) {
	relaxRowARM64(rowI, rowJ, jip, n)
}
