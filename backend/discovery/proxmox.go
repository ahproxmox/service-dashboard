package discovery

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Container struct {
	Id     string
	Name   string
	Status string // "running", "stopped"
	Ip     string
}

type proxmoxContainer struct {
	Vmid     int    `json:"vmid"`
	Hostname string `json:"name"`
	Status   string `json:"status"`
	Ip       string `json:"ip"`
}

type proxmoxResponse struct {
	Data []proxmoxContainer `json:"data"`
}

type ProxmoxClient struct {
	apiUrl      string
	tokenId     string
	tokenSecret string
	httpClient  *http.Client
}

func NewProxmoxClient(apiUrl, tokenId, tokenSecret string) *ProxmoxClient {
	// Ignore self-signed certs for homelab
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	return &ProxmoxClient{
		apiUrl:      apiUrl,
		tokenId:     tokenId,
		tokenSecret: tokenSecret,
		httpClient:  client,
	}
}

func (p *ProxmoxClient) GetContainers() ([]Container, error) {
	req, _ := http.NewRequest("GET", p.apiUrl+"/api2/json/nodes/pve/lxc", nil)
	req.Header.Set("Authorization", fmt.Sprintf("PVEAPIToken=%s=%s", p.tokenId, p.tokenSecret))

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("proxmox request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var proxmoxResp proxmoxResponse
	if err := json.Unmarshal(body, &proxmoxResp); err != nil {
		return nil, fmt.Errorf("parse proxmox response: %w", err)
	}

	var containers []Container
	for _, pc := range proxmoxResp.Data {
		containers = append(containers, Container{
			Id:     fmt.Sprintf("%d", pc.Vmid),
			Name:   pc.Hostname,
			Status: pc.Status,
			Ip:     pc.Ip,
		})
	}

	return containers, nil
}
