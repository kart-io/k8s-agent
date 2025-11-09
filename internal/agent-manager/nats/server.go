package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/kart-io/k8s-agent/internal/agent-manager/agent"
	"github.com/kart-io/k8s-agent/internal/agent-manager/event"
	"github.com/kart-io/k8s-agent/pkg/types"
	"github.com/kart-io/logger/core"
)

// ServerOptions NATS Server 配置选项.
type ServerOptions struct {
	url                    string
	maxReconnect           int
	reconnectWait          time.Duration
	pingInterval           time.Duration
	maxPingsOut            int
	enableJetStream        bool
	reconnectDelayMax      time.Duration // 最大重连延迟时间（指数退避上限）
	reconnectDelayInitial  time.Duration // 初始重连延迟时间
	reconnectBackoffFactor float64       // 重连退避因子（每次重连后延迟增加的倍数）
}

// ServerOption NATS Server 配置选项函数.
type ServerOption func(*ServerOptions)

// WithURL 设置 NATS 服务器地址.
func WithURL(url string) ServerOption {
	return func(o *ServerOptions) {
		o.url = url
	}
}

// WithMaxReconnect 设置最大重连次数.
func WithMaxReconnect(n int) ServerOption {
	return func(o *ServerOptions) {
		o.maxReconnect = n
	}
}

// WithReconnectWait 设置重连等待时间.
func WithReconnectWait(d time.Duration) ServerOption {
	return func(o *ServerOptions) {
		o.reconnectWait = d
	}
}

// WithPingInterval 设置 Ping 间隔时间.
func WithPingInterval(d time.Duration) ServerOption {
	return func(o *ServerOptions) {
		o.pingInterval = d
	}
}

// WithMaxPingsOut 设置最大未响应 Ping 数量.
func WithMaxPingsOut(n int) ServerOption {
	return func(o *ServerOptions) {
		o.maxPingsOut = n
	}
}

// WithEnableJetStream 启用 JetStream.
func WithEnableJetStream(enable bool) ServerOption {
	return func(o *ServerOptions) {
		o.enableJetStream = enable
	}
}

// WithReconnectDelayMax 设置最大重连延迟时间（指数退避上限）.
func WithReconnectDelayMax(d time.Duration) ServerOption {
	return func(o *ServerOptions) {
		o.reconnectDelayMax = d
	}
}

// WithReconnectDelayInitial 设置初始重连延迟时间.
func WithReconnectDelayInitial(d time.Duration) ServerOption {
	return func(o *ServerOptions) {
		o.reconnectDelayInitial = d
	}
}

// WithReconnectBackoffFactor 设置重连退避因子.
func WithReconnectBackoffFactor(factor float64) ServerOption {
	return func(o *ServerOptions) {
		o.reconnectBackoffFactor = factor
	}
}

// defaultServerOptions 返回默认 NATS Server 配置.
func defaultServerOptions() *ServerOptions {
	return &ServerOptions{
		url:                    "nats://localhost:4222",
		maxReconnect:           10,
		reconnectWait:          2 * time.Second,
		pingInterval:           20 * time.Second,
		maxPingsOut:            2,
		enableJetStream:        false,
		reconnectDelayInitial:  1 * time.Second,  // 初始延迟 1 秒
		reconnectDelayMax:      30 * time.Second, // 最大延迟 30 秒
		reconnectBackoffFactor: 2.0,              // 每次重连延迟翻倍
	}
}

// CommandResultHandler is a callback for handling command results.
type CommandResultHandler func(ctx context.Context, result *types.CommandResult) error

// Server manages NATS server connection and subscriptions.
type Server struct {
	conn    *nats.Conn
	logger  core.Logger
	options *ServerOptions

	// Components
	registry             *agent.Registry
	eventProcessor       *event.Processor
	commandResultHandler CommandResultHandler // Handler for command results

	// Subscriptions
	subscriptions []*nats.Subscription
	mu            sync.RWMutex
	stopCh        chan struct{}
	wg            sync.WaitGroup

	// Reconnection state
	reconnectCount        int64 // 重连次数计数器
	reconnectSuccess      int64 // 重连成功计数器
	reconnectFailed       int64 // 重连失败计数器
	lastReconnectTime     time.Time
	currentReconnectDelay time.Duration // 当前重连延迟时间

	// Metrics
	messagesReceived int64
	messagesSent     int64
	errorCount       int64
}

