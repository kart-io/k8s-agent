package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/kart-io/k8s-agent/internal/monitor/storage"
	"github.com/kart-io/k8s-agent/internal/monitor/types"
	"github.com/kart-io/logger/core"
)

type MonitorService struct {
	db    *storage.PostgresStorage
	redis *storage.RedisStorage
	log   core.Logger
}

func NewMonitorService(db *storage.PostgresStorage, redis *storage.RedisStorage, logger core.Logger) *MonitorService {
	return &MonitorService{
		db:    db,
		redis: redis,
		log:   logger,
	}
}

// GetMetricsSummary 获取监控概览
func (s *MonitorService) GetMetricsSummary(ctx context.Context) (*types.MetricsSummary, error) {
	// 尝试从缓存获取
	cached, err := s.redis.GetMetricsSummary(ctx, "metrics:summary")
	if err == nil && cached != "" {
		var summary types.MetricsSummary
		if err := json.Unmarshal([]byte(cached), &summary); err == nil {
			return &summary, nil
		}
	}

	// 从数据库查询最新的监控概览
	query := `
		SELECT total_agents, online_agents, offline_agents, total_events,
		       critical_events, total_commands, running_commands, cpu_usage,
		       memory_usage, disk_usage, network_in, network_out, created_at
		FROM metrics_summary
		ORDER BY created_at DESC
		LIMIT 1
	`

	var summary types.MetricsSummary
	err = s.db.DB().QueryRowContext(ctx, query).Scan(
		&summary.TotalAgents,
		&summary.OnlineAgents,
		&summary.OfflineAgents,
		&summary.TotalEvents,
		&summary.CriticalEvents,
		&summary.TotalCommands,
		&summary.RunningCommands,
		&summary.CPUUsage,
		&summary.MemoryUsage,
		&summary.DiskUsage,
		&summary.NetworkIn,
		&summary.NetworkOut,
		&summary.LastUpdateTime,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// 返回空数据
			return &types.MetricsSummary{
				LastUpdateTime: time.Now(),
			}, nil
		}
		return nil, err
	}

	// 缓存结果
	data, _ := json.Marshal(summary)
	s.redis.SetMetricsSummary(ctx, "metrics:summary", string(data), 30*time.Second)

	return &summary, nil
}

// SaveMetricsSummary 保存监控概览
func (s *MonitorService) SaveMetricsSummary(ctx context.Context, summary *types.MetricsSummary) error {
	query := `
		INSERT INTO metrics_summary
		(total_agents, online_agents, offline_agents, total_events,
		 critical_events, total_commands, running_commands, cpu_usage,
		 memory_usage, disk_usage, network_in, network_out)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := s.db.DB().ExecContext(ctx, query,
		summary.TotalAgents,
		summary.OnlineAgents,
		summary.OfflineAgents,
		summary.TotalEvents,
		summary.CriticalEvents,
		summary.TotalCommands,
		summary.RunningCommands,
		summary.CPUUsage,
		summary.MemoryUsage,
		summary.DiskUsage,
		summary.NetworkIn,
		summary.NetworkOut,
	)

	// 清除缓存
	if err == nil {
		s.redis.Client().Del(ctx, "metrics:summary")
	}

	return err
}

// GetAgentMetrics 获取 Agent 指标列表
func (s *MonitorService) GetAgentMetrics(ctx context.Context, limit, offset int) ([]types.AgentMetrics, error) {
	query := `
		SELECT agent_id, agent_name, cluster_id, status, cpu_usage,
		       memory_usage, event_count, command_count, last_heartbeat, uptime
		FROM agent_metrics
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.DB().QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []types.AgentMetrics
	for rows.Next() {
		var m types.AgentMetrics
		if err := rows.Scan(
			&m.AgentID,
			&m.AgentName,
			&m.ClusterID,
			&m.Status,
			&m.CPUUsage,
			&m.MemoryUsage,
			&m.EventCount,
			&m.CommandCount,
			&m.LastHeartbeat,
			&m.Uptime,
		); err != nil {
			s.log.Errorw("Failed to scan agent metrics", "error", err)
			continue
		}
		metrics = append(metrics, m)
	}

	return metrics, nil
}

// GetTrendData 获取趋势数据
func (s *MonitorService) GetTrendData(ctx context.Context, hours int) ([]types.TrendData, error) {
	query := `
		SELECT timestamp, metrics
		FROM trend_data
		WHERE timestamp >= $1
		ORDER BY timestamp ASC
	`

	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	rows, err := s.db.DB().QueryContext(ctx, query, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trends []types.TrendData
	for rows.Next() {
		var t types.TrendData
		var metricsJSON []byte
		if err := rows.Scan(&t.Timestamp, &metricsJSON); err != nil {
			s.log.Errorw("Failed to scan trend data", "error", err)
			continue
		}
		if err := json.Unmarshal(metricsJSON, &t.Metrics); err != nil {
			s.log.Errorw("Failed to unmarshal metrics", "error", err)
			continue
		}
		trends = append(trends, t)
	}

	return trends, nil
}
