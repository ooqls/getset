package valkey

import (
	"context"
	"testing"
	"time"

	"github.com/ooqls/getset/db/containers"
	"github.com/ooqls/getset/db/valkey"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	timeout := time.Second * 30
	cont := containers.StartValkey(context.Background())
	defer cont.Stop(context.Background(), &timeout)

	m.Run()
}

func TestConnectValkey(t *testing.T) {
	ctx := context.Background()
	con := valkey.GetConnection(ctx)
	assert.NotNilf(t, con, "valkey client should not be nil")
}