// NewServer creates a new NATS server instance.
func NewServer(
	registry *agent.Registry,
	eventProcessor *event.Processor,
	logger core.Logger,
	opts ...ServerOption,
) *Server {
	// 应用默认配置
	options := defaultServerOptions()

	// 应用用户配置
	for _, opt := range opts {
		opt(options)
	}

	return &Server{
		options:               options,
		registry:              registry,
		eventProcessor:        eventProcessor,
		logger:                logger.With("component", "nats-server"),
		stopCh:                make(chan struct{}),
		currentReconnectDelay: options.reconnectDelayInitial, // 初始化重连延迟
	}
}

// SetCommandResultHandler sets the handler for command results.
func (s *Server) SetCommandResultHandler(handler CommandResultHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.commandResultHandler = handler
}

// Start starts the NATS server and subscriptions.
func (s *Server) Start(ctx context.Context) error {
	s.logger.Infow("Starting NATS server", "url", s.options.url)

	// Connect to NATS
	if err := s.connect(); err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// Setup subscriptions
	if err := s.setupSubscriptions(); err != nil {
		return fmt.Errorf("failed to setup subscriptions: %w", err)
	}

	// Start connection monitor
	s.wg.Add(1)
	go s.connectionMonitor()

	s.logger.Info("NATS server started successfully")

	return nil
}

// Stop stops the NATS server.
func (s *Server) Stop() error {
	s.logger.Info("Stopping NATS server")

	close(s.stopCh)
	s.wg.Wait()

	// Unsubscribe all
	s.mu.Lock()
	for _, sub := range s.subscriptions {
		if err := sub.Unsubscribe(); err != nil {
			s.logger.Warnw("Failed to unsubscribe", "error", err)
		}
	}
	s.subscriptions = nil
	s.mu.Unlock()

	// Close connection
	if s.conn != nil {
		s.conn.Close()
	}

	s.logger.Info("NATS server stopped")

	return nil
}

// connect establishes connection to NATS server.
func (s *Server) connect() error {
	opts := []nats.Option{
		nats.Name("agent-manager"),
		nats.MaxReconnects(s.options.maxReconnect),
		nats.ReconnectWait(s.options.reconnectWait),
		nats.PingInterval(s.options.pingInterval),
		nats.MaxPingsOutstanding(s.options.maxPingsOut),
		nats.DisconnectErrHandler(s.handleDisconnect),
		nats.ReconnectHandler(s.handleReconnect),
		nats.ErrorHandler(s.handleError),
		nats.ClosedHandler(s.handleClosed),
		// 使用自定义重连延迟函数实现指数退避
		nats.CustomReconnectDelay(s.customReconnectDelay),
	}

	conn, err := nats.Connect(s.options.url, opts...)
	if err != nil {
		return err
	}

	s.conn = conn
	s.logger.Infow("Connected to NATS", "url", s.options.url)

	return nil
}

// setupSubscriptions sets up all NATS subscriptions.
func (s *Server) setupSubscriptions() error {
	// Subscribe to agent registration
	if err := s.subscribeRegister(); err != nil {
		return fmt.Errorf("failed to subscribe to register: %w", err)
	}

	// Subscribe to agent heartbeat
	if err := s.subscribeHeartbeat(); err != nil {
		return fmt.Errorf("failed to subscribe to heartbeat: %w", err)
	}

	// Subscribe to agent events
	if err := s.subscribeEvents(); err != nil {
		return fmt.Errorf("failed to subscribe to events: %w", err)
	}

	// Subscribe to agent metrics
	if err := s.subscribeMetrics(); err != nil {
		return fmt.Errorf("failed to subscribe to metrics: %w", err)
	}

	// Subscribe to command results
	if err := s.subscribeResults(); err != nil {
		return fmt.Errorf("failed to subscribe to results: %w", err)
	}

	return nil
}

// subscribeRegister subscribes to agent registration messages.
func (s *Server) subscribeRegister() error {
	subject := "aetherius.agent.*.register"

	sub, err := s.conn.Subscribe(subject, func(msg *nats.Msg) {
		s.handleRegister(msg)
	})
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.subscriptions = append(s.subscriptions, sub)
	s.mu.Unlock()

	s.logger.Infow("Subscribed to agent registration", "subject", subject)

	return nil
}

// subscribeHeartbeat subscribes to agent heartbeat messages.
func (s *Server) subscribeHeartbeat() error {
	subject := "aetherius.agent.*.heartbeat"

	sub, err := s.conn.Subscribe(subject, func(msg *nats.Msg) {
		s.handleHeartbeat(msg)
	})
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.subscriptions = append(s.subscriptions, sub)
	s.mu.Unlock()

	s.logger.Infow("Subscribed to agent heartbeat", "subject", subject)

	return nil
}

