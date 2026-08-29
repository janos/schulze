// Copyright (c) 2022, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schulze

type bitSet struct {
	single uint64
	words  []uint64
}

func newBitset(size uint64) bitSet {
	if size <= 64 {
		return bitSet{}
	}
	return bitSet{words: make([]uint64, (size+63)/64)}
}

func (s *bitSet) set(i uint64) {
	if len(s.words) == 0 {
		s.single |= 1 << i
		return
	}
	s.words[i/64] |= 1 << (i % 64)
}

func (s bitSet) isSet(i uint64) bool {
	if len(s.words) == 0 {
		return s.single&(1<<i) != 0
	}
	return s.words[i/64]&(1<<(i%64)) != 0
}
