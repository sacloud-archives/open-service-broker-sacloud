package broker

import (
	"context"
	"fmt"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/sacloud/open-service-broker-sacloud/broker/handler"
)

const defaultPort = 8080

// Broker is interface of broker-api-server
type Broker interface {
	Start(context.Context) error
}

type broker struct {
	config  *Config
	router  http.Handler
	handler func(context.Context) error
}

// NewBroker returns new Broker
func NewBroker(cfg *Config) Broker {
	b := &broker{config: cfg}
	b.handler = b.listenAndServe
	return b
}

// Starts implements Broker.Start
func (b *broker) Start(ctx context.Context) error {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	errChan := make(chan error)

	// Start server
	go func() {
		select {
		case errChan <- b.start(ctx):
		case <-ctx.Done():
		}
	}()

	select {
	case <-ctx.Done():
		log.Debug("context canceled; broker shutting down")
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}

func (b *broker) start(ctx context.Context) error {

	var username, password string
	if b.config != nil {
		username = b.config.BasicAuthUsername
		password = b.config.BasicAuthPassword
	}

	b.router = handler.Router(username, password)
	return b.handler(ctx)
}

func (b *broker) listenAndServe(ctx context.Context) error {
	errChan := make(chan error)
	port := defaultPort
	if b.config != nil && b.config.Port > 0 {
		port = b.config.Port
	}

	s := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: b.router,
	}
	go func() {
		log.WithField(
			"Listen",
			fmt.Sprintf("http://0.0.0.0:%d", port),
		).Info("Service Broker API Server is listening")

		select {
		case errChan <- s.ListenAndServe():
		case <-ctx.Done():
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			time.Second*5,
		)
		defer cancel()
		s.Shutdown(shutdownCtx) // nolint
		return ctx.Err()
	}
}
