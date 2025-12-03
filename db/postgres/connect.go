package postgres

import (
	"time"

	"github.com/ooqls/getset/log"
	"go.uber.org/zap"
)

var l *zap.Logger = log.NewLogger("postgres")

func Retry(ping func() error) error {
	var err error
	for retry := 0; retry < 3; retry++ { // Fix: change condition from `>` to `<`
		err = ping()
		if err != nil {
			l.Warn("failed to connect, retrying...", zap.Int("retry", retry))
			time.Sleep(time.Second)
		} else {
			l.Info("connected successfully")
			return nil
		}
	}

	if err != nil {
		l.Error("failed to connect after retries", zap.Error(err))
	}

	return err
}
