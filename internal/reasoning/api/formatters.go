package api

import (
	"fmt"
	"strings"

	"github.com/kart-io/k8s-agent/internal/reasoning/orchestrator"
)

// formatOrchestratorAnalysis 将 Orchestrator 响应格式化为 HTML（兼容旧格式）.
func formatOrchestratorAnalysis(orchResp *orchestrator.AnalysisResponse) string {
	var html string

	// 诊断结果表格
	if orchResp.RootCause != nil {
		html += buildDiagnosisSection(orchResp)
	}

	// 故障描述（如果有）
	if orchResp.Description != nil {
		html += buildDescriptionSection(orchResp)
	}

	// 建议解决方案
	if orchResp.RootCause != nil && len(orchResp.RootCause.Recommendations) > 0 {
		html += buildSolutionsSection(orchResp)
	}

	// 相似案例（新增，旧 API 没有）
	if len(orchResp.SimilarCases) > 0 {
		html += buildSimilarCasesSection(orchResp)
	}

	if html == "" {
		return "<p>无法分析该事件。</p>"
	}

	return html
}

// buildDiagnosisSection 构建诊断结果部分.
func buildDiagnosisSection(orchResp *orchestrator.AnalysisResponse) string {
	var html string
	html += `<div class="diagnosis-section">` + "\n"
	html += `<h3>🔍 诊断结果</h3>` + "\n"
	html += `<table class="diagnosis-table">` + "\n"
	html += `<tbody>` + "\n"

	// 问题类型
	html += `<tr>` + "\n"
	html += `<td class="label-cell">问题类型</td>` + "\n"
	html += fmt.Sprintf(`<td class="value-cell"><span class="type-badge">%s</span></td>`, orchResp.RootCause.Category) + "\n"
	html += `</tr>` + "\n"

	// 置信度
	html += `<tr>` + "\n"
	html += `<td class="label-cell">置信度</td>` + "\n"
	html += fmt.Sprintf(`<td class="value-cell"><span class="confidence-badge">%.0f%%</span></td>`, orchResp.RootCause.Confidence*100) + "\n"
	html += `</tr>` + "\n"

	// 问题描述
	html += `<tr>` + "\n"
	html += `<td class="label-cell">问题描述</td>` + "\n"
	html += fmt.Sprintf(`<td class="value-cell"><div class="problem-desc">%s</div></td>`, orchResp.RootCause.RootCause) + "\n"
	html += `</tr>` + "\n"

	// 推理过程
	if orchResp.RootCause.Reasoning != "" {
		html += `<tr>` + "\n"
		html += `<td class="label-cell">分析推理</td>` + "\n"
		html += fmt.Sprintf(`<td class="value-cell"><div class="reasoning">%s</div></td>`, orchResp.RootCause.Reasoning) + "\n"
		html += `</tr>` + "\n"
	}

	html += `</tbody>` + "\n"
	html += `</table>` + "\n"
	html += `</div>` + "\n"

	return html
}

// buildDescriptionSection 构建故障描述部分.
func buildDescriptionSection(orchResp *orchestrator.AnalysisResponse) string {
	var html string
	html += `<div class="description-section">` + "\n"
	html += `<h3>📝 故障描述</h3>` + "\n"
	html += `<table class="description-table">` + "\n"
	html += `<tbody>` + "\n"

	html += `<tr>` + "\n"
	html += `<td class="label-cell">标题</td>` + "\n"
	html += fmt.Sprintf(`<td class="value-cell">%s</td>`, orchResp.Description.Title) + "\n"
	html += `</tr>` + "\n"

	html += `<tr>` + "\n"
	html += `<td class="label-cell">摘要</td>` + "\n"
	html += fmt.Sprintf(`<td class="value-cell">%s</td>`, orchResp.Description.Summary) + "\n"
	html += `</tr>` + "\n"

	// 影响组件
	if len(orchResp.Description.AffectedComponents) > 0 {
		html += `<tr>` + "\n"
		html += `<td class="label-cell">影响组件</td>` + "\n"
		html += `<td class="value-cell">` + strings.Join(orchResp.Description.AffectedComponents, ", ") + `</td>` + "\n"
		html += `</tr>` + "\n"
	}

	html += `</tbody>` + "\n"
	html += `</table>` + "\n"
	html += `</div>` + "\n"

	return html
}

// buildSolutionsSection 构建建议解决方案部分.
func buildSolutionsSection(orchResp *orchestrator.AnalysisResponse) string {
	var html string
	html += `<div class="solutions-section">` + "\n"
	html += `<h3>💡 建议解决方案</h3>` + "\n"
	html += `<table class="solutions-table">` + "\n"
	html += `<tbody>` + "\n"

	for i, rec := range orchResp.RootCause.Recommendations {
		html += `<tr>` + "\n"
		html += fmt.Sprintf(`<td class="number-cell">%d</td>`, i+1) + "\n"
		html += `<td class="solution-cell">` + "\n"
		html += fmt.Sprintf(`<div class="solution-title">%s</div>`, rec.Action) + "\n"
		html += fmt.Sprintf(`<div class="solution-desc">%s</div>`, rec.Description) + "\n"

		// 命令
		if len(rec.Commands) > 0 {
			html += `<div class="solution-command">` + "\n"
			html += `<div class="command-label">🔧 执行命令:</div>` + "\n"
			html += `<pre class="command-code">` + strings.Join(rec.Commands, "\n") + `</pre>` + "\n"
			html += `</div>` + "\n"
		}

		// 影响和风险级别
		if rec.Impact != "" || rec.RiskLevel != "" {
			html += `<div class="solution-meta">` + "\n"
			if rec.Impact != "" {
				html += fmt.Sprintf(`<span class="meta-tag impact-tag">影响: %s</span>`, rec.Impact) + "\n"
			}
			if rec.RiskLevel != "" {
				html += fmt.Sprintf(`<span class="meta-tag risk-tag risk-%s">风险: %s</span>`, rec.RiskLevel, rec.RiskLevel) + "\n"
			}
			html += `</div>` + "\n"
		}

		html += `</td>` + "\n"
		html += `</tr>` + "\n"
	}

	html += `</tbody>` + "\n"
	html += `</table>` + "\n"
	html += `</div>` + "\n"

	return html
}

// buildSimilarCasesSection 构建相似案例部分.
func buildSimilarCasesSection(orchResp *orchestrator.AnalysisResponse) string {
	var html string
	html += `<div class="similar-cases-section">` + "\n"
	html += `<h3>📚 相似案例</h3>` + "\n"
	html += `<table class="similar-cases-table">` + "\n"
	html += `<tbody>` + "\n"

	for i, sc := range orchResp.SimilarCases {
		html += `<tr>` + "\n"
		html += fmt.Sprintf(`<td class="number-cell">%d</td>`, i+1) + "\n"
		html += `<td class="case-cell">` + "\n"
		html += fmt.Sprintf(`<div class="case-desc">%s</div>`, sc.Description) + "\n"
		html += fmt.Sprintf(`<div class="case-solution">解决方案: %s</div>`, sc.Solution) + "\n"
		html += fmt.Sprintf(`<div class="case-similarity">相似度: %.0f%%</div>`, sc.Similarity*100) + "\n"
		html += `</td>` + "\n"
		html += `</tr>` + "\n"
	}

	html += `</tbody>` + "\n"
	html += `</table>` + "\n"
	html += `</div>` + "\n"

	return html
}
