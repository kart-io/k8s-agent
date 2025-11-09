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
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"pageSize" form:"pageSize"`
	Sort     string `json:"sort" form:"sort"`   // 排序字段
	Order    string `json:"order" form:"order"` // asc 或 desc
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
	sort := c.DefaultQuery("sort", "created_at")
	order := c.DefaultQuery("order", "desc")

	// Validate order
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	return &Params{
		Page:     page,
		PageSize: pageSize,
		Sort:     sort,
		Order:    order,
	}
}

// ParseWithDefaults 解析分页参数（自定义默认值）
func ParseWithDefaults(c *gin.Context, defaultPage, defaultPageSize int) *Params {
	page, _ := strconv.Atoi(c.DefaultQuery("page", strconv.Itoa(defaultPage)))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", strconv.Itoa(defaultPageSize)))
	sort := c.DefaultQuery("sort", "created_at")
	order := c.DefaultQuery("order", "desc")

	// Validate order
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	return &Params{
		Page:     page,
		PageSize: pageSize,
		Sort:     sort,
		Order:    order,
	}
}

// Response 分页响应结构
type Response struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	TotalPages int         `json:"totalPages"` // 总页数
}

// NewResponse 创建分页响应
func NewResponse(items interface{}, total int64, params *Params) *Response {
	return &Response{
		Items:      items,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.GetPageSize(),
		TotalPages: CalculateTotalPages(total, params.GetPageSize()),
	}
}

// CalculateTotalPages 计算总页数
func CalculateTotalPages(total int64, pageSize int) int {
	if pageSize <= 0 {
		return 0
	}
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}
	return totalPages
}

// BuildOrderBy 构建 ORDER BY SQL 子句
// allowedFields: 允许排序的字段映射（key: 请求字段名, value: 数据库字段名）
func BuildOrderBy(params *Params, allowedFields map[string]string) string {
	// 检查排序字段是否允许
	field, ok := allowedFields[params.Sort]
	if !ok {
		// 如果字段不允许，默认使用 created_at
		field = "created_at"
	}

	order := "DESC"
	if params.Order == "asc" {
		order = "ASC"
	}

	return field + " " + order
}
