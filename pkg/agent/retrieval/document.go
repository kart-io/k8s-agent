package retrieval

import (
	"fmt"
	"sort"
	"strings"
)

// Document 文档结构
//
// 表示检索系统中的基本文档单元，包含内容和元数据
type Document struct {
	// PageContent 文档的文本内容
	PageContent string

	// Metadata 文档的元数据
	Metadata map[string]interface{}

	// ID 文档的唯一标识符
	ID string

	// Score 检索得分（由检索器设置）
	Score float64
}

// NewDocument 创建新文档
func NewDocument(content string, metadata map[string]interface{}) *Document {
	return &Document{
		PageContent: content,
		Metadata:    metadata,
		ID:          generateID(),
	}
}

// NewDocumentWithID 创建带 ID 的文档
func NewDocumentWithID(id, content string, metadata map[string]interface{}) *Document {
	return &Document{
		ID:          id,
		PageContent: content,
		Metadata:    metadata,
	}
}

// Clone 克隆文档
func (d *Document) Clone() *Document {
	metadata := make(map[string]interface{})
	for k, v := range d.Metadata {
		metadata[k] = v
	}

	return &Document{
		ID:          d.ID,
		PageContent: d.PageContent,
		Metadata:    metadata,
		Score:       d.Score,
	}
}

// GetMetadata 获取元数据字段
func (d *Document) GetMetadata(key string) (interface{}, bool) {
	if d.Metadata == nil {
		return nil, false
	}
	val, ok := d.Metadata[key]
	return val, ok
}

// SetMetadata 设置元数据字段
func (d *Document) SetMetadata(key string, value interface{}) {
	if d.Metadata == nil {
		d.Metadata = make(map[string]interface{})
	}
	d.Metadata[key] = value
}

// String 返回文档的字符串表示
func (d *Document) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Document(ID=%s, Score=%.4f", d.ID, d.Score))
	if len(d.Metadata) > 0 {
		sb.WriteString(", Metadata={")
		keys := make([]string, 0, len(d.Metadata))
		for k := range d.Metadata {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for i, k := range keys {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%s: %v", k, d.Metadata[k]))
		}
		sb.WriteString("}")
	}
	sb.WriteString(fmt.Sprintf(", Content='%s')", truncateString(d.PageContent, 50)))
	return sb.String()
}

// DocumentCollection 文档集合
type DocumentCollection []*Document

// Len 实现 sort.Interface
func (dc DocumentCollection) Len() int {
	return len(dc)
}

// Less 按分数降序排序
func (dc DocumentCollection) Less(i, j int) bool {
	return dc[i].Score > dc[j].Score
}

// Swap 交换元素
func (dc DocumentCollection) Swap(i, j int) {
	dc[i], dc[j] = dc[j], dc[i]
}

// SortByScore 按分数排序
func (dc DocumentCollection) SortByScore() {
	sort.Sort(dc)
}

// Top 获取前 N 个文档
func (dc DocumentCollection) Top(n int) DocumentCollection {
	if n >= len(dc) {
		return dc
	}
	return dc[:n]
}

// Filter 过滤文档
func (dc DocumentCollection) Filter(predicate func(*Document) bool) DocumentCollection {
	result := make(DocumentCollection, 0)
	for _, doc := range dc {
		if predicate(doc) {
			result = append(result, doc)
		}
	}
	return result
}

// Map 映射文档
func (dc DocumentCollection) Map(mapper func(*Document) *Document) DocumentCollection {
	result := make(DocumentCollection, len(dc))
	for i, doc := range dc {
		result[i] = mapper(doc)
	}
	return result
}

// Deduplicate 去重（基于 ID）
func (dc DocumentCollection) Deduplicate() DocumentCollection {
	seen := make(map[string]bool)
	result := make(DocumentCollection, 0)

	for _, doc := range dc {
		if !seen[doc.ID] {
			seen[doc.ID] = true
			result = append(result, doc)
		}
	}

	return result
}

// 辅助函数

var idCounter int64

// generateID 生成唯一 ID
func generateID() string {
	idCounter++
	return fmt.Sprintf("doc_%d", idCounter)
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