// subscribeEvents subscribes to agent event messages.
func (s *Server) subscribeEvents() error {
	subject := "aetherius.agent.*.event"

	sub, err := s.conn.Subscribe(subject, func(msg *nats.Msg) {
		s.handleEvent(msg)
	})
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.subscriptions = append(s.subscriptions, sub)
	s.mu.Unlock()

	s.logger.Infow("Subscribed to agent events", "subject", subject)

	return nil
}

// subscribeMetrics subscribes to agent metrics messages.
func (s *Server) subscribeMetrics() error {
	subject := "aetherius.agent.*.metrics"

	sub, err := s.conn.Subscribe(subject, func(msg *nats.Msg) {
		s.handleMetrics(msg)
	})
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.subscriptions = append(s.subscriptions, sub)
	s.mu.Unlock()

	s.logger.Infow("Subscribed to agent metrics", "subject", subject)

	return nil
}

// subscribeResults subscribes to command result messages.
func (s *Server) subscribeResults() error {
	subject := "aetherius.agent.*.result"

	sub, err := s.conn.Subscribe(subject, func(msg *nats.Msg) {
		s.handleResult(msg)
	})
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.subscriptions = append(s.subscriptions, sub)
	s.mu.Unlock()

	s.logger.Infow("Subscribed to command results", "subject", subject)

	return nil
}

// Message handlers

// handleRegister handles agent registration messages.
func (s *Server) handleRegister(msg *nats.Msg) {
	s.messagesReceived++

	// Unmarshal into a generic map to extract both agent and cluster info
	var rawData map[string]interface{}
	if err := json.Unmarshal(msg.Data, &rawData); err != nil {
		s.logger.Errorw("Failed to unmarshal register message", "error", err)
		s.errorCount++
		return
	}

	// Extract cluster information if provided
	k8sVersion, _ := rawData["k8s_version"].(string)
	apiServer, _ := rawData["api_server"].(string)

	// Unmarshal again into Agent struct
	var agentInfo types.Agent
	if err := json.Unmarshal(msg.Data, &agentInfo); err != nil {
		s.logger.Errorw("Failed to unmarshal register message", "error", err)
		s.errorCount++
		return
	}

	ctx := context.Background()
	if err := s.registry.RegisterAgent(ctx, &agentInfo); err != nil {
		s.logger.Errorw("Failed to register agent",
			"cluster_id", agentInfo.ClusterID,
			"error", err)
		s.errorCount++
		return
	}

	// Update cluster with K8s information if provided
	if k8sVersion != "" || apiServer != "" {
		if err := s.registry.UpdateClusterInfo(ctx, agentInfo.ClusterID, k8sVersion, apiServer); err != nil {
			s.logger.Warnw("Failed to update cluster info",
				"cluster_id", agentInfo.ClusterID,
				"error", err)
		}
	}

	s.logger.Infow("Agent registered successfully",
		"agent_id", agentInfo.ID,
		"cluster_id", agentInfo.ClusterID)

	// Send acknowledgment
	ack := map[string]interface{}{
		"status":   "registered",
		"agent_id": agentInfo.ID,
	}
	s.sendResponse(msg, ack)
}

// handleHeartbeat handles agent heartbeat messages.
func (s *Server) handleHeartbeat(msg *nats.Msg) {
	s.messagesReceived++

	var heartbeat struct {
		AgentID   string    `json:"agent_id"`
		ClusterID string    `json:"cluster_id"`
		Timestamp time.Time `json:"timestamp"`
	}

	if err := json.Unmarshal(msg.Data, &heartbeat); err != nil {
		s.logger.Errorw("Failed to unmarshal heartbeat message", "error", err)
		s.errorCount++
		return
	}

	ctx := context.Background()
	if err := s.registry.UpdateHeartbeat(ctx, heartbeat.AgentID); err != nil {
		s.logger.Warnw("Failed to update heartbeat",
			"agent_id", heartbeat.AgentID,
			"error", err)
		s.errorCount++
		return
	}

	s.logger.Debugw("Heartbeat received",
		"agent_id", heartbeat.AgentID,
		"cluster_id", heartbeat.ClusterID)
}

