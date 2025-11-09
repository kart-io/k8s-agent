package initializers

import (
	"context"

	"google.golang.org/grpc"

	commonserver "github.com/kart-io/k8s-agent/common/server"
	"github.com/kart-io/k8s-agent/internal/monitor/service"
	monitorv1 "github.com/kart-io/k8s-agent/pkg/api/monitor/v1"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	commoninitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// GRPCServerInitializer initializes the gRPC server for monitor service.
//
// DESIGN NOTE: This initializer:
// 1. Wraps pkg/initializers.GRPCServerInitializer for standard gRPC lifecycle
// 2. Creates three gRPC service implementations (Monitor, AlertRule, MetricsCollector)
// 3. Registers all services with the gRPC server during initialization
// 4. Skips initialization if GRPC is disabled in config
type GRPCServerInitializer struct {
	standardInit *commoninitializers.GRPCServerInitializer
	opts         *commonapp.StandardOptions
	logger       core.Logger

	// Dependencies (injected by Wire)
	dbInit    *DatabaseInitializer
	redisInit *RedisInitializer

	// gRPC service implementations (created during initialization)
	monitorService          *MonitorGRPCService
	alertRuleService        *AlertRuleGRPCService
	metricsCollectorService *MetricsCollectorGRPCService
}

// NewGRPCServerInitializer 创建gRPC服务器初始化器
func NewGRPCServerInitializer(
	opts *commonapp.StandardOptions,
	logger core.Logger,
	dbInit *DatabaseInitializer,
	redisInit *RedisInitializer,
) *GRPCServerInitializer {
	return &GRPCServerInitializer{
		opts:      opts,
		logger:    logger,
		dbInit:    dbInit,
		redisInit: redisInit,
	}
}

// Name 返回初始化器名称
func (g *GRPCServerInitializer) Name() string {
	return "monitor-grpc-server"
}

// Priority 返回初始化优先级
func (g *GRPCServerInitializer) Priority() int {
	return bootstrap.PriorityGRPC
}

// Initialize 执行初始化
func (g *GRPCServerInitializer) Initialize(ctx context.Context) error {
	if !g.opts.GRPC.Enable {
		g.logger.Infow("gRPC server is disabled, skipping initialization")
		return nil
	}

	g.logger.Infow("Initializing gRPC server for monitor service",
		"host", g.opts.GRPC.Host,
		"port", g.opts.GRPC.Port,
	)

	// 创建Monitor Service（复用现有的service.MonitorService）
	monitorSvc := service.NewMonitorService(g.dbInit.Storage(), g.redisInit.Storage(), g.logger)
	g.monitorService = NewMonitorGRPCService(monitorSvc, g.logger)

	// 创建AlertRule Service
	g.alertRuleService = NewAlertRuleGRPCService(g.logger)

	// 创建MetricsCollector Service
	g.metricsCollectorService = NewMetricsCollectorGRPCService(g.logger)

	serverConfig := &commoninitializers.GRPCServerConfig{
		Name:     g.Name(),
		Priority: g.Priority(),
		Config:   g.opts.GRPC,
		ServiceRegister: func(s *grpc.Server) error {
			// 注册Monitor服务
			monitorv1.RegisterMonitorServiceServer(s, g.monitorService)
			// 注册AlertRule服务
			monitorv1.RegisterAlertRuleServiceServer(s, g.alertRuleService)
			// 注册MetricsCollector服务
			monitorv1.RegisterMetricsCollectorServiceServer(s, g.metricsCollectorService)

			g.logger.Info("Monitor gRPC services registered")
			return nil
		},
	}

	g.standardInit = commoninitializers.NewGRPCServerInitializer(serverConfig, g.logger)
	return g.standardInit.Initialize(ctx)
}

// Close 关闭gRPC服务器
func (g *GRPCServerInitializer) Close(ctx context.Context) error {
	if g.standardInit == nil {
		return nil
	}
	return g.standardInit.Close(ctx)
}

// GetServer 返回服务器实例（实现ServerProvider接口）
func (g *GRPCServerInitializer) GetServer() commonserver.Server {
	if g.standardInit == nil {
		return nil
	}
	return g.standardInit.GetServer()
}

// ============================================================================
// MonitorGRPCService - 监控服务gRPC实现
// ============================================================================

// MonitorGRPCService 实现 monitorv1.MonitorServiceServer
type MonitorGRPCService struct {
	monitorv1.UnimplementedMonitorServiceServer
	monitorService *service.MonitorService
	logger         core.Logger
}

// NewMonitorGRPCService 创建监控gRPC服务
func NewMonitorGRPCService(monitorService *service.MonitorService, logger core.Logger) *MonitorGRPCService {
	return &MonitorGRPCService{
		monitorService: monitorService,
		logger:         logger.With("component", "monitor-grpc-service"),
	}
}

// GetMetricsSummary 获取监控概览
func (s *MonitorGRPCService) GetMetricsSummary(ctx context.Context, req *monitorv1.GetMetricsSummaryRequest) (*monitorv1.MetricsSummary, error) {
	s.logger.Infow("GetMetricsSummary RPC called")

	// 调用现有的服务方法
	summary, err := s.monitorService.GetMetricsSummary(ctx)
	if err != nil {
		return nil, err
	}

	// TODO: 转换为Proto消息
	_ = summary

	return &monitorv1.MetricsSummary{}, nil
}

// GetAgentMetrics 获取Agent指标
func (s *MonitorGRPCService) GetAgentMetrics(ctx context.Context, req *monitorv1.GetAgentMetricsRequest) (*monitorv1.GetAgentMetricsResponse, error) {
	s.logger.Infow("GetAgentMetrics RPC called")

	// TODO: 实现
	return &monitorv1.GetAgentMetricsResponse{
		Metrics: []*monitorv1.AgentMetrics{},
	}, nil
}

// QueryMetrics 查询指标数据
func (s *MonitorGRPCService) QueryMetrics(ctx context.Context, req *monitorv1.QueryMetricsRequest) (*monitorv1.QueryMetricsResponse, error) {
	s.logger.Infow("QueryMetrics RPC called", "metric_name", req.MetricName)

	// TODO: 实现
	return &monitorv1.QueryMetricsResponse{
		MetricName: req.MetricName,
		DataPoints: []*monitorv1.MetricDataPoint{},
	}, nil
}

// GetAlert 获取告警
func (s *MonitorGRPCService) GetAlert(ctx context.Context, req *monitorv1.GetAlertRequest) (*monitorv1.Alert, error) {
	s.logger.Infow("GetAlert RPC called", "alert_id", req.AlertId)

	// TODO: 实现
	return &monitorv1.Alert{
		Id: req.AlertId,
	}, nil
}

// ListAlerts 列出告警
func (s *MonitorGRPCService) ListAlerts(ctx context.Context, req *monitorv1.ListAlertsRequest) (*monitorv1.ListAlertsResponse, error) {
	s.logger.Infow("ListAlerts RPC called")

	// TODO: 实现
	return &monitorv1.ListAlertsResponse{
		Alerts: []*monitorv1.Alert{},
	}, nil
}

// CreateAlert 创建告警
func (s *MonitorGRPCService) CreateAlert(ctx context.Context, req *monitorv1.CreateAlertRequest) (*monitorv1.Alert, error) {
	s.logger.Infow("CreateAlert RPC called", "name", req.Name)

	// TODO: 实现
	return &monitorv1.Alert{
		Name: req.Name,
	}, nil
}

// UpdateAlert 更新告警
func (s *MonitorGRPCService) UpdateAlert(ctx context.Context, req *monitorv1.UpdateAlertRequest) (*monitorv1.Alert, error) {
	s.logger.Infow("UpdateAlert RPC called", "alert_id", req.AlertId)

	// TODO: 实现
	return &monitorv1.Alert{
		Id: req.AlertId,
	}, nil
}

// DeleteAlert 删除告警
func (s *MonitorGRPCService) DeleteAlert(ctx context.Context, req *monitorv1.DeleteAlertRequest) (*monitorv1.DeleteAlertResponse, error) {
	s.logger.Infow("DeleteAlert RPC called", "alert_id", req.AlertId)

	// TODO: 实现
	return &monitorv1.DeleteAlertResponse{
		Message: "Alert deleted successfully",
	}, nil
}

// AcknowledgeAlert 确认告警
func (s *MonitorGRPCService) AcknowledgeAlert(ctx context.Context, req *monitorv1.AcknowledgeAlertRequest) (*monitorv1.AcknowledgeAlertResponse, error) {
	s.logger.Infow("AcknowledgeAlert RPC called", "alert_id", req.AlertId)

	// TODO: 实现
	return &monitorv1.AcknowledgeAlertResponse{
		Message: "Alert acknowledged",
	}, nil
}

// ============================================================================
// AlertRuleGRPCService - 告警规则服务gRPC实现
// ============================================================================

// AlertRuleGRPCService 实现 monitorv1.AlertRuleServiceServer
type AlertRuleGRPCService struct {
	monitorv1.UnimplementedAlertRuleServiceServer
	logger core.Logger
}

// NewAlertRuleGRPCService 创建告警规则gRPC服务
func NewAlertRuleGRPCService(logger core.Logger) *AlertRuleGRPCService {
	return &AlertRuleGRPCService{
		logger: logger.With("component", "alert-rule-grpc-service"),
	}
}

// GetAlertRule 获取告警规则
func (s *AlertRuleGRPCService) GetAlertRule(ctx context.Context, req *monitorv1.GetAlertRuleRequest) (*monitorv1.AlertRule, error) {
	s.logger.Infow("GetAlertRule RPC called", "rule_id", req.RuleId)

	// TODO: 实现
	return &monitorv1.AlertRule{
		Id: req.RuleId,
	}, nil
}

// ListAlertRules 列出告警规则
func (s *AlertRuleGRPCService) ListAlertRules(ctx context.Context, req *monitorv1.ListAlertRulesRequest) (*monitorv1.ListAlertRulesResponse, error) {
	s.logger.Infow("ListAlertRules RPC called")

	// TODO: 实现
	return &monitorv1.ListAlertRulesResponse{
		Rules: []*monitorv1.AlertRule{},
	}, nil
}

// CreateAlertRule 创建告警规则
func (s *AlertRuleGRPCService) CreateAlertRule(ctx context.Context, req *monitorv1.CreateAlertRuleRequest) (*monitorv1.AlertRule, error) {
	s.logger.Infow("CreateAlertRule RPC called", "name", req.Name)

	// TODO: 实现
	return &monitorv1.AlertRule{
		Name: req.Name,
	}, nil
}

// UpdateAlertRule 更新告警规则
func (s *AlertRuleGRPCService) UpdateAlertRule(ctx context.Context, req *monitorv1.UpdateAlertRuleRequest) (*monitorv1.AlertRule, error) {
	s.logger.Infow("UpdateAlertRule RPC called", "rule_id", req.RuleId)

	// TODO: 实现
	return &monitorv1.AlertRule{
		Id: req.RuleId,
	}, nil
}

// DeleteAlertRule 删除告警规则
func (s *AlertRuleGRPCService) DeleteAlertRule(ctx context.Context, req *monitorv1.DeleteAlertRuleRequest) (*monitorv1.DeleteAlertRuleResponse, error) {
	s.logger.Infow("DeleteAlertRule RPC called", "rule_id", req.RuleId)

	// TODO: 实现
	return &monitorv1.DeleteAlertRuleResponse{
		Message: "Alert rule deleted successfully",
	}, nil
}

// ToggleAlertRule 启用/禁用告警规则
func (s *AlertRuleGRPCService) ToggleAlertRule(ctx context.Context, req *monitorv1.ToggleAlertRuleRequest) (*monitorv1.AlertRule, error) {
	s.logger.Infow("ToggleAlertRule RPC called", "rule_id", req.RuleId, "enabled", req.Enabled)

	// TODO: 实现
	return &monitorv1.AlertRule{
		Id:      req.RuleId,
		Enabled: req.Enabled,
	}, nil
}

// ============================================================================
// MetricsCollectorGRPCService - 指标收集服务gRPC实现
// ============================================================================

// MetricsCollectorGRPCService 实现 monitorv1.MetricsCollectorServiceServer
type MetricsCollectorGRPCService struct {
	monitorv1.UnimplementedMetricsCollectorServiceServer
	logger core.Logger
}

// NewMetricsCollectorGRPCService 创建指标收集gRPC服务
func NewMetricsCollectorGRPCService(logger core.Logger) *MetricsCollectorGRPCService {
	return &MetricsCollectorGRPCService{
		logger: logger.With("component", "metrics-collector-grpc-service"),
	}
}

// ReportMetrics 上报指标（流式）
func (s *MetricsCollectorGRPCService) ReportMetrics(stream monitorv1.MetricsCollectorService_ReportMetricsServer) error {
	s.logger.Infow("ReportMetrics stream RPC called")

	// TODO: 实现流式接收指标数据
	// 持续接收客户端发送的指标数据
	receivedCount := 0
	for {
		metricData, err := stream.Recv()
		if err != nil {
			// 流结束
			break
		}

		// 处理接收到的指标数据
		s.logger.Debugw("Received metric data",
			"agent_id", metricData.AgentId,
			"metric_name", metricData.MetricName,
			"value", metricData.Value,
		)
		receivedCount++

		// TODO: 存储指标数据
	}

	// 返回响应
	return stream.SendAndClose(&monitorv1.ReportMetricsResponse{
		ReceivedCount: int32(receivedCount),
		Message:       "Metrics received successfully",
	})
}

// GetCollectorConfig 获取采集器配置
func (s *MetricsCollectorGRPCService) GetCollectorConfig(ctx context.Context, req *monitorv1.GetCollectorConfigRequest) (*monitorv1.CollectorConfig, error) {
	s.logger.Infow("GetCollectorConfig RPC called", "agent_id", req.AgentId)

	// TODO: 实现
	return &monitorv1.CollectorConfig{
		IntervalSeconds: 30,
		EnabledMetrics:  []string{"cpu", "memory", "disk", "network"},
	}, nil
}
