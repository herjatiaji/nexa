package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// WebTool provides web search, web page fetching, and real-time weather capabilities.
type WebTool struct {
	client *http.Client
}

// New creates a new WebTool instance.
func New() *WebTool {
	return &WebTool{
		client: &http.Client{
			Timeout: 12 * time.Second,
		},
	}
}

func (w *WebTool) Name() string {
	return "web"
}

func (w *WebTool) Description() string {
	return "Live web search, web page fetching, and weather capabilities. Supported actions: " +
		"'search' (search the live internet for news, information, or answers), " +
		"'fetch' (read text content from a web page URL), " +
		"'weather' (get current weather and forecast for any city or location)."
}

func (w *WebTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Web action to perform: 'search', 'fetch', or 'weather'",
				"enum":        []interface{}{"search", "fetch", "weather"},
			},
			"query": map[string]interface{}{
				"type":        "string",
				"description": "Search query keywords (required for 'search' action)",
			},
			"url": map[string]interface{}{
				"type":        "string",
				"description": "Target web page URL (required for 'fetch' action)",
			},
			"location": map[string]interface{}{
				"type":        "string",
				"description": "City or country name e.g. 'Jakarta', 'Tokyo', 'London' (required for 'weather' action)",
			},
		},
		"required": []interface{}{"action"},
	}
}

type webInput struct {
	Action   string `json:"action"`
	Query    string `json:"query,omitempty"`
	URL      string `json:"url,omitempty"`
	Location string `json:"location,omitempty"`
}

func (w *WebTool) Execute(input string) (string, error) {
	var params webInput
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	switch params.Action {
	case "search":
		return w.searchWeb(params.Query)
	case "fetch":
		return w.fetchURL(params.URL)
	case "weather":
		return w.getWeather(params.Location)
	default:
		return "", fmt.Errorf("unknown action: %s", params.Action)
	}
}

func (w *WebTool) searchWeb(query string) (string, error) {
	if query == "" {
		return "", fmt.Errorf("query parameter is required for search action")
	}

	postData := url.Values{}
	postData.Set("q", query)

	req, err := http.NewRequest("POST", "https://html.duckduckgo.com/html/", strings.NewReader(postData.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create search request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := w.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("web search network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("search engine HTTP error status: %s", resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read search response: %w", err)
	}

	htmlText := string(bodyBytes)

	reTitle := regexp.MustCompile(`<a[^>]*class="result__a"[^>]*>([\s\S]*?)</a>`)
	reSnippet := regexp.MustCompile(`<a[^>]*class="result__snippet"[^>]*>([\s\S]*?)</a>`)

	titles := reTitle.FindAllStringSubmatch(htmlText, 6)
	snippets := reSnippet.FindAllStringSubmatch(htmlText, 6)

	var results []string
	count := len(snippets)
	if len(titles) < count {
		count = len(titles)
	}

	for i := 0; i < count; i++ {
		t := stripHTML(titles[i][1])
		s := stripHTML(snippets[i][1])
		if t != "" && s != "" {
			results = append(results, fmt.Sprintf("%d. %s\n   %s", i+1, t, s))
		}
	}

	if len(results) == 0 {
		return fmt.Sprintf("No live search results found online for %q.", query), nil
	}

	return fmt.Sprintf("🌐 Live Web Search Results for %q:\n\n%s", query, strings.Join(results, "\n\n")), nil
}

func (w *WebTool) fetchURL(targetURL string) (string, error) {
	if targetURL == "" {
		return "", fmt.Errorf("url parameter is required for fetch action")
	}

	if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
		targetURL = "https://" + targetURL
	}

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := w.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("web page returned HTTP status: %s", resp.Status)
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 500000))
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	cleanText := stripHTML(string(bodyBytes))
	cleanText = regexp.MustCompile(`\n\s*\n`).ReplaceAllString(cleanText, "\n\n")

	const maxLen = 3000
	if len(cleanText) > maxLen {
		cleanText = cleanText[:maxLen] + "\n\n... [content truncated]"
	}

	if strings.TrimSpace(cleanText) == "" {
		return fmt.Sprintf("Fetched %s but no readable text content was found.", targetURL), nil
	}

	return fmt.Sprintf("📄 Content from %s:\n\n%s", targetURL, cleanText), nil
}

func (w *WebTool) getWeather(location string) (string, error) {
	if location == "" {
		location = "auto"
	}

	// Use wttr.in weather API service
	weatherURL := fmt.Sprintf("https://wttr.in/%s?0&m&T", url.QueryEscape(location))
	req, _ := http.NewRequest("GET", weatherURL, nil)
	req.Header.Set("User-Agent", "curl/7.68.0")

	resp, err := w.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch weather: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read weather response: %w", err)
	}

	weatherText := string(bodyBytes)
	// Strip ANSI escape codes
	reANSI := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	weatherText = reANSI.ReplaceAllString(weatherText, "")
	weatherText = strings.TrimSpace(weatherText)

	if len(weatherText) == 0 {
		return fmt.Sprintf("Unable to retrieve weather data for %s.", location), nil
	}

	return fmt.Sprintf("🌤️ Weather Report for %s:\n\n%s", location, weatherText), nil
}

func stripHTML(input string) string {
	// Remove script and style elements
	reScript := regexp.MustCompile(`(?i)<script[\s\S]*?</script>`)
	input = reScript.ReplaceAllString(input, "")
	reStyle := regexp.MustCompile(`(?i)<style[\s\S]*?</style>`)
	input = reStyle.ReplaceAllString(input, "")

	// Replace HTML tags with spaces
	reTag := regexp.MustCompile(`<[^>]*>`)
	input = reTag.ReplaceAllString(input, " ")

	// Decode common HTML entities
	input = strings.ReplaceAll(input, "&nbsp;", " ")
	input = strings.ReplaceAll(input, "&amp;", "&")
	input = strings.ReplaceAll(input, "&lt;", "<")
	input = strings.ReplaceAll(input, "&gt;", ">")
	input = strings.ReplaceAll(input, "&quot;", "\"")
	input = strings.ReplaceAll(input, "&#39;", "'")

	// Collapse spaces
	reSpace := regexp.MustCompile(`[ \t\r\n]+`)
	input = reSpace.ReplaceAllString(input, " ")

	return strings.TrimSpace(input)
}
