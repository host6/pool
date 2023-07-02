/*
 * Copyright (c) 2020-present unTill Pro, Ltd.
 */

package pool

import (
	"bytes"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
)

var (
	m               sync.Mutex = sync.Mutex{}
	objectsCounters []func() uint64
	isDebug         bool
	objAmounts      map[string]int = map[string]int{}
)

func (st stackTrace) string() string {
	buf := bytes.NewBufferString("")
	for _, sf := range st {
		buf.WriteString(fmt.Sprintf("%s\n\t%s:%d\n", sf.fn, sf.file, sf.line))
	}
	return buf.String()
}

func (p *implPool[T]) Get() T {
	obj := p.get()
	releaseable := obj.(IReleaser)
	releaseable.reset()
	releaseable.init(obj)
	atomic.AddUint64(&p.objectsInUse, 1)
	if isDebug {
		st := getStackTrace().string()
		releaseable.setBorrowStackTrace(st)
		m.Lock()
		count := objAmounts[st]
		count++
		objAmounts[st] = count
		m.Unlock()
	}
	return obj.(T)
}

func (p *implPool[T]) get() any {
	var obj any
	if p.isStub {
		releaser := &implIReleaser[T]{
			ownerPool: p,
		}
		obj = p.instantiator(releaser)
		releaser.cleanupIntf, _ = obj.(interface{ Cleanup() })
		releaser.obj = obj.(T)
	} else {
		obj = p.Pool.Get()
	}
	return obj
}

func (p *implPool[T]) GetOwned(owner IReleaser) T {
	obj := p.get()
	atomic.AddUint64(&p.objectsInUse, 1)
	releaseable := obj.(IReleaser)
	releaseable.reset()
	releaseable.setIsOwned()
	releaseable.setOwnedTail(owner.getOwnedTail())
	owner.setOwnedTail(releaseable)
	releaseable.init(obj)
	if isDebug {
		st := getStackTrace().string()
		releaseable.setBorrowStackTrace(st)
		m.Lock()
		count := objAmounts[st]
		count++
		objAmounts[st] = count
		m.Unlock()
	}
	return obj.(T)
}

func (p *implPool[T]) GetObjectsInUse() uint64 {
	return atomic.LoadUint64(&p.objectsInUse)
}

func (r *implIReleaser[T]) Release() {
	if r.isOwned {
		panic("must be released by owner")
	}
	r.releaseOwned()
}

func (r *implIReleaser[T]) reset() {
	r.isReleased = false
	r.isOwned = false
}

func (r *implIReleaser[T]) IsOwned() bool {
	return r.isOwned
}

func (r *implIReleaser[T]) setIsOwned() {
	r.isOwned = true
}

func (r *implIReleaser[T]) setBorrowStackTrace(stackTrace string) {
	r.borrowStackTrace = stackTrace
}

func (r *implIReleaser[T]) releaseOwned() {
	if r.isReleased {
		panic("already released")
	}
	if r.cleanupIntf != nil {
		r.cleanupIntf.Cleanup()
	}
	if r.ownedTail != nil {
		r.ownedTail.(IReleaser).releaseOwned()
		r.ownedTail = nil
	}
	r.isReleased = true
	atomic.AddUint64(&r.ownerPool.objectsInUse, ^uint64(0))
	if isDebug {
		m.Lock()
		amount := objAmounts[r.borrowStackTrace]
		objAmounts[r.borrowStackTrace] = amount - 1
		m.Unlock()
	}
	if !r.ownerPool.isStub {
		r.ownerPool.Put(r.obj)
	}
}

func (r *implIReleaser[T]) init(obj interface{}) {
	if !r.isInitIntfDetermined {
		r.initIntf, _ = obj.(interface{ Init() })
		r.isInitIntfDetermined = true
	}
	if r.initIntf != nil {
		r.initIntf.Init()
	}
}

func (r *implIReleaser[T]) setOwnedTail(tail interface{}) {
	r.ownedTail = tail
}

func (r *implIReleaser[T]) getOwnedTail() interface{} {
	return r.ownedTail
}

// NewPoolStub creates pool which does not act as a pool. I.e. just creates a new instance on each Get()
// Release() does nothing more but Cleaunp() call if it exists
// does not track borrow source code points in debug mode
// useful for investigations
func NewPoolStub[T any](instantiator func(releaser IReleaser) any) IPool[T] {
	res := newPool[T](instantiator)
	res.instantiator = instantiator
	res.isStub = true
	return res
}

func NewPool[T any](instantiator func(releaser IReleaser) any) IPool[T] {
	res := newPool[T](nil)
	res.Pool = sync.Pool{
		New: func() interface{} {
			releaser := &implIReleaser[T]{
				ownerPool: res,
			}
			newInstance := instantiator(releaser)
			releaser.cleanupIntf, _ = newInstance.(interface{ Cleanup() })
			releaser.obj = newInstance.(T)
			return newInstance
		},
	}
	return res
}

func newPool[T any](instantiator func(releaser IReleaser) any) *implPool[T] {
	res := &implPool[T]{instantiator: instantiator}
	RegisterObjectsInUseCounter(func() uint64 { return res.GetObjectsInUse() })
	return res
}

func getStackTrace() stackTrace {
	pc := make([]uintptr, 100) // can't estimate
	n := runtime.Callers(3, pc)
	frames := runtime.CallersFrames(pc[:n])
	st := stackTrace{}
	for {
		frame, more := frames.Next()
		st = append(st, stackFrame{
			fn:   frame.Function,
			file: frame.File,
			line: frame.Line,
		})
		if !more {
			break
		}
	}
	return st
}