// handleEvent handles agent event messages.
func (s *Server) handleEvent(msg *nats.Msg) {
	s.messagesReceived++

	var event types.Event
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		s.logger.Errorw("Failed to unmarshal event message", "error", err)
		s.errorCount++
		return
	}

	ctx := context.Background()
	if err := s.eventProcessor.ProcessEvent(ctx, &event); err != nil {
		s.logger.Errorw("Failed to process event",
			"event_id", event.ID,
			"cluster_id", event.ClusterID,
			"error", err)
		s.errorCount++
		return
	}

	s.logger.Debugw("Event processed",
		"event_id", event.ID,
		"cluster_id", event.ClusterID,
		"severity", event.Severity)
}

// handleMetrics handles agent metrics messages.
func (s *Server) handleMetrics(msg *nats.Msg) {
	s.messagesReceived++

	// Parse as generic map to access Data field from collect-agent
	var rawMetrics map[string]interface{}
	if err := json.Unmarshal(msg.Data, &rawMetrics); err != nil {
		s.logger.Errorw("Failed to unmarshal metrics message", "error", err)
		s.errorCount++
		return
	}

	clusterID, _ := rawMetrics["cluster_id"].(string)
	if clusterID == "" {
		s.logger.Warn("Metrics message missing cluster_id")
		return
	}

	// Extract node and pod counts from metrics data
	var nodeCount, podCount int

	if data, ok := rawMetrics["data"].(map[string]interface{}); ok {
		if nodesData, ok := data["nodes"].(map[string]interface{}); ok {
			if total, ok := nodesData["total"].(float64); ok {
				nodeCount = int(total)
			}
		}

		if podsData, ok := data["pods"].(map[string]interface{}); ok {
			if total, ok := podsData["total"].(float64); ok {
				podCount = int(total)
			}
		}
	}

	// Update cluster metrics
	ctx := context.Background()
	if err := s.registry.UpdateClusterMetrics(ctx, clusterID, nodeCount, podCount); err != nil {
		s.logger.Warnw("Failed to update cluster metrics",
			"cluster_id", clusterID,
			"error", err)
	}

	s.logger.Debugw("Metrics received and processed",
		"cluster_id", clusterID,
		"node_count", nodeCount,
		"pod_count", podCount)
}

// handleResult handles command result messages.
func (s *Server) handleResult(msg *nats.Msg) {
	s.messagesReceived++

	var result types.CommandResult
	if err := json.Unmarshal(msg.Data, &result); err != nil {
		s.logger.Errorw("Failed to unmarshal result message", "error", err)
		s.errorCount++
		return
	}

	s.logger.Infow("Command result received",
		"command_id", result.CommandID,
		"cluster_id", result.ClusterID,
		"status", result.Status)

	// Process command result through handler
	s.mu.RLock()
	handler := s.commandResultHandler
	s.mu.RUnlock()

	if handler != nil {
		ctx := context.Background()
		if err := handler(ctx, &result); err != nil {
			s.logger.Errorw("Failed to process command result",
				"command_id", result.CommandID,
				"error", err)
			s.errorCount++
			return
		}
		s.logger.Debugw("Command result processed successfully",
			"command_id", result.CommandID)
	} else {
		s.logger.Warnw("No command result handler configured - result not processed",
			"command_id", result.CommandID)
	}
}

// PublishCommand publishes a command to an agent.
func (s *Server) PublishCommand(clusterID string, cmd *types.Command) error {
	subject := fmt.Sprintf("aetherius.agent.%s.command", clusterID)

	data, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("failed to marshal command: %w", err)
	}

	if err := s.conn.Publish(subject, data); err != nil {
		s.errorCount++
		return fmt.Errorf("failed to publish command: %w", err)
	}

	s.messagesSent++
	s.logger.Infow("Command published",
		"command_id", cmd.ID,
		"cluster_id", clusterID,
		"subject", subject)

	return nil
}

// sendResponse sends a response message.
func (s *Server) sendResponse(msg *nats.Msg, response interface{}) {
	data, err := json.Marshal(response)
	if err != nil {
		s.logger.Errorw("Failed to marshal response", "error", err)
		return
	}

	if err := msg.Respond(data); err != nil {
		s.logger.Errorw("Failed to send response", "error", err)
		return
	}

	s.messagesSent++
}

// Connection event handlers

