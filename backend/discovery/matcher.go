package discovery

import (
	"fmt"
	"strings"
)

type MatchedService struct {
	Id       string
	Name     string
	Status   string
	HttpsUrl *string
}

type Matcher struct{}

func NewMatcher() *Matcher {
	return &Matcher{}
}

func (m *Matcher) Match(containers []Container, routes []Route) []MatchedService {
	var services []MatchedService

	for _, container := range containers {
		var httpsUrl *string
		var url string

		// Primary: IP matching
		for _, route := range routes {
			if route.BackendIp == container.Ip {
				url = fmt.Sprintf("https://%s", route.Domain)
				httpsUrl = &url
				break
			}
		}

		// Fallback: Hostname matching (check both directions)
		if httpsUrl == nil {
			for _, route := range routes {
				subdomain := strings.Split(route.Domain, ".")[0]
				if strings.Contains(route.Domain, container.Name) ||
					strings.Contains(container.Name, subdomain) {
					url = fmt.Sprintf("https://%s", route.Domain)
					httpsUrl = &url
					break
				}
			}
		}

		services = append(services, MatchedService{
			Id:       container.Id,
			Name:     container.Name,
			Status:   container.Status,
			HttpsUrl: httpsUrl,
		})
	}

	return services
}
