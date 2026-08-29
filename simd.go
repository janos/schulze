// Copyright (c) 2026, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schulze

import (
	"unsafe"
)

// relaxRow applies the inner Floyd-Warshall widest path update on rowJ using rowI:
// rowJ[k] = max(rowJ[k], min(jip, rowI[k])) for all k in [0, n).
func relaxRow[N Number](rowI, rowJ unsafe.Pointer, jip N, n int) {
	if n <= 0 {
		return
	}

	// Fast SIMD path for standard int (int64 on 64-bit platforms)

	if _, ok := any(jip).(int); ok {
		relaxRowInt(rowI, rowJ, int(jip), n)
		return
	}

	// Generic pure Go unrolled fallback for other numeric types
	elemSize := unsafe.Sizeof(N(0))
	const step = 8
	mod := n % step
	end := n - mod

	for k := 0; k < end; k += step {
		setStrengthValue(rowI, rowJ, uintptr(k), jip, elemSize)
		setStrengthValue(rowI, rowJ, uintptr(k+1), jip, elemSize)
		setStrengthValue(rowI, rowJ, uintptr(k+2), jip, elemSize)
		setStrengthValue(rowI, rowJ, uintptr(k+3), jip, elemSize)
		setStrengthValue(rowI, rowJ, uintptr(k+4), jip, elemSize)
		setStrengthValue(rowI, rowJ, uintptr(k+5), jip, elemSize)
		setStrengthValue(rowI, rowJ, uintptr(k+6), jip, elemSize)
		setStrengthValue(rowI, rowJ, uintptr(k+7), jip, elemSize)
	}
	for k := end; k < n; k++ {
		setStrengthValue(rowI, rowJ, uintptr(k), jip, elemSize)
	}
}

func setStrengthValue[N Number](rowI, rowJ unsafe.Pointer, k uintptr, jip N, elemSize uintptr) {
	ikp := (*N)(unsafe.Add(rowI, k*elemSize))
	jkp := (*N)(unsafe.Add(rowJ, k*elemSize))
	m := min(jip, *ikp)
	if m > *jkp {
		*jkp = m
	}
}
