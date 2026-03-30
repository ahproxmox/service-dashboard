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

		// Primary: IP matching
		for _, route := range routes {
			if route.BackendIp == container.Ip {
				url := fmt.Sprintf("https://%s", route.Domain)
				httpsUrl = &url
				break
			}
		}

		// Fallback: Hostname matching
		if httpsUrl == nil {
			for _, route := range routes {
				if strings.Contains(route.Domain, container.Name) {
					url := fmt.Sprintf("https://%s", route.Domain)
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
