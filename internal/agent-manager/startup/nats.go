// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package startup

import (
	"context"

	"github.com/kart-io/k8s-agent/internal/agent-manager/nats"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/logger/core"
)

// NATSInitializer initializes NATS server for agent communication.
type NATSInitializer struct {
	opts     *commonapp.StandardOptions
	logger   core.Logger
	services *BusinessServicesInitializer

	natsServer *nats.Server
}

// NewNATSInitializer creates a new NATS initializer.
func NewNATSInitializer(
	opts *commonapp.StandardOptions,
	logger core.Logger,
	services *BusinessServicesInitializer,
) *NATSInitializer {
	return &NATSInitializer{
		opts:     opts,
		logger:   logger,
		services: services,
	}
}

// Name returns the initializer name.
func (n *NATSInitializer) Name() string {
	return "nats"
}

// Priority returns initialization priority.
func (n *NATSInitializer) Priority() int {
	return 500 // After infrastructure, before dispatcher
}

// Initialize creates and starts NATS server.
func (n *NATSInitializer) Initialize(ctx context.Context) error {
	n.logger.Infow("Initializing NATS server",
		"url", n.opts.NATS.URL,
		"max_reconnect", n.opts.NATS.MaxReconnect,
		"reconnect_delay_initial", n.opts.NATS.ReconnectDelayInitial.String(),
		"reconnect_delay_max", n.opts.NATS.ReconnectDelayMax.String(),
	)

	// Get business services
	businessServices := n.services.Services()

	// Create NATS server
	n.natsServer = nats.NewServer(
		businessServices.Registry,
		businessServices.EventProcessor,
		n.logger,
		nats.WithURL(n.opts.NATS.URL),
		nats.WithMaxReconnect(n.opts.NATS.MaxReconnect),
		nats.WithReconnectWait(n.opts.NATS.ReconnectWait),
		nats.WithPingInterval(n.opts.NATS.PingInterval),
		nats.WithMaxPingsOut(n.opts.NATS.MaxPingsOut),
		nats.WithReconnectDelayInitial(n.opts.NATS.ReconnectDelayInitial),
		nats.WithReconnectDelayMax(n.opts.NATS.ReconnectDelayMax),
		nats.WithReconnectBackoffFactor(n.opts.NATS.ReconnectBackoffFactor),
	)

	// Start NATS server
	if err := n.natsServer.Start(ctx); err != nil {
		return err
	}

	n.logger.Infow("NATS server initialized successfully")
	return nil
}

// Server returns the NATS server instance.
func (n *NATSInitializer) Server() *nats.Server {
	return n.natsServer
}

// Close stops the NATS server.
func (n *NATSInitializer) Close(ctx context.Context) error {
	if n.natsServer != nil {
		n.logger.Infow("Stopping NATS server")
		return n.natsServer.Stop()
	}
	return nil
}

// DispatcherInitializer wires up the command dispatcher with NATS.
type DispatcherInitializer struct {
	logger   core.Logger
	services *BusinessServicesInitializer
	nats     *NATSInitializer
}

// NewDispatcherInitializer creates a new dispatcher initializer.
func NewDispatcherInitializer(
	logger core.Logger,
	services *BusinessServicesInitializer,
	nats *NATSInitializer,
) *DispatcherInitializer {
	return &DispatcherInitializer{
		logger:   logger,
		services: services,
		nats:     nats,
	}
}

// Name returns the initializer name.
func (d *DispatcherInitializer) Name() string {
	return "dispatcher"
}

// Priority returns initialization priority.
func (d *DispatcherInitializer) Priority() int {
	return 550 // After NATS, before servers
}

// Initialize wires up command result handler.
func (d *DispatcherInitializer) Initialize(ctx context.Context) error {
	d.logger.Infow("Initializing command dispatcher")

	// Get business services
	businessServices := d.services.Services()

	// Wire up command result handler: NATS server calls dispatcher's HandleCommandResult
	d.nats.Server().SetCommandResultHandler(businessServices.Dispatcher.HandleCommandResult)

	d.logger.Infow("Command dispatcher initialized successfully with NATS result handler")
	return nil
}

// Close is a no-op.
func (d *DispatcherInitializer) Close(ctx context.Context) error {
	return nil
}
