package practical

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/kart-io/k8s-agent/pkg/agent/tools"
)

// WebScraperTool scrapes web pages and extracts structured data
type WebScraperTool struct {
	httpClient *http.Client
	maxRetries int
	userAgent  string
}

// NewWebScraperTool creates a new web scraper tool
func NewWebScraperTool() *WebScraperTool {
	return &WebScraperTool{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
		maxRetries: 3,
		userAgent:  "Mozilla/5.0 (compatible; AgentFramework/1.0)",
	}
}

// Name returns the tool name
func (t *WebScraperTool) Name() string {
	return "web_scraper"
}

// Description returns the tool description
func (t *WebScraperTool) Description() string {
	return "Scrapes web pages and extracts structured data including text, links, images, and metadata"
}

// InputSchema returns the input schema
func (t *WebScraperTool) InputSchema() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"url": map[string]interface{}{
				"type":        "string",
				"description": "The URL to scrape",
			},
			"selectors": map[string]interface{}{
				"type":        "object",
				"description": "CSS selectors to extract specific elements",
				"properties": map[string]interface{}{
					"title":   map[string]interface{}{"type": "string"},
					"content": map[string]interface{}{"type": "string"},
					"links":   map[string]interface{}{"type": "string"},
					"images":  map[string]interface{}{"type": "string"},
					"custom":  map[string]interface{}{"type": "object"},
				},
			},
			"extract_metadata": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to extract page metadata",
				"default":     true,
			},
			"max_content_length": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum content length to extract",
				"default":     10000,
			},
		},
		"required": []string{"url"},
	}
}

// OutputSchema returns the output schema
func (t *WebScraperTool) OutputSchema() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"url":      map[string]interface{}{"type": "string"},
			"title":    map[string]interface{}{"type": "string"},
			"content":  map[string]interface{}{"type": "string"},
			"links":    map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
			"images":   map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
			"metadata": map[string]interface{}{"type": "object"},
			"custom":   map[string]interface{}{"type": "object"},
			"error":    map[string]interface{}{"type": "string"},
		},
	}
}

// Execute runs the web scraper
func (t *WebScraperTool) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	params, err := t.parseInput(input)
	if err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	// Validate URL
	parsedURL, err := url.Parse(params.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("only HTTP(S) URLs are supported")
	}

	// Fetch the page
	doc, err := t.fetchPage(ctx, params.URL)
	if err != nil {
		return map[string]interface{}{
			"url":   params.URL,
			"error": err.Error(),
		}, err
	}

	// Extract data
	result := map[string]interface{}{
		"url": params.URL,
	}

	// Extract title
	if params.Selectors.Title != "" {
		result["title"] = doc.Find(params.Selectors.Title).First().Text()
	} else {
		result["title"] = doc.Find("title").First().Text()
	}

	// Extract content
	if params.Selectors.Content != "" {
		content := t.extractText(doc, params.Selectors.Content, params.MaxContentLength)
		result["content"] = content
	} else {
		// Default: extract main content areas
		content := t.extractMainContent(doc, params.MaxContentLength)
		result["content"] = content
	}

	// Extract links
	if params.Selectors.Links != "" {
		result["links"] = t.extractLinks(doc, params.Selectors.Links, parsedURL)
	} else {
		result["links"] = t.extractAllLinks(doc, parsedURL)
	}

	// Extract images
	if params.Selectors.Images != "" {
		result["images"] = t.extractImages(doc, params.Selectors.Images, parsedURL)
	} else {
		result["images"] = t.extractAllImages(doc, parsedURL)
	}

	// Extract metadata
	if params.ExtractMetadata {
		result["metadata"] = t.extractMetadata(doc)
	}

	// Extract custom selectors
	if len(params.Selectors.Custom) > 0 {
		result["custom"] = t.extractCustom(doc, params.Selectors.Custom)
	}

	return result, nil
}

