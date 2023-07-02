package pool

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/valyala/bytebufferpool"
)

type owner struct {
	IReleaser
	nested *nested
	bb     *bytebufferpool.ByteBuffer
}

type nested struct {
	IReleaser
	internal *internal
	bb       *bytebufferpool.ByteBuffer
}

type internal struct {
	IReleaser
}

func (n *nested) Init() {
	n.internal = poolInternal.GetOwned(n)
	n.bb = bytebufferpool.Get()
}

func (n *nested) Cleanup() {
	bytebufferpool.Put(n.bb)
	n.bb = nil
}

func (o *owner) Init() {
	// lifetime of owner.nested must not be shorter than owner's so let's obtain *nested by GetOwned()
	// so that it will be released automatically on owner.Release()
	o.nested = poolNested.GetOwned(o)
	o.bb = bytebufferpool.Get()
}

func (o *owner) Cleanup() {
	bytebufferpool.Put(o.bb)
	o.bb = nil
}

var (
	poolOwner = NewPool[*owner](func(releaser IReleaser) any {
		return &owner{IReleaser: releaser}
	})
	poolNested = NewPool[*nested](func(releaser IReleaser) any {
		return &nested{IReleaser: releaser}
	})
	poolInternal = NewPool[*internal](func(releaser IReleaser) any {
		return &internal{IReleaser: releaser}
	})
)

func TestBasicUsage_Owned(t *testing.T) {
	require := require.New(t)

	owner := poolOwner.Get()

	require.Equal(uint64(3), GetObjectsInUse())
	require.NotNil(owner.bb)
	require.NotNil(owner.nested.bb)
	require.NotNil(owner.nested.internal)

	// unale to release nested got by GetOwned()
	// its lifetime is not under developer's control
	// it will be released automatically on owner.Release() to prevent the released owner.nested usage
	require.Panics(func() { owner.nested.Release() })

	owner.Release()
	// owner is released, nested is released automatically as well
	// neither owner, nor any of its fields, nor owner.nested itself and its fields must not be touched from now on

	require.Equal(uint64(0), GetObjectsInUse())
}

func BenchmarkOwned(b *testing.B) {
	b.Run("basic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			owner := poolOwner.Get()
			owner.Release()
		}
	})
}
