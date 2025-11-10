// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package startup

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/internal/orchestrator/subscriber"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// EventSubscriberInitializer initializes event subscriber.
type EventSubscriberInitializer struct {
	logger       core.Logger
	natsInit     *pkginitializers.NATSInitializer
	coreServices *CoreServicesInitializer

	subscriber *subscriber.Subscriber
}

// NewEventSubscriberInitializer creates a new event subscriber initializer.
func NewEventSubscriberInitializer(
	logger core.Logger,
	natsInit *pkginitializers.NATSInitializer,
	coreServices *CoreServicesInitializer,
) *EventSubscriberInitializer {
	return &EventSubscriberInitializer{
		logger:       logger,
		natsInit:     natsInit,
		coreServices: coreServices,
	}
}

// Name returns the initializer name.
func (e *EventSubscriberInitializer) Name() string {
	return "event-subscriber"
}

// Priority returns initialization priority.
func (e *EventSubscriberInitializer) Priority() int {
	return 700 // After core services (600)
}

// Initialize starts the event subscriber.
func (e *EventSubscriberInitializer) Initialize(ctx context.Context) error {
	e.logger.Info("Initializing event subscriber")

	// Verify dependencies
	natsConn := e.natsInit.Conn()
	if natsConn == nil {
		return fmt.Errorf("NATS connection not initialized")
	}

	services := e.coreServices.Services()
	if services == nil {
		return fmt.Errorf("core services not initialized")
	}

	if services.StrategyManager == nil {
		return fmt.Errorf("strategy manager not initialized")
	}

	// Create and start subscriber
	e.subscriber = subscriber.NewSubscriber(
		natsConn,
		services.StrategyManager,
		e.logger,
	)

	if err := e.subscriber.Start(ctx); err != nil {
		return fmt.Errorf("failed to start subscriber: %w", err)
	}

	e.logger.Info("Event subscriber started successfully")
	e.logger.Info("Listening for events on NATS channels:")
	e.logger.Info("  - internal.event.critical")
	e.logger.Info("  - internal.event.anomaly")
	e.logger.Info("  - internal.event.* (debug)")

	return nil
}

// Close stops the event subscriber.
func (e *EventSubscriberInitializer) Close(ctx context.Context) error {
	if e.subscriber != nil {
		if err := e.subscriber.Stop(); err != nil {
			e.logger.Errorw("Failed to stop subscriber", "error", err)
			return err
		}
	}
	return nil
}

// HealthCheck checks subscriber health.
func (e *EventSubscriberInitializer) HealthCheck(ctx context.Context) error {
	// Subscriber health is checked via NATS connection
	return nil
}
