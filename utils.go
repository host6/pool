/*
 * Copyright (c) 2023-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package pool

import (
	"fmt"
	"io"
	"strings"
)

// GetObjectsInUse returns total amount of objects taken from all pools but not returned
// useful in tests
func GetObjectsInUse() uint64 {
	res := uint64(0)
	m.Lock()
	for _, oc := range objectsCounters {
		res += oc()
	}
	m.Unlock()
	return res
}

// RegisterObjectsInUseCounter registers pooled objects counter which will be considered by GetObjectsInUse()
// called automatically on each NewPool() to track the new pool
// useful if e.g. we have different pool somewhere else it is useful to register its counter here and use pool.GetObjectsInUse() only as a single pooled objects counter
// note: func counter must be thread-safe
func RegisterObjectsInUseCounter(oc func() uint64) {
	m.Lock()
	objectsCounters = append(objectsCounters, oc)
	m.Unlock()
}

// PrintNonReleased prints stacktraces that explains where non-released objects were borrowed
// note: debug mode must be turned on by `pool.SetDebug(true)` call
func PrintNonReleased(w io.Writer) {
	nr := getNonReleased()
	if len(nr) == 0 {
		return
	}
	fmt.Fprintln(w, "objects borrowed from pools but not released:")
	for st, amount := range nr {
		st = "\t" + strings.ReplaceAll(st, "\n", "\n\t")
		st = st[:len(st)-1]
		fmt.Fprintf(w, "%d not released borrowed at:\n%s", amount, st)
	}
}

// SetDebug switches debug mode. In debug mode pool engine tracks amounts of non-released objects
// per each borrow source code point (for all pools)
// use PrintNonReleased() to get explanations
// useful for investigations only, decreases performance
func SetDebug(IsDebug bool) {
	isDebug = IsDebug
}

func getNonReleased() map[string]int {
	m.Lock()
	res := map[string]int{}
	for k, v := range objAmounts {
		if v > 0 {
			res[k] = v
		}
	}
	m.Unlock()
	return res
}
