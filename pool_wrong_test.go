/*
 * Copyright (c) 2023-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package pool

import (
	"log"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/valyala/bytebufferpool"
	"golang.org/x/exp/slices"
)

// example struct to be pooled
type pooled_wrong struct {
	b *bytebufferpool.ByteBuffer
}

var pool = sync.Pool{
	New: func() interface{} {
		return &pooled_wrong{}
	},
}

func (p *pooled_wrong) Release() {
	// problem: called twice -> put same pointer to the pool twice -> get same pointer from the pool twice
	// (solution: introduce `isReleased` flag, check it and set in Release())
	// problem: should write such boilerplate per each pooled struct
	bytebufferpool.Put(p.b)
	pool.Put(p)
}

func GetPooledStruct() *pooled_wrong {
	// problem: how to track borrowed but not returned?
	// see TestPool_WrongExample()
	res := pool.Get().(*pooled_wrong)
	res.b = bytebufferpool.Get()
	return res
}

func TestPool_WrongExample(t *testing.T) {
	log.Println(os.Args)
	if slices.Contains(os.Args, "-race") {
		t.Skip("problem does not appear in -race mode")
	}
	wrong := GetPooledStruct()
	wrong.Release()
	wrong.Release()

	new1 := GetPooledStruct()
	new2 := GetPooledStruct()
	// log.Println(new1)
	// log.Println(new2)

	// problem: 2 objects taken from the sync pool are actually the same object -> sync pool is not a sync pool anymore
	// problem: there is no error here. Errors will appear as memory damages in random parts of the application
	require.True(t, new1 == new2)
}
