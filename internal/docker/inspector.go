package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
)

const dockerSock = "/var/run/docker.sock"

// Container holds Traefik-relevant info for a running container.
type Container struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Image         string            `json:"image"`
	State         string            `json:"state"`
	TraefikLabels map[string]string `json:"traefikLabels"`
	Enabled       bool              `json:"enabled"`
}

var sockClient = &http.Client{
	Transport: &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", dockerSock)
		},
	},
}

// Available returns true if the Docker socket is accessible.
func Available() bool {
	resp, err := sockClient.Get("http://docker/v1.41/_ping")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// ListContainers returns all containers with their Traefik labels.
func ListContainers() ([]Container, error) {
	resp, err := sockClient.Get("http://docker/v1.41/containers/json?all=true")
	if err != nil {
		return nil, fmt.Errorf("connecting to docker: %w", err)
	}
	defer resp.Body.Close()

	var raw []struct {
		ID     string            `json:"Id"`
		Names  []string          `json:"Names"`
		Image  string            `json:"Image"`
		State  string            `json:"State"`
		Labels map[string]string `json:"Labels"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decoding containers: %w", err)
	}

	containers := make([]Container, 0, len(raw))
	for _, c := range raw {
		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}

		traefikLabels := map[string]string{}
		for k, v := range c.Labels {
			if strings.HasPrefix(k, "traefik.") {
				traefikLabels[k] = v
			}
		}

		containers = append(containers, Container{
			ID:            c.ID[:func() int { if len(c.ID) < 12 { return len(c.ID) }; return 12 }()],
			Name:          name,
			Image:         c.Image,
			State:         c.State,
			TraefikLabels: traefikLabels,
			Enabled:       c.Labels["traefik.enable"] == "true",
		})
	}
	return containers, nil
}

