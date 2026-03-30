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
	resp, err := c.httpClient.Get(c.apiUrl + "/admin/api/config/apps/http/servers/default/routes")
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

	// Parse the Caddy routes array
	var routes []interface{}
	if err := json.Unmarshal(body, &routes); err != nil {
		return nil, fmt.Errorf("parse caddy routes: %w", err)
	}

	var result []Route
	for _, routeData := range routes {
		// Safe type assertion with ok check
		route, ok := routeData.(map[string]interface{})
		if !ok {
			continue // Skip invalid route
		}

		// Extract domain from match
		var domain string
		if matches, ok := route["match"].([]interface{}); ok && len(matches) > 0 {
			if match, ok := matches[0].(map[string]interface{}); ok {
				if hosts, ok := match["host"].([]interface{}); ok && len(hosts) > 0 {
					if host, ok := hosts[0].(string); ok {
						domain = host
					}
				}
			}
		}

		// Extract backend IP from handle
		var backendIp string
		if handles, ok := route["handle"].([]interface{}); ok && len(handles) > 0 {
			if handle, ok := handles[0].(map[string]interface{}); ok {
				if upstreams, ok := handle["upstreams"].([]interface{}); ok && len(upstreams) > 0 {
					if upstream, ok := upstreams[0].(map[string]interface{}); ok {
						if dial, ok := upstream["dial"].(string); ok {
							// dial is "IP:port", extract IP
							backendIp = strings.Split(dial, ":")[0]
						}
					}
				}
			}
		}

		if domain != "" && backendIp != "" {
			result = append(result, Route{
				Domain:    domain,
				BackendIp: backendIp,
			})
		}
	}

	return result, nil
}
