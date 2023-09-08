/*
 * Copyright (c) 2023-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
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
