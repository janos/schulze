// Copyright (c) 2026, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build amd64

#include "textflag.h"

// func relaxRowAMD64(rowI, rowJ unsafe.Pointer, jip int, n int)
TEXT ·relaxRowAMD64(SB), NOSPLIT, $0-32
	MOVQ	rowI+0(FP), AX
	MOVQ	rowJ+8(FP), BX
	MOVQ	jip+16(FP), CX
	MOVQ	n+24(FP), DI

	CMPQ	DI, $0
	JLE	done

	// Broadcast jip (CX) into X0
	MOVQ	CX, X0
	PUNPCKLQDQ X0, X0

loop2:
	CMPQ	DI, $2
	JL	tail

	// Load 2 elements from rowI into X1, rowJ into X2
	MOVOU	(AX), X1
	MOVOU	(BX), X2

	// X3 = min(X0, X1)
	MOVOU	X0, X3
	PCMPGTQ	X1, X3
	MOVOU	X1, X4
	PAND	X3, X4
	PANDN	X0, X3
	POR	X4, X3 // X3 is now min(X0, X1)

	// X2 = max(X2, X3)
	MOVOU	X3, X5
	PCMPGTQ	X2, X5
	MOVOU	X3, X6
	PAND	X5, X6
	PANDN	X2, X5
	POR	X6, X5

	// Store back to rowJ
	MOVOU	X5, (BX)

	ADDQ	$16, AX
	ADDQ	$16, BX
	SUBQ	$2, DI
	JMP	loop2

tail:
	CMPQ	DI, $0
	JLE	done

	MOVQ	(AX), R8
	MOVQ	(BX), R9

	// m = min(jip, ikv)
	CMPQ	R8, CX
	MOVQ	CX, R10
	CMOVQLT	R8, R10

	// max(jkv, m)
	CMPQ	R9, R10
	CMOVQGT	R10, R9
	MOVQ	R9, (BX)

done:
	RET