// fetchPage fetches and parses an HTML page
func (t *WebScraperTool) fetchPage(ctx context.Context, url string) (*goquery.Document, error) {
	var lastErr error
	for i := 0; i < t.maxRetries; i++ {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", t.userAgent)

		resp, err := t.httpClient.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		// Parse HTML
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}

		return doc, nil
	}

	return nil, fmt.Errorf("failed after %d retries: %w", t.maxRetries, lastErr)
}

// extractText extracts text from selected elements
func (t *WebScraperTool) extractText(doc *goquery.Document, selector string, maxLength int) string {
	var texts []string
	totalLength := 0

	doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" && totalLength < maxLength {
			remaining := maxLength - totalLength
			if len(text) > remaining {
				text = text[:remaining]
			}
			texts = append(texts, text)
			totalLength += len(text)
		}
	})

	return strings.Join(texts, "\n\n")
}

// extractMainContent extracts main content from common areas
func (t *WebScraperTool) extractMainContent(doc *goquery.Document, maxLength int) string {
	selectors := []string{
		"main", "article", "[role='main']", "#content", ".content",
		"#main", ".main", "body",
	}

	for _, selector := range selectors {
		content := t.extractText(doc, selector, maxLength)
		if len(content) > 100 {
			return content
		}
	}

	// Fallback to body
	return t.extractText(doc, "body", maxLength)
}

// extractLinks extracts links
func (t *WebScraperTool) extractLinks(doc *goquery.Document, selector string, baseURL *url.URL) []string {
	var links []string
	seen := make(map[string]bool)

	doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			absolute := t.makeAbsolute(href, baseURL)
			if !seen[absolute] {
				links = append(links, absolute)
				seen[absolute] = true
			}
		}
	})

	return links
}

// extractAllLinks extracts all links
func (t *WebScraperTool) extractAllLinks(doc *goquery.Document, baseURL *url.URL) []string {
	return t.extractLinks(doc, "a[href]", baseURL)
}

// extractImages extracts image URLs
func (t *WebScraperTool) extractImages(doc *goquery.Document, selector string, baseURL *url.URL) []string {
	var images []string
	seen := make(map[string]bool)

	doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists {
			absolute := t.makeAbsolute(src, baseURL)
			if !seen[absolute] {
				images = append(images, absolute)
				seen[absolute] = true
			}
		}
	})

	return images
}

// extractAllImages extracts all image URLs
func (t *WebScraperTool) extractAllImages(doc *goquery.Document, baseURL *url.URL) []string {
	return t.extractImages(doc, "img[src]", baseURL)
}

// extractMetadata extracts page metadata
func (t *WebScraperTool) extractMetadata(doc *goquery.Document) map[string]interface{} {
	metadata := make(map[string]interface{})

	// Open Graph tags
	og := make(map[string]string)
	doc.Find("meta[property^='og:']").Each(func(i int, s *goquery.Selection) {
		if property, exists := s.Attr("property"); exists {
			if content, exists := s.Attr("content"); exists {
				key := strings.TrimPrefix(property, "og:")
				og[key] = content
			}
		}
	})
	if len(og) > 0 {
		metadata["opengraph"] = og
	}

	// Twitter cards
	twitter := make(map[string]string)
	doc.Find("meta[name^='twitter:']").Each(func(i int, s *goquery.Selection) {
		if name, exists := s.Attr("name"); exists {
			if content, exists := s.Attr("content"); exists {
				key := strings.TrimPrefix(name, "twitter:")
				twitter[key] = content
			}
		}
	})
	if len(twitter) > 0 {
		metadata["twitter"] = twitter
	}

	// Standard meta tags
	doc.Find("meta[name]").Each(func(i int, s *goquery.Selection) {
		if name, exists := s.Attr("name"); exists {
			if content, exists := s.Attr("content"); exists {
				// Skip Twitter tags already processed
				if !strings.HasPrefix(name, "twitter:") {
					metadata[name] = content
				}
			}
		}
	})

	// Canonical URL
	if canonical, exists := doc.Find("link[rel='canonical']").Attr("href"); exists {
		metadata["canonical"] = canonical
	}

	// Language
	if lang, exists := doc.Find("html").Attr("lang"); exists {
		metadata["language"] = lang
	}

	return metadata
}

