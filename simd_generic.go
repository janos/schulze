// Copyright (c) 2026, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !arm64

package schulze

import "unsafe"

func relaxRowInt(rowI, rowJ unsafe.Pointer, jip int, n int) {
	rowISlice := unsafe.Slice((*int)(rowI), n)
	rowJSlice := unsafe.Slice((*int)(rowJ), n)

	const step = 8
	mod := n % step
	end := n - mod

	for k := 0; k < end; k += step {
		if m := min(jip, rowISlice[k]); m > rowJSlice[k] {
			rowJSlice[k] = m
		}
		if m := min(jip, rowISlice[k+1]); m > rowJSlice[k+1] {
			rowJSlice[k+1] = m
		}
		if m := min(jip, rowISlice[k+2]); m > rowJSlice[k+2] {
			rowJSlice[k+2] = m
		}
		if m := min(jip, rowISlice[k+3]); m > rowJSlice[k+3] {
			rowJSlice[k+3] = m
		}
		if m := min(jip, rowISlice[k+4]); m > rowJSlice[k+4] {
			rowJSlice[k+4] = m
		}
		if m := min(jip, rowISlice[k+5]); m > rowJSlice[k+5] {
			rowJSlice[k+5] = m
		}
		if m := min(jip, rowISlice[k+6]); m > rowJSlice[k+6] {
			rowJSlice[k+6] = m
		}
		if m := min(jip, rowISlice[k+7]); m > rowJSlice[k+7] {
			rowJSlice[k+7] = m
		}
	}

	for k := end; k < n; k++ {
		if m := min(jip, rowISlice[k]); m > rowJSlice[k] {
			rowJSlice[k] = m
		}
	}
}
