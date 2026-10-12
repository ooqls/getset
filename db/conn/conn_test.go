package conn

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeClient is the client type for a fake adapter.
type fakeClient struct {
	id     int
	closed atomic.Bool
}

type fakeBackend struct {
	opens atomic.Int32
	fail  atomic.Bool
}

func (b *fakeBackend) open() (*fakeClient, error) {
	if b.fail.Load() {
		return nil, errors.New("backend down")
	}
	return &fakeClient{id: int(b.opens.Add(1))}, nil
}

func closeFake(c *fakeClient) error {
	c.closed.Store(true)
	return nil
}

func newFake() (*Conn[*fakeClient], *fakeBackend) {
	b := &fakeBackend{}
	return New("fake", b.open, closeFake), b
}

func TestGet_OpensDefaultOnceAcrossGoroutines(t *testing.T) {
	c, b := newFake()

	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			_, err := c.Get()
			assert.NoError(t, err)
		})
	}
	wg.Wait()

	assert.Equal(t, int32(1), b.opens.Load())
}

func TestGet_ReturnsSameClientUntilReplaced(t *testing.T) {
	c, _ := newFake()
	first := c.MustGet()
	assert.Same(t, first, c.MustGet())
}

func TestInit_ReplacesAndClosesOpenClient(t *testing.T) {
	c, _ := newFake()
	old := c.MustGet()

	replacement := &fakeClient{id: 99}
	require.NoError(t, c.Init(func() (*fakeClient, error) { return replacement, nil }))

	assert.Same(t, replacement, c.MustGet())
	assert.True(t, old.closed.Load(), "the replaced client should be closed")
}

func TestInit_FailureKeepsOpenClient(t *testing.T) {
	c, _ := newFake()
	old := c.MustGet()

	err := c.Init(func() (*fakeClient, error) { return nil, errors.New("bad options") })

	assert.ErrorContains(t, err, "bad options")
	assert.Same(t, old, c.MustGet())
	assert.False(t, old.closed.Load())
}

func TestInitDefault_IsNoOpWhenOpen(t *testing.T) {
	c, b := newFake()
	replacement := &fakeClient{id: 99}
	require.NoError(t, c.Init(func() (*fakeClient, error) { return replacement, nil }))

	require.NoError(t, c.InitDefault())

	assert.Same(t, replacement, c.MustGet(), "InitDefault must not replace options given to Init")
	assert.Equal(t, int32(0), b.opens.Load())
}

func TestGet_RetriesDefaultAfterFailure(t *testing.T) {
	c, b := newFake()
	b.fail.Store(true)

	_, err := c.Get()
	assert.ErrorContains(t, err, "backend down")
	assert.PanicsWithError(t, "fake: backend down", func() { c.MustGet() })

	b.fail.Store(false)
	client, err := c.Get()
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestClose_ClosesAndNextGetReopens(t *testing.T) {
	c, b := newFake()
	first := c.MustGet()

	require.NoError(t, c.Close())
	assert.True(t, first.closed.Load())

	second := c.MustGet()
	assert.NotSame(t, first, second)
	assert.Equal(t, int32(2), b.opens.Load())
}

func TestClose_WhenNothingOpenIsNoOp(t *testing.T) {
	c, _ := newFake()
	assert.NoError(t, c.Close())
}

func TestDefaultOpenerPanicDoesNotDeadlock(t *testing.T) {
	c := New("fake", func() (*fakeClient, error) { panic("no registry") }, closeFake)

	assert.Panics(t, func() { c.Get() })
	assert.Panics(t, func() { c.Get() }, "a second Get should panic again, not block on the lock")
}
