/*
 * Copyright (c) 2023-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package pool

// IPool s.e.
// use NewPool() and NewPoolStub()
type IPool[T any] interface {
	Get() T

	// borrows an object from pool which can be released by releaser func only. obj.Release() causes panic.
	// use case: pooled root object owns a nested pooled object. Borrow nested by GetOwned to avoid nested release before root release
	GetOwned(owner IReleaser) T
}

// IReleaser provides ability to return the instance which holds the IReleaser to the pool
// owner instance must set its internal IReleaser to the implementation obtained from the pool
// see NewPool() instantiator argument
type IReleaser interface {
	// Release returns the owner instance to the pool
	// panics if released already avoiding returning the same object to the pool twice
	// calls owner's Cleanup() if exists before returning to pool
	Release()
	IsOwned() bool

	// for internal use
	releaseOwned()
	reset()
	setIsOwned()
	setBorrowStackTrace(stackTrace string)
	init(obj interface{})
	setOwnedTail(interface{})
	getOwnedTail() interface{}
}
