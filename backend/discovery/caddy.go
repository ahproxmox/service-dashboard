package discovery

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Route struct {
	Domain    string
	BackendIp string
}

type CaddyClient struct {
	apiUrl string
}

func NewCaddyClient(apiUrl string) *CaddyClient {
	return &CaddyClient{apiUrl: apiUrl}
}

func (c *CaddyClient) GetRoutes() ([]Route, error) {
	resp, err := http.Get(c.apiUrl + "/admin/api/config/apps/http/servers/default/routes")
	if err != nil {
		return nil, fmt.Errorf("caddy request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Parse the Caddy routes array
	var routes []interface{}
	if err := json.Unmarshal(body, &routes); err != nil {
		return nil, fmt.Errorf("parse caddy routes: %w", err)
	}

	var result []Route
	for _, routeData := range routes {
		route := routeData.(map[string]interface{})

		// Extract domain from match
		var domain string
		if matches, ok := route["match"].([]interface{}); ok && len(matches) > 0 {
			match := matches[0].(map[string]interface{})
			if hosts, ok := match["host"].([]interface{}); ok && len(hosts) > 0 {
				domain = hosts[0].(string)
			}
		}

		// Extract backend IP from handle
		var backendIp string
		if handles, ok := route["handle"].([]interface{}); ok && len(handles) > 0 {
			handle := handles[0].(map[string]interface{})
			if upstreams, ok := handle["upstreams"].([]interface{}); ok && len(upstreams) > 0 {
				upstream := upstreams[0].(map[string]interface{})
				if dial, ok := upstream["dial"].(string); ok {
					// dial is "IP:port", extract IP
					backendIp = strings.Split(dial, ":")[0]
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
