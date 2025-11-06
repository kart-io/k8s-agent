package api

import (
	"strings"
	"time"

	"github.com/kart-io/k8s-agent/internal/reasoning/agents/k8s_tool"
	"github.com/kart-io/k8s-agent/internal/reasoning/orchestrator"
)

// convertK8sEventToOrchestratorRequest 将 K8s Event 请求转换为 Orchestrator 请求.
func convertK8sEventToOrchestratorRequest(req *K8sEventRequest) *orchestrator.AnalysisRequest {
	orchReq := &orchestrator.AnalysisRequest{
		ClusterID:   req.ClusterID,
		Timestamp:   time.Now(),
		Language:    "zh-CN", // K8s Event API 默认中文
		DetailLevel: "normal",
	}

	// 从 event 中提取信息
	if req.Event != nil {
		extractEventBasicInfo(req.Event, orchReq)
		extractInvolvedObjectInfo(req.Event, orchReq)
		orchReq.Events = buildEventInfo(req.Event)
	}

	// 设置默认值
	setDefaultValues(orchReq)

	return orchReq
}

// extractEventBasicInfo 提取事件基本信息.
func extractEventBasicInfo(event map[string]interface{}, orchReq *orchestrator.AnalysisRequest) {
	// 提取 reason (作为 failure_type 和 error_message)
	if reason, ok := event["reason"].(string); ok {
		orchReq.FailureType = reason
		orchReq.ErrorMessage = reason
	}

	// 提取 message
	if message, ok := event["message"].(string); ok {
		if orchReq.ErrorMessage == "" {
			orchReq.ErrorMessage = message
		}
	}
}

// extractInvolvedObjectInfo 提取 involvedObject 信息.
func extractInvolvedObjectInfo(event map[string]interface{}, orchReq *orchestrator.AnalysisRequest) {
	involvedObj, ok := event["involvedObject"].(map[string]interface{})
	if !ok {
		return
	}

	if namespace, ok := involvedObj["namespace"].(string); ok {
		orchReq.Namespace = namespace
	}
	if name, ok := involvedObj["name"].(string); ok {
		orchReq.ResourceName = name
	}
	if kind, ok := involvedObj["kind"].(string); ok {
		orchReq.ResourceType = strings.ToLower(kind)
	}
}

// buildEventInfo 构建事件信息.
func buildEventInfo(event map[string]interface{}) []k8s_tool.EventInfo {
	eventInfo := k8s_tool.EventInfo{
		LastTimestamp: time.Now(),
	}

	if eventType, ok := event["type"].(string); ok {
		eventInfo.Type = eventType
	}
	if reason, ok := event["reason"].(string); ok {
		eventInfo.Reason = reason
	}
	if message, ok := event["message"].(string); ok {
		eventInfo.Message = message
	}
	if source, ok := event["source"].(map[string]interface{}); ok {
		if component, ok := source["component"].(string); ok {
			eventInfo.Source = component
		}
	}

	return []k8s_tool.EventInfo{eventInfo}
}

// setDefaultValues 设置默认值.
func setDefaultValues(orchReq *orchestrator.AnalysisRequest) {
	if orchReq.FailureType == "" {
		orchReq.FailureType = "unknown"
	}
	if orchReq.ResourceType == "" {
		orchReq.ResourceType = "pod"
	}
	if orchReq.ResourceName == "" {
		orchReq.ResourceName = "unknown"
	}
	if orchReq.Namespace == "" {
		orchReq.Namespace = "default"
	}
}

// convertOrchestratorToK8sEventResponse 将 Orchestrator 响应转换为 K8s Event 响应.
func convertOrchestratorToK8sEventResponse(orchResp *orchestrator.AnalysisResponse) K8sEventAnalysisResponse {
	response := K8sEventAnalysisResponse{
		Confidence:      0.0,
		Recommendations: []string{},
	}

	// 转换根因分析
	if orchResp.RootCause != nil {
		response.RootCause = orchResp.RootCause.RootCause
		response.Confidence = orchResp.RootCause.Confidence

		// 从根因推荐中提取建议描述
		if len(orchResp.RootCause.Recommendations) > 0 {
			for _, rec := range orchResp.RootCause.Recommendations {
				response.Recommendations = append(response.Recommendations, rec.Description)
			}
		}
	}

	// 生成 HTML 格式的分析内容
	response.Analysis = formatOrchestratorAnalysis(orchResp)

	return response
}
