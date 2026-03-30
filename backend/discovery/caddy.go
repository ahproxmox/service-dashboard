package discovery

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Route struct {
	Domain    string
	BackendIp string
}

type CaddyClient struct {
	apiUrl     string
	httpClient *http.Client
}

func NewCaddyClient(apiUrl string) *CaddyClient {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	return &CaddyClient{
		apiUrl:     apiUrl,
		httpClient: client,
	}
}

func (c *CaddyClient) GetRoutes() ([]Route, error) {
	resp, err := c.httpClient.Get(c.apiUrl + "/config/apps/http/servers/srv0/routes")
	if err != nil {
		return nil, fmt.Errorf("caddy request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("caddy api error: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read caddy response: %w", err)
	}

	// Parse the top-level routes array
	var topRoutes []json.RawMessage
	if err := json.Unmarshal(body, &topRoutes); err != nil {
		return nil, fmt.Errorf("parse caddy routes: %w", err)
	}

	var result []Route
	// Walk the JSON tree looking for host matches paired with reverse_proxy upstreams
	findRoutes(body, &result)
	return result, nil
}

// findRoutes recursively extracts domain→backend mappings from Caddy's nested JSON
func findRoutes(data []byte, result *[]Route) {
	var arr []map[string]json.RawMessage
	if err := json.Unmarshal(data, &arr); err != nil {
		return
	}

	for _, route := range arr {
		// Try to extract a domain from match
		domain := extractDomain(route["match"])

		// Try to extract a backend IP from handle (may be nested in subroutes)
		backendIp := extractBackendIp(route["handle"])

		if domain != "" && backendIp != "" {
			*result = append(*result, Route{Domain: domain, BackendIp: backendIp})
		}

		// Recurse into subroute handlers
		if handleRaw, ok := route["handle"]; ok {
			var handles []map[string]json.RawMessage
			if json.Unmarshal(handleRaw, &handles) == nil {
				for _, h := range handles {
					if routesRaw, ok := h["routes"]; ok {
						findRoutes(routesRaw, result)
					}
				}
			}
		}
	}
}

func extractDomain(matchRaw json.RawMessage) string {
	var matches []map[string]json.RawMessage
	if json.Unmarshal(matchRaw, &matches) != nil || len(matches) == 0 {
		return ""
	}
	var hosts []string
	if json.Unmarshal(matches[0]["host"], &hosts) != nil || len(hosts) == 0 {
		return ""
	}
	// Skip wildcard matches
	if strings.Contains(hosts[0], "*") {
		return ""
	}
	return hosts[0]
}

func extractBackendIp(handleRaw json.RawMessage) string {
	var handles []map[string]json.RawMessage
	if json.Unmarshal(handleRaw, &handles) != nil {
		return ""
	}
	for _, h := range handles {
		// Direct reverse_proxy handler
		if upRaw, ok := h["upstreams"]; ok {
			var upstreams []map[string]string
			if json.Unmarshal(upRaw, &upstreams) == nil && len(upstreams) > 0 {
				if dial, ok := upstreams[0]["dial"]; ok {
					return strings.Split(dial, ":")[0]
				}
			}
		}
		// Nested subroute — dig into its routes
		if routesRaw, ok := h["routes"]; ok {
			var subRoutes []map[string]json.RawMessage
			if json.Unmarshal(routesRaw, &subRoutes) == nil {
				for _, sr := range subRoutes {
					if ip := extractBackendIp(sr["handle"]); ip != "" {
						return ip
					}
				}
			}
		}
	}
	return ""
}
