/*
 * Copyright (c) 2021-present unTill Pro, Ltd.
 */

package pool

import "sync"

type implPool[T any] struct {
	sync.Pool
	isStub       bool
	objectsInUse uint64
	instantiator func(releaser IReleaser) any
}

type implIReleaser[T any] struct {
	obj                  T
	isReleased           bool
	ownerPool            *implPool[T]
	isOwned              bool
	cleanupIntf          interface{ Cleanup() }
	borrowStackTrace     string
	initIntf             interface{ Init() }
	isInitIntfDetermined bool
	ownedTail            interface{}
}

type stackFrame struct {
	fn   string
	file string
	line int
}

type stackTrace []stackFrame
