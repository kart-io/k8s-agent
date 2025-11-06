package pagination

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestContext(queryParams url.Values) *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create request with query parameters
	req, _ := http.NewRequest("GET", "/?"+queryParams.Encode(), nil)
	c.Request = req

	return c
}

// TestParamsGetOffset tests the GetOffset method
func TestParamsGetOffset(t *testing.T) {
	tests := []struct {
		name           string
		page           int
		pageSize       int
		expectedOffset int
	}{
		{
			name:           "first page",
			page:           1,
			pageSize:       10,
			expectedOffset: 0,
		},
		{
			name:           "second page",
			page:           2,
			pageSize:       10,
			expectedOffset: 10,
		},
		{
			name:           "third page with custom page size",
			page:           3,
			pageSize:       20,
			expectedOffset: 40,
		},
		{
			name:           "zero page defaults to first page",
			page:           0,
			pageSize:       10,
			expectedOffset: 0,
		},
		{
			name:           "negative page defaults to first page",
			page:           -1,
			pageSize:       10,
			expectedOffset: 0,
		},
		{
			name:           "large page number",
			page:           100,
			pageSize:       10,
			expectedOffset: 990,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Params{
				Page:     tt.page,
				PageSize: tt.pageSize,
			}
			assert.Equal(t, tt.expectedOffset, p.GetOffset())
		})
	}
}

// TestParamsGetPageSize tests the GetPageSize method
func TestParamsGetPageSize(t *testing.T) {
	tests := []struct {
		name             string
		pageSize         int
		expectedPageSize int
	}{
		{
			name:             "valid page size",
			pageSize:         10,
			expectedPageSize: 10,
		},
		{
			name:             "zero page size defaults to default",
			pageSize:         0,
			expectedPageSize: DefaultPageSize,
		},
		{
			name:             "negative page size defaults to default",
			pageSize:         -1,
			expectedPageSize: DefaultPageSize,
		},
		{
			name:             "page size exceeds maximum",
			pageSize:         150,
			expectedPageSize: MaxPageSize,
		},
		{
			name:             "page size equals maximum",
			pageSize:         MaxPageSize,
			expectedPageSize: MaxPageSize,
		},
		{
			name:             "page size just below maximum",
			pageSize:         99,
			expectedPageSize: 99,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Params{
				PageSize: tt.pageSize,
			}
			assert.Equal(t, tt.expectedPageSize, p.GetPageSize())
		})
	}
}

// TestParamsGetLimit tests the GetLimit method
func TestParamsGetLimit(t *testing.T) {
	tests := []struct {
		name          string
		pageSize      int
		expectedLimit int
	}{
		{
			name:          "normal page size",
			pageSize:      20,
			expectedLimit: 20,
		},
		{
			name:          "zero page size",
			pageSize:      0,
			expectedLimit: DefaultPageSize,
		},
		{
			name:          "exceeds max",
			pageSize:      200,
			expectedLimit: MaxPageSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Params{
				PageSize: tt.pageSize,
			}
			assert.Equal(t, tt.expectedLimit, p.GetLimit())
		})
	}
}

// TestParse tests the Parse function
func TestParse(t *testing.T) {
	tests := []struct {
		name         string
		queryParams  url.Values
		expectedPage int
		expectedSize int
		expectedSort string
		expectedOrder string
	}{
		{
			name: "default values",
			queryParams: url.Values{},
			expectedPage: 1,
			expectedSize: 10,
			expectedSort: "created_at",
			expectedOrder: "desc",
		},
		{
			name: "custom values",
			queryParams: url.Values{
				"page":     []string{"3"},
				"pageSize": []string{"25"},
				"sort":     []string{"name"},
				"order":    []string{"asc"},
			},
			expectedPage: 3,
			expectedSize: 25,
			expectedSort: "name",
			expectedOrder: "asc",
		},
		{
			name: "invalid order defaults to desc",
			queryParams: url.Values{
				"page":     []string{"1"},
				"pageSize": []string{"10"},
				"sort":     []string{"id"},
				"order":    []string{"invalid"},
			},
			expectedPage: 1,
			expectedSize: 10,
			expectedSort: "id",
			expectedOrder: "desc",
		},
		{
			name: "invalid page number",
			queryParams: url.Values{
				"page":     []string{"invalid"},
				"pageSize": []string{"20"},
			},
			expectedPage: 0, // strconv.Atoi returns 0 for invalid input
			expectedSize: 20,
			expectedSort: "created_at",
			expectedOrder: "desc",
		},
		{
			name: "invalid page size",
			queryParams: url.Values{
				"page":     []string{"2"},
				"pageSize": []string{"invalid"},
			},
			expectedPage: 2,
			expectedSize: 0, // strconv.Atoi returns 0 for invalid input
			expectedSort: "created_at",
			expectedOrder: "desc",
		},
		{
			name: "order is asc",
			queryParams: url.Values{
				"order": []string{"asc"},
			},
			expectedPage: 1,
			expectedSize: 10,
			expectedSort: "created_at",
			expectedOrder: "asc",
		},
		{
			name: "order is desc",
			queryParams: url.Values{
				"order": []string{"desc"},
			},
			expectedPage: 1,
			expectedSize: 10,
			expectedSort: "created_at",
			expectedOrder: "desc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := setupTestContext(tt.queryParams)
			params := Parse(c)

			assert.Equal(t, tt.expectedPage, params.Page)
			assert.Equal(t, tt.expectedSize, params.PageSize)
			assert.Equal(t, tt.expectedSort, params.Sort)
			assert.Equal(t, tt.expectedOrder, params.Order)
		})
	}
}

