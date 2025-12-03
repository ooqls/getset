package log

import (
	"testing"

	"github.com/ooqls/getset/log/api/v1/gen"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestSetForm(t *testing.T) {
	l := NewLogger("test")
	l.Info("test")

	err := SetFormat(gen.JSON)
	l = NewLogger("test")
	assert.Nilf(t, err, "SetFormat should not return an error")
	l.Info("test", zap.String("format", "text"))
	l.Debug("hello world")

}
