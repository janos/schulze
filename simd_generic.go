// Copyright (c) 2026, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !arm64 && !amd64

package schulze

import "unsafe"

func relaxRowInt(rowI, rowJ unsafe.Pointer, jip int, n int) {
	for k := 0; k < n; k++ {
		ikp := (*int)(unsafe.Add(rowI, uintptr(k)*8))
		jkp := (*int)(unsafe.Add(rowJ, uintptr(k)*8))
		m := min(jip, *ikp)
		if m > *jkp {
			*jkp = m
		}
	}
}