// TestParseWithDefaults tests the ParseWithDefaults function
func TestParseWithDefaults(t *testing.T) {
	tests := []struct {
		name             string
		queryParams      url.Values
		defaultPage      int
		defaultPageSize  int
		expectedPage     int
		expectedSize     int
	}{
		{
			name:            "use custom defaults when no query params",
			queryParams:     url.Values{},
			defaultPage:     5,
			defaultPageSize: 50,
			expectedPage:    5,
			expectedSize:    50,
		},
		{
			name: "query params override defaults",
			queryParams: url.Values{
				"page":     []string{"2"},
				"pageSize": []string{"30"},
			},
			defaultPage:     5,
			defaultPageSize: 50,
			expectedPage:    2,
			expectedSize:    30,
		},
		{
			name:            "zero defaults",
			queryParams:     url.Values{},
			defaultPage:     0,
			defaultPageSize: 0,
			expectedPage:    0,
			expectedSize:    0,
		},
		{
			name: "partial override",
			queryParams: url.Values{
				"page": []string{"10"},
			},
			defaultPage:     1,
			defaultPageSize: 20,
			expectedPage:    10,
			expectedSize:    20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := setupTestContext(tt.queryParams)
			params := ParseWithDefaults(c, tt.defaultPage, tt.defaultPageSize)

			assert.Equal(t, tt.expectedPage, params.Page)
			assert.Equal(t, tt.expectedSize, params.PageSize)
		})
	}
}

// TestNewResponse tests the NewResponse function
func TestNewResponse(t *testing.T) {
	tests := []struct {
		name           string
		items          interface{}
		total          int64
		page           int
		pageSize       int
		expectedPages  int
	}{
		{
			name:          "normal pagination",
			items:         []int{1, 2, 3},
			total:         100,
			page:          1,
			pageSize:      10,
			expectedPages: 10,
		},
		{
			name:          "last page with remainder",
			items:         []int{1, 2},
			total:         22,
			page:          3,
			pageSize:      10,
			expectedPages: 3,
		},
		{
			name:          "zero items",
			items:         []int{},
			total:         0,
			page:          1,
			pageSize:      10,
			expectedPages: 0,
		},
		{
			name:          "single page",
			items:         []int{1, 2, 3},
			total:         3,
			page:          1,
			pageSize:      10,
			expectedPages: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := &Params{
				Page:     tt.page,
				PageSize: tt.pageSize,
			}
			resp := NewResponse(tt.items, tt.total, params)

			assert.Equal(t, tt.items, resp.Items)
			assert.Equal(t, tt.total, resp.Total)
			assert.Equal(t, tt.page, resp.Page)
			assert.Equal(t, tt.pageSize, resp.PageSize)
			assert.Equal(t, tt.expectedPages, resp.TotalPages)
		})
	}
}

// TestCalculateTotalPages tests the CalculateTotalPages function
func TestCalculateTotalPages(t *testing.T) {
	tests := []struct {
		name          string
		total         int64
		pageSize      int
		expectedPages int
	}{
		{
			name:          "exact division",
			total:         100,
			pageSize:      10,
			expectedPages: 10,
		},
		{
			name:          "with remainder",
			total:         105,
			pageSize:      10,
			expectedPages: 11,
		},
		{
			name:          "zero total",
			total:         0,
			pageSize:      10,
			expectedPages: 0,
		},
		{
			name:          "zero page size",
			total:         100,
			pageSize:      0,
			expectedPages: 0,
		},
		{
			name:          "negative page size",
			total:         100,
			pageSize:      -10,
			expectedPages: 0,
		},
		{
			name:          "total less than page size",
			total:         5,
			pageSize:      10,
			expectedPages: 1,
		},
		{
			name:          "total equals page size",
			total:         10,
			pageSize:      10,
			expectedPages: 1,
		},
		{
			name:          "single item",
			total:         1,
			pageSize:      10,
			expectedPages: 1,
		},
		{
			name:          "large numbers",
			total:         1000000,
			pageSize:      100,
			expectedPages: 10000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateTotalPages(tt.total, tt.pageSize)
			assert.Equal(t, tt.expectedPages, result)
		})
	}
}

