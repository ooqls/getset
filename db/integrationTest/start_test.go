package integrationtest

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/ooqls/getset/db/containers"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
)

func TestStart(t *testing.T) {
	startFuncs := []func(ctx context.Context, opts ...containers.Options) testcontainers.Container{
		containers.StartElasticsearch,
		containers.StartPostgres,
		containers.StartRedis,
		containers.StartValkey,
	}

	ctx := context.Background()
	wg := sync.WaitGroup{}
	conts := []testcontainers.Container{}
	wg.Add(len(startFuncs))

	for _, sf := range startFuncs {
		go func() {
			c := sf(ctx)
			conts = append(conts, c)
			wg.Done()
		}()
	}

	wg.Wait()
	wg.Add(len(conts))

	for _, c := range conts {
		go func(stopC testcontainers.Container) {
			defer wg.Done()
			timeout := time.Second * 5
			err := stopC.Stop(ctx, &timeout)
			assert.Nilf(t, err, "should not have gotten an error when stopping container")
		}(c)
	}

	wg.Wait()
}