// extractCustom extracts custom selectors
func (t *WebScraperTool) extractCustom(doc *goquery.Document, selectors map[string]string) map[string]interface{} {
	result := make(map[string]interface{})

	for key, selector := range selectors {
		selection := doc.Find(selector)
		if selection.Length() == 0 {
			result[key] = nil
		} else if selection.Length() == 1 {
			// Single element
			result[key] = map[string]interface{}{
				"text": strings.TrimSpace(selection.Text()),
				"html": t.getOuterHTML(selection),
			}
		} else {
			// Multiple elements
			var elements []map[string]interface{}
			selection.Each(func(i int, s *goquery.Selection) {
				elements = append(elements, map[string]interface{}{
					"text": strings.TrimSpace(s.Text()),
					"html": t.getOuterHTML(s),
				})
			})
			result[key] = elements
		}
	}

	return result
}

// getOuterHTML gets the outer HTML of a selection
func (t *WebScraperTool) getOuterHTML(s *goquery.Selection) string {
	html, _ := s.Html()
	return html
}

// makeAbsolute converts a relative URL to absolute
func (t *WebScraperTool) makeAbsolute(href string, baseURL *url.URL) string {
	u, err := url.Parse(href)
	if err != nil {
		return href
	}
	return baseURL.ResolveReference(u).String()
}

// parseInput parses the tool input
func (t *WebScraperTool) parseInput(input interface{}) (*scraperParams, error) {
	var params scraperParams

	// Handle different input types
	switch v := input.(type) {
	case string:
		params.URL = v
		params.ExtractMetadata = true
		params.MaxContentLength = 10000
	case map[string]interface{}:
		data, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, &params); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported input type: %T", input)
	}

	if params.MaxContentLength == 0 {
		params.MaxContentLength = 10000
	}

	return &params, nil
}

type scraperParams struct {
	URL              string           `json:"url"`
	Selectors        scraperSelectors `json:"selectors"`
	ExtractMetadata  bool             `json:"extract_metadata"`
	MaxContentLength int              `json:"max_content_length"`
}

type scraperSelectors struct {
	Title   string            `json:"title"`
	Content string            `json:"content"`
	Links   string            `json:"links"`
	Images  string            `json:"images"`
	Custom  map[string]string `json:"custom"`
}

// WebScraperRuntimeTool extends WebScraperTool with runtime support
type WebScraperRuntimeTool struct {
	*WebScraperTool
}

// NewWebScraperRuntimeTool creates a runtime-aware web scraper
func NewWebScraperRuntimeTool() *WebScraperRuntimeTool {
	return &WebScraperRuntimeTool{
		WebScraperTool: NewWebScraperTool(),
	}
}

// ExecuteWithRuntime executes with runtime support
func (t *WebScraperRuntimeTool) ExecuteWithRuntime(ctx context.Context, input interface{}, runtime *tools.ToolRuntime) (interface{}, error) {
	// Stream status
	if runtime.StreamWriter != nil {
		runtime.StreamWriter(map[string]interface{}{
			"status": "starting",
			"tool":   t.Name(),
		})
	}

	// Execute the scraping
	result, err := t.Execute(ctx, input)

	// Cache successful results
	if err == nil && runtime != nil {
		params, _ := t.parseInput(input)
		if params != nil {
			// Store in runtime for caching
			cacheKey := fmt.Sprintf("scrape_%s", params.URL)
			runtime.PutToStore([]string{"web_cache"}, cacheKey, result)
		}
	}

	// Stream completion
	if runtime.StreamWriter != nil {
		runtime.StreamWriter(map[string]interface{}{
			"status": "completed",
			"tool":   t.Name(),
			"error":  err,
		})
	}

	return result, err
}