// TestBuildOrderBy tests the BuildOrderBy function
func TestBuildOrderBy(t *testing.T) {
	tests := []struct {
		name          string
		params        *Params
		allowedFields map[string]string
		expected      string
	}{
		{
			name: "allowed field with asc order",
			params: &Params{
				Sort:  "name",
				Order: "asc",
			},
			allowedFields: map[string]string{
				"name": "user_name",
				"date": "created_date",
			},
			expected: "user_name ASC",
		},
		{
			name: "allowed field with desc order",
			params: &Params{
				Sort:  "name",
				Order: "desc",
			},
			allowedFields: map[string]string{
				"name": "user_name",
				"date": "created_date",
			},
			expected: "user_name DESC",
		},
		{
			name: "disallowed field defaults to created_at",
			params: &Params{
				Sort:  "invalid_field",
				Order: "asc",
			},
			allowedFields: map[string]string{
				"name": "user_name",
				"date": "created_date",
			},
			expected: "created_at ASC",
		},
		{
			name: "empty sort field defaults to created_at",
			params: &Params{
				Sort:  "",
				Order: "desc",
			},
			allowedFields: map[string]string{
				"name": "user_name",
			},
			expected: "created_at DESC",
		},
		{
			name: "invalid order defaults to DESC",
			params: &Params{
				Sort:  "name",
				Order: "invalid",
			},
			allowedFields: map[string]string{
				"name": "user_name",
			},
			expected: "user_name DESC",
		},
		{
			name: "empty allowedFields map",
			params: &Params{
				Sort:  "name",
				Order: "asc",
			},
			allowedFields: map[string]string{},
			expected:      "created_at ASC",
		},
		{
			name: "nil allowedFields map",
			params: &Params{
				Sort:  "name",
				Order: "asc",
			},
			allowedFields: nil,
			expected:      "created_at ASC",
		},
		{
			name: "field mapping preserves database column name",
			params: &Params{
				Sort:  "id",
				Order: "desc",
			},
			allowedFields: map[string]string{
				"id":   "user_id",
				"name": "full_name",
			},
			expected: "user_id DESC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildOrderBy(tt.params, tt.allowedFields)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestResponseStructure tests the Response structure
func TestResponseStructure(t *testing.T) {
	items := []string{"item1", "item2", "item3"}
	resp := Response{
		Items:      items,
		Total:      100,
		Page:       2,
		PageSize:   10,
		TotalPages: 10,
	}

	assert.Equal(t, items, resp.Items)
	assert.Equal(t, int64(100), resp.Total)
	assert.Equal(t, 2, resp.Page)
	assert.Equal(t, 10, resp.PageSize)
	assert.Equal(t, 10, resp.TotalPages)
}

// TestConstants tests pagination constants
func TestConstants(t *testing.T) {
	assert.Equal(t, 1, DefaultPage)
	assert.Equal(t, 10, DefaultPageSize)
	assert.Equal(t, 100, MaxPageSize)
}

// TestParamsValidation tests edge cases for Params validation
func TestParamsValidation(t *testing.T) {
	tests := []struct {
		name           string
		params         *Params
		expectedOffset int
		expectedLimit  int
	}{
		{
			name: "negative page and page size",
			params: &Params{
				Page:     -5,
				PageSize: -10,
			},
			expectedOffset: 0,
			expectedLimit:  DefaultPageSize,
		},
		{
			name: "very large page number",
			params: &Params{
				Page:     999999,
				PageSize: 10,
			},
			expectedOffset: 9999980,
			expectedLimit:  10,
		},
		{
			name: "page size exceeds max by large margin",
			params: &Params{
				Page:     1,
				PageSize: 10000,
			},
			expectedOffset: 0,
			expectedLimit:  MaxPageSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedOffset, tt.params.GetOffset())
			assert.Equal(t, tt.expectedLimit, tt.params.GetLimit())
		})
	}
}

// TestIntegrationParseAndResponse tests the integration of Parse and NewResponse
func TestIntegrationParseAndResponse(t *testing.T) {
	queryParams := url.Values{
		"page":     []string{"2"},
		"pageSize": []string{"20"},
		"sort":     []string{"name"},
		"order":    []string{"asc"},
	}
	c := setupTestContext(queryParams)

	params := Parse(c)
	items := []string{"item1", "item2"}
	total := int64(100)

	resp := NewResponse(items, total, params)

	assert.Equal(t, items, resp.Items)
	assert.Equal(t, total, resp.Total)
	assert.Equal(t, 2, resp.Page)
	assert.Equal(t, 20, resp.PageSize)
	assert.Equal(t, 5, resp.TotalPages)
}
