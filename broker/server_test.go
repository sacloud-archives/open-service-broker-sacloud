package broker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestServerStart(t *testing.T) {
	dummyErr := errors.New("dummy")
	cfg := &Config{
		BasicAuthPassword: "test",
		BasicAuthUsername: "password",
	}

	t.Run("Should block until returns error", func(t *testing.T) {
		b := &broker{
			handler: func(ctx context.Context) error {
				return dummyErr
			},
			config: cfg,
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err := b.Start(ctx)
		assert.Equal(t, dummyErr, err)
	})

	t.Run("Should block until returns nil", func(t *testing.T) {
		b := &broker{
			handler: func(ctx context.Context) error {
				return nil
			},
			config: cfg,
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err := b.Start(ctx)
		assert.Nil(t, err)
	})

	t.Run("Should block until timeout", func(t *testing.T) {
		b := &broker{
			handler: func(ctx context.Context) error {
				<-ctx.Done()
				return ctx.Err()
			},
			config: cfg,
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err := b.Start(ctx)
		assert.Equal(t, ctx.Err(), err)
	})
}
