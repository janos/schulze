// Copyright (c) 2022, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schulze

import (
	"math/rand"
	"testing"
	"time"
)

func TestBitSet(t *testing.T) {
	contains := func(i uint, values []uint) bool {
		for _, v := range values {
			if i == v {
				return true
			}
		}
		return false
	}

	seed := time.Now().UnixNano()
	r := rand.New(rand.NewSource(seed))
	size := uint(r.Intn(12345))
	s := newBitset(size)
	values := make([]uint, 0)
	for i, count := uint(0), uint(r.Intn(100)); i < count; i++ {
		values = append(values, uint(r.Intn(int(size))))
	}
	for _, v := range values {
		s.set(v)
	}
	for i := uint(0); i < size; i++ {
		if contains(i, values) {
			if !s.isSet(i) {
				t.Errorf("expected value %v is not set (seed %v)", i, seed)
			}
		} else {
			if s.isSet(i) {
				t.Errorf("value %v is unexpectedly set (seed %v)", i, seed)
			}
		}
	}
}
