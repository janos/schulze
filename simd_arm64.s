// Copyright (c) 2026, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build arm64

#include "textflag.h"

// func relaxRowARM64(rowI, rowJ unsafe.Pointer, jip int, n int)
TEXT ·relaxRowARM64(SB), NOSPLIT, $0-32
	MOVD	rowI+0(FP), R0
	MOVD	rowJ+8(FP), R1
	MOVD	jip+16(FP), R2
	MOVD	n+24(FP), R3

	CMP	$0, R3
	BLE	done

	// Broadcast jip (R2) into V0.D2
	FMOVD	R2, F0
	VDUP	V0.D[0], V0.D2

loop4:
	CMP	$4, R3
	BLT	loop2

	VLD1.P	32(R0), [V1.D2, V2.D2]
	VLD1	(R1), [V5.D2, V6.D2]

	// 1. min_val for 4 elements
	VCMGT	V1.D2, V0.D2, V16.D2
	VAND	V1.B16, V16.B16, V20.B16
	VBIC	V16.B16, V0.B16, V21.B16
	VORR	V20.B16, V21.B16, V16.B16 // V16 = min(V0, V1)

	VCMGT	V2.D2, V0.D2, V17.D2
	VAND	V2.B16, V17.B16, V20.B16
	VBIC	V17.B16, V0.B16, V21.B16
	VORR	V20.B16, V21.B16, V17.B16 // V17 = min(V0, V2)

	// 2. max(rowJ, min_val) for 4 elements
	VCMGT	V5.D2, V16.D2, V20.D2
	VAND	V16.B16, V20.B16, V22.B16
	VBIC	V20.B16, V5.B16, V23.B16
	VORR	V22.B16, V23.B16, V1.B16 // V1 = max(V5, V16)

	VCMGT	V6.D2, V17.D2, V20.D2
	VAND	V17.B16, V20.B16, V22.B16
	VBIC	V20.B16, V6.B16, V23.B16
	VORR	V22.B16, V23.B16, V2.B16 // V2 = max(V6, V17)

	VST1.P	[V1.D2, V2.D2], 32(R1)

	SUB	$4, R3
	B	loop4

loop2:
	CMP	$2, R3
	BLT	tail

	VLD1.P	16(R0), [V1.D2]
	VLD1	(R1), [V2.D2]

	// min_val = min(V0, V1)
	VCMGT	V1.D2, V0.D2, V3.D2
	VAND	V1.B16, V3.B16, V4.B16
	VBIC	V3.B16, V0.B16, V5.B16
	VORR	V4.B16, V5.B16, V6.B16 // V6 = min_val

	// new_rowJ = max(V2, V6)
	VCMGT	V2.D2, V6.D2, V7.D2
	VAND	V6.B16, V7.B16, V8.B16
	VBIC	V7.B16, V2.B16, V9.B16
	VORR	V8.B16, V9.B16, V10.B16 // V10 = max(V2, V6)

	VST1.P	[V10.D2], 16(R1)

	SUB	$2, R3
	B	loop2

tail:
	CMP	$0, R3
	BLE	done

	// Scalar tail for last 1 element
	MOVD.P	8(R0), R4
	MOVD	(R1), R5

	CMP	R2, R4
	CSEL	LT, R4, R2, R6 // R6 = min(R2, R4)

	CMP	R5, R6
	CSEL	GT, R6, R5, R7 // R7 = max(R5, R6)
	MOVD.P	R7, 8(R1)

done:
	RET
