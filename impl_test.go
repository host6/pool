package pool

import (
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/valyala/bytebufferpool"
)

type myStruct struct {
	// each pooled struct must include IReleaser field that provides Release() ability.
	// this field will initialized in the instantiator
	IReleaser

	// exmaple nested field that requires personal handling (e.g. borrow\release)
	bb   *bytebufferpool.ByteBuffer
	fld1 int
}

// optional Clenaup() will be called automatically right before returning the myStruct instance to the pool
func (ms *myStruct) Cleanup() {
	bytebufferpool.Put(ms.bb)
	ms.bb = nil
}

// optional Init() will be called automatically on each myStruct instance borrow. It should init the current instance
func (ms *myStruct) Init() {
	ms.bb = bytebufferpool.Get()
}

func TestBasicUsage_Simple(t *testing.T) {
	require := require.New(t)
	p := NewPool[*myStruct](func(releaser IReleaser) any {
		// instantiator must manually initialize IReleaser field with the provided implementation
		return &myStruct{IReleaser: releaser}
	})

	myStructInstance := p.Get()

	// internal initialization is done in myStruct.Init()
	require.NotNil(myStructInstance.bb)

	// 1 object in use
	require.Equal(uint64(1), GetObjectsInUse())

	// return the instance back to the pool
	myStructInstance.Release()
	// myStruct.bb is returned back to `bytebufferpool` by myStruct.Cleanup()
	// myStructInstance is returned to the pool
	// myStructInstance as well as its any member must not be used (even touched) from now on

	// unable to return the same object to the pool twice
	require.Panics(func() { myStructInstance.Release() })

	// no objects in use
	require.Zero(GetObjectsInUse())
}

func TestObjectsUsageTrackInDebugMode(t *testing.T) {
	require := require.New(t)
	SetDebug(true)
	defer SetDebug(false)
	p := NewPool[*myStruct](func(releaser IReleaser) any {
		return &myStruct{IReleaser: releaser}
	})

	roots := []*myStruct{}
	for i := 0; i < 10; i++ {
		roots = append(roots, p.Get())
	}

	// one more as an example
	roots = append(roots, p.Get())

	// release one as an example
	roots[5].Release()

	// prints code points where objects were borrowed but not released
	PrintNonReleased(os.Stdout)

	for i, root := range roots {
		if i != 5 {
			root.Release()
		}
	}

	// prints nothing
	PrintNonReleased(os.Stdout)

	require.Zero(GetObjectsInUse())
}

func TestStub(t *testing.T) {
	require := require.New(t)
	poolOwner := NewPoolStub[*owner](func(releaser IReleaser) any {
		return &owner{
			IReleaser: releaser,
		}
	})
	poolNested = NewPoolStub[*nested](func(releaser IReleaser) any {
		return &nested{
			IReleaser: releaser,
		}
	})

	// borrow pooled struct, initialize fields
	owner := poolOwner.Get()
	require.Equal(uint64(3), GetObjectsInUse())

	// owned struct can not be accidentally released
	require.Panics(func() { owner.nested.Release() })

	// owner will release its internal fields and an owned struct using its special releaser
	// after that `owner` struct itself will be returned to the pool engine
	owner.Release()

	// unable to release twice in stub mode as well to avoid cleanup() unexpected execution
	require.Panics(func() { owner.Release() })

	require.Zero(GetObjectsInUse())
}

func TestStress(t *testing.T) {
	p := NewPool[*myStruct](func(releaser IReleaser) any { return &myStruct{IReleaser: releaser} })
	ch := make(chan *myStruct)
	nch := make(chan int, 1000)
	for i := 0; i < 1000; i++ {
		go func(i int) {
			ts1 := p.Get()
			ts1.fld1 = i
			ch <- ts1
		}(i)
		go func() {
			obj := <-ch
			n := obj.fld1
			obj.Release()
			nch <- n
		}()
	}

	numbers := map[int]struct{}{}
	for i := 0; i < 1000; i++ {
		n := <-nch
		require.Less(t, n, 1000, n)
		if _, exists := numbers[n]; exists {
			t.Fatal()
		}
		numbers[n] = struct{}{}
	}
	require.Zero(t, GetObjectsInUse())
}

func BenchmarkBasic(b *testing.B) {
	p := NewPool[*myStruct](func(releaser IReleaser) any {
		return &myStruct{IReleaser: releaser}
	})

	b.Run("basic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			myStructInstance := p.Get()
			myStructInstance.Release()
		}
	})
}

type simpleStruct struct {
	IReleaser
	isReleased bool
}

// BenchmarkExample/pool-4        20090349	        71.98 ns/op	       0 B/op	       0 allocs/op
// BenchmarkExample/sync.Pool-4   39997732	        27.70 ns/op	       0 B/op	       0 allocs/op
func BenchmarkExample(b *testing.B) {
	pool := NewPool[*simpleStruct](func(releaser IReleaser) any { return &simpleStruct{IReleaser: releaser} })

	b.Run("pool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ps := pool.Get()
			ps.Release()
		}
	})

	b.Run("sync.Pool", func(b *testing.B) {
		syncPool := sync.Pool{
			New: func() interface{} {
				return &simpleStruct{}
			},
		}
		objectsInUse := uint64(0)

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			obj := syncPool.Get().(*simpleStruct)
			atomic.AddUint64(&objectsInUse, uint64(1))
			obj.isReleased = false

			if obj.isReleased {
				panic("already released")
			}
			syncPool.Put(obj)
			atomic.AddUint64(&objectsInUse, ^uint64(0))
		}
	})

	require.Zero(b, GetObjectsInUse())
}
