package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	// DefaultPage 默认页码
	DefaultPage = 1
	// DefaultPageSize 默认每页数量
	DefaultPageSize = 10
	// MaxPageSize 最大每页数量
	MaxPageSize = 100
)

// Params 分页参数
type Params struct {
	Page     int `json:"page" form:"page"`
	PageSize int `json:"pageSize" form:"pageSize"`
}

// GetOffset 计算偏移量
func (p *Params) GetOffset() int {
	if p.Page <= 0 {
		p.Page = DefaultPage
	}
	return (p.Page - 1) * p.GetPageSize()
}

// GetPageSize 获取每页数量
func (p *Params) GetPageSize() int {
	if p.PageSize <= 0 {
		return DefaultPageSize
	}
	if p.PageSize > MaxPageSize {
		return MaxPageSize
	}
	return p.PageSize
}

// GetLimit 获取限制数量
func (p *Params) GetLimit() int {
	return p.GetPageSize()
}

// Parse 从 Gin Context 中解析分页参数
func Parse(c *gin.Context) *Params {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	return &Params{
		Page:     page,
		PageSize: pageSize,
	}
}

// ParseWithDefaults 解析分页参数（自定义默认值）
func ParseWithDefaults(c *gin.Context, defaultPage, defaultPageSize int) *Params {
	page, _ := strconv.Atoi(c.DefaultQuery("page", strconv.Itoa(defaultPage)))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", strconv.Itoa(defaultPageSize)))

	return &Params{
		Page:     page,
		PageSize: pageSize,
	}
}

// Response 分页响应结构
type Response struct {
	Items    interface{} `json:"items"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

// NewResponse 创建分页响应
func NewResponse(items interface{}, total int64, params *Params) *Response {
	return &Response{
		Items:    items,
		Total:    total,
		Page:     params.Page,
		PageSize: params.GetPageSize(),
	}
}
