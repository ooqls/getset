package cache

import (
	"context"
	"time"

	"github.com/valkey-io/valkey-go"
)

type valkeyStore struct {
	vc valkey.Client
}

// NewValkeyStore returns a Store backed by vc.
func NewValkeyStore(vc valkey.Client) Store {
	return &valkeyStore{vc: vc}
}

func (s *valkeyStore) Get(ctx context.Context, key string) ([]byte, error) {
	b, err := s.vc.Do(ctx, s.vc.B().Get().Key(key).Build()).AsBytes()
	if valkey.IsValkeyNil(err) {
		return nil, ErrNotFound
	}
	return b, err
}

func (s *valkeyStore) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	set := s.vc.B().Set().Key(key).Value(valkey.BinaryString(value))
	if ttl > 0 {
		return s.vc.Do(ctx, set.Px(ttl).Build()).Error()
	}
	return s.vc.Do(ctx, set.Build()).Error()
}

func (s *valkeyStore) Delete(ctx context.Context, key string) error {
	return s.vc.Do(ctx, s.vc.B().Del().Key(key).Build()).Error()
}

func (s *valkeyStore) DeletePrefix(ctx context.Context, prefix string) error {
	var cursor uint64
	for {
		entry, err := s.vc.Do(ctx, s.vc.B().Scan().Cursor(cursor).Match(globPrefix(prefix)).Count(scanCount).Build()).AsScanEntry()
		if err != nil {
			return err
		}
		if len(entry.Elements) > 0 {
			if err := s.vc.Do(ctx, s.vc.B().Del().Key(entry.Elements...).Build()).Error(); err != nil {
				return err
			}
		}
		if entry.Cursor == 0 {
			return nil
		}
		cursor = entry.Cursor
	}
}
