// Copyright (c) 2022, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schulze

type bitSet []uint64

func newBitset(size uint) bitSet {
	return bitSet(make([]uint64, size/64+1))
}

func (s bitSet) set(i uint) {
	s[i/64] |= 1 << (i % 64)
}

func (s bitSet) isSet(i uint) bool {
	return s[i/64]&(1<<(i%64)) != 0
}