// customReconnectDelay 实现指数退避重连策略.
func (s *Server) customReconnectDelay(attempts int) time.Duration {
	// 计算指数退避延迟: 初始延迟 * (2^(attempts-1))
	// 使用 float64 计算指数，然后转换回 time.Duration
	baseDelay := float64(s.options.reconnectDelayInitial)
	exponentialFactor := float64(uint(1) << uint(attempts-1)) // 2^(attempts-1)

	delay := time.Duration(baseDelay * exponentialFactor)

	// 应用退避因子
	if s.options.reconnectBackoffFactor > 1.0 {
		delay = time.Duration(float64(delay) * s.options.reconnectBackoffFactor)
	}

	// 限制最大延迟
	if delay > s.options.reconnectDelayMax {
		delay = s.options.reconnectDelayMax
	}

	// 更新当前重连延迟
	s.currentReconnectDelay = delay

	s.logger.Infow("Calculating reconnect delay",
		"attempt", attempts,
		"delay", delay.String(),
		"max_delay", s.options.reconnectDelayMax.String(),
	)

	return delay
}

func (s *Server) handleDisconnect(conn *nats.Conn, err error) {
	s.reconnectCount++
	s.logger.Warnw("Disconnected from NATS",
		"error", err,
		"url", s.options.url,
		"reconnect_count", s.reconnectCount,
		"next_delay", s.currentReconnectDelay.String(),
	)
}

func (s *Server) handleReconnect(conn *nats.Conn) {
	s.reconnectSuccess++
	s.lastReconnectTime = time.Now()

	// 重连成功后重置延迟为初始值
	s.currentReconnectDelay = s.options.reconnectDelayInitial

	s.logger.Infow("Reconnected to NATS",
		"url", conn.ConnectedUrl(),
		"reconnect_count", s.reconnectCount,
		"success_count", s.reconnectSuccess,
	)

	// 重连成功后需要恢复订阅
	if err := s.resubscribeAll(); err != nil {
		s.logger.Errorw("Failed to resubscribe after reconnection", "error", err)
	}
}

func (s *Server) handleClosed(conn *nats.Conn) {
	s.logger.Warnw("NATS connection closed",
		"reconnect_count", s.reconnectCount,
		"success_count", s.reconnectSuccess,
		"failed_count", s.reconnectFailed,
	)
}

func (s *Server) handleError(conn *nats.Conn, sub *nats.Subscription, err error) {
	subject := "unknown"
	if sub != nil {
		subject = sub.Subject
	}

	s.logger.Errorw("NATS error",
		"error", err,
		"subject", subject,
	)
	s.errorCount++
}

// connectionMonitor monitors connection health.
func (s *Server) connectionMonitor() {
	defer s.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			if s.conn == nil || !s.conn.IsConnected() {
				s.logger.Warn("NATS connection lost, attempting reconnect")
				s.reconnectFailed++
			}
		}
	}
}

// resubscribeAll 重新订阅所有主题（在重连成功后调用）.
func (s *Server) resubscribeAll() error {
	s.logger.Infow("Resubscribing to all subjects after reconnection")

	s.mu.Lock()
	defer s.mu.Unlock()

	// 清除旧的订阅（这些订阅已经失效）
	s.subscriptions = nil

	// 重新建立所有订阅
	if err := s.setupSubscriptions(); err != nil {
		return fmt.Errorf("failed to setup subscriptions: %w", err)
	}

	s.logger.Infow("Successfully resubscribed to all subjects",
		"subscription_count", len(s.subscriptions),
	)

	return nil
}

// GetStatistics returns server statistics.
func (s *Server) GetStatistics() map[string]interface{} {
	var connected bool
	var connectedURL string

	if s.conn != nil {
		connected = s.conn.IsConnected()
		connectedURL = s.conn.ConnectedUrl()
	}

	return map[string]interface{}{
		"connected":               connected,
		"connected_url":           connectedURL,
		"messages_received":       s.messagesReceived,
		"messages_sent":           s.messagesSent,
		"error_count":             s.errorCount,
		"subscription_count":      len(s.subscriptions),
		"reconnect_count":         s.reconnectCount,
		"reconnect_success":       s.reconnectSuccess,
		"reconnect_failed":        s.reconnectFailed,
		"last_reconnect_time":     s.lastReconnectTime,
		"current_reconnect_delay": s.currentReconnectDelay.String(),
	}
}

// Health checks NATS server health.
func (s *Server) Health() error {
	if s.conn == nil {
		return fmt.Errorf("not connected")
	}
	if !s.conn.IsConnected() {
		return fmt.Errorf("connection lost")
	}
	return nil
}

// GetConnection returns the underlying NATS connection.
func (s *Server) GetConnection() *nats.Conn {
	return s.conn
}
