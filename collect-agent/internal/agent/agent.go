package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/kart/k8s-agent/collect-agent/internal/types"
	"github.com/kart/k8s-agent/collect-agent/internal/utils"
)

// Agent represents the main collect agent that coordinates all components
type Agent struct {
	config     *types.AgentConfig
	clusterID  string
	clientset  kubernetes.Interface
	logger     *zap.Logger

	// Components
	eventWatcher         *EventWatcher
	metricsCollector     *MetricsCollector
	commandExecutor      *CommandExecutor
	communicationManager *CommunicationManager

	// Channels for inter-component communication
	eventChan   chan *types.Event
	metricsChan chan *types.Metrics
	commandChan chan *types.Command
	resultChan  chan *types.CommandResult

	// Control
	stopCh  chan struct{}
	wg      sync.WaitGroup
	running bool
	mu      sync.RWMutex

	// Metrics
	startTime time.Time
}

// New creates a new Agent instance
func New(config *types.AgentConfig, logger *zap.Logger) (*Agent, error) {
	// Create Kubernetes clientset
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	// Detect cluster ID if not provided
	clusterID := config.ClusterID
	if clusterID == "" {
		detector := utils.NewClusterIDDetector(clientset, logger)
		clusterID, err = detector.DetectClusterID(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to detect cluster ID: %w", err)
		}
		config.ClusterID = clusterID
	}

	agent := &Agent{
		config:    config,
		clusterID: clusterID,
		clientset: clientset,
		logger:    logger.With(zap.String("cluster_id", clusterID)),

		eventChan:   make(chan *types.Event, config.BufferSize),
		metricsChan: make(chan *types.Metrics, 100),
		commandChan: make(chan *types.Command, 100),
		resultChan:  make(chan *types.CommandResult, 100),

		stopCh:    make(chan struct{}),
		startTime: time.Now(),
	}

	// Initialize components
	if err := agent.initializeComponents(); err != nil {
		return nil, fmt.Errorf("failed to initialize components: %w", err)
	}

	return agent, nil
}

// initializeComponents initializes all agent components
func (a *Agent) initializeComponents() error {
	// Initialize event watcher
	if a.config.EnableEvents {
		a.eventWatcher = NewEventWatcher(a.clientset, a.clusterID, a.eventChan, a.logger)
	}

	// Initialize metrics collector
	if a.config.EnableMetrics {
		a.metricsCollector = NewMetricsCollector(a.clientset, a.clusterID, a.metricsChan, a.logger)
	}

	// Initialize command executor
	a.commandExecutor = NewCommandExecutor(a.clientset, a.clusterID, a.logger)

	// Initialize communication manager
	a.communicationManager = NewCommunicationManager(
		a.config,
		a.clusterID,
		a.eventChan,
		a.metricsChan,
		a.resultChan,
		a.handleCommand,
		a.logger,
	)

	return nil
}

// Start starts the agent and all its components
func (a *Agent) Start(ctx context.Context) error {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return fmt.Errorf("agent already running")
	}
	a.running = true
	a.mu.Unlock()

	a.logger.Info("Starting collect agent",
		zap.String("cluster_id", a.clusterID),
		zap.String("central_endpoint", a.config.CentralEndpoint),
		zap.Duration("heartbeat_interval", a.config.HeartbeatInterval),
		zap.Duration("metrics_interval", a.config.MetricsInterval))

	// Start communication manager first
	if err := a.communicationManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start communication manager: %w", err)
	}

	// Start event watcher
	if a.eventWatcher != nil {
		if err := a.eventWatcher.Start(ctx); err != nil {
			return fmt.Errorf("failed to start event watcher: %w", err)
		}
	}

	// Start metrics collector
	if a.metricsCollector != nil {
		a.wg.Add(1)
		go func() {
			defer a.wg.Done()
			a.metricsCollector.Start(ctx, a.config.MetricsInterval)
		}()
	}

	// Start command processor
	a.wg.Add(1)
	go a.processCommands(ctx)

	a.logger.Info("Collect agent started successfully")

	// Wait for context cancellation
	<-ctx.Done()
	return a.Stop()
}

// Stop stops the agent and all its components
func (a *Agent) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.running {
		return nil
	}

	a.logger.Info("Stopping collect agent")

	// Signal all goroutines to stop
	close(a.stopCh)

	// Stop components
	if a.eventWatcher != nil {
		a.eventWatcher.Stop()
	}

	if a.metricsCollector != nil {
		a.metricsCollector.Stop()
	}

	if a.communicationManager != nil {
		a.communicationManager.Stop()
	}

	// Wait for all goroutines to finish
	a.wg.Wait()

	// Close channels
	close(a.eventChan)
	close(a.metricsChan)
	close(a.commandChan)
	close(a.resultChan)

	a.running = false
	a.logger.Info("Collect agent stopped")

	return nil
}

// handleCommand handles incoming commands from the communication manager
func (a *Agent) handleCommand(cmd *types.Command) {
	select {
	case a.commandChan <- cmd:
		a.logger.Debug("Command queued for processing", zap.String("command_id", cmd.ID))
	default:
		a.logger.Warn("Command channel full, dropping command", zap.String("command_id", cmd.ID))
	}
}

// processCommands processes commands from the command channel
func (a *Agent) processCommands(ctx context.Context) {
	defer a.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-a.stopCh:
			return
		case cmd := <-a.commandChan:
			if cmd == nil {
				continue
			}

			a.logger.Info("Processing command",
				zap.String("command_id", cmd.ID),
				zap.String("tool", cmd.Tool),
				zap.String("action", cmd.Action))

			// Execute command
			result := a.commandExecutor.Execute(ctx, *cmd)

			// Send result back
			select {
			case a.resultChan <- result:
				a.logger.Debug("Command result queued", zap.String("command_id", cmd.ID))
			default:
				a.logger.Warn("Result channel full, dropping result", zap.String("command_id", cmd.ID))
			}
		}
	}
}

// GetStatus returns the current status of the agent
func (a *Agent) GetStatus() AgentStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return AgentStatus{
		ClusterID:         a.clusterID,
		Running:           a.running,
		StartTime:         a.startTime,
		Uptime:           time.Since(a.startTime),
		EventQueueSize:   len(a.eventChan),
		MetricsQueueSize: len(a.metricsChan),
		CommandQueueSize: len(a.commandChan),
		ResultQueueSize:  len(a.resultChan),
		Connected:        a.communicationManager != nil && a.communicationManager.IsConnected(),
	}
}

// AgentStatus represents the current status of the agent
type AgentStatus struct {
	ClusterID         string        `json:"cluster_id"`
	Running           bool          `json:"running"`
	StartTime         time.Time     `json:"start_time"`
	Uptime           time.Duration `json:"uptime"`
	EventQueueSize   int           `json:"event_queue_size"`
	MetricsQueueSize int           `json:"metrics_queue_size"`
	CommandQueueSize int           `json:"command_queue_size"`
	ResultQueueSize  int           `json:"result_queue_size"`
	Connected        bool          `json:"connected"`
}

// IsHealthy returns true if the agent is healthy
func (a *Agent) IsHealthy() bool {
	status := a.GetStatus()
	return status.Running && status.Connected
}

// IsReady returns true if the agent is ready to serve
func (a *Agent) IsReady() bool {
	return a.IsHealthy()
}