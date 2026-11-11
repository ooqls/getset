package initv1_test

import (
	"context"
	"testing"

	"github.com/ooqls/getset/eventsource/eventsourcingv1"
	"github.com/ooqls/getset/eventsource/initv1"
	"github.com/stretchr/testify/require"
)

func TestInit_Empty(t *testing.T) {
	t.Parallel()
	require.NoError(t, initv1.Init(t.Context()))
}

func TestInit_Canceled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	require.ErrorIs(t, initv1.Init(ctx, eventsourcingv1.EventSource("test")), context.Canceled)
}
