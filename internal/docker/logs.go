package docker

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ContainerLogLines fetches the last n lines from a container's stdout/stderr.
func ContainerLogLines(containerName string, tail int) ([]string, error) {
	url := fmt.Sprintf(
		"http://docker/v1.41/containers/%s/logs?stdout=true&stderr=true&tail=%d",
		containerName, tail,
	)
	resp, err := sockClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("docker logs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("container %q not found", containerName)
	}

	return readMultiplexed(resp.Body)
}

// StreamContainerLogs streams new log lines from a container into the lines channel.
// Blocks until ctx is cancelled or the container stops.
func StreamContainerLogs(ctx context.Context, containerName string, lines chan<- string) error {
	url := fmt.Sprintf(
		"http://docker/v1.41/containers/%s/logs?stdout=true&stderr=true&follow=true&tail=0",
		containerName,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := sockClient.Do(req)
	if err != nil {
		return fmt.Errorf("docker stream: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("container %q not found", containerName)
	}

	hdr := make([]byte, 8)
	for {
		if _, err := io.ReadFull(resp.Body, hdr); err != nil {
			return err
		}
		size := binary.BigEndian.Uint32(hdr[4:8])
		buf := make([]byte, size)
		if _, err := io.ReadFull(resp.Body, buf); err != nil {
			return err
		}
		for _, line := range strings.Split(strings.TrimRight(string(buf), "\n"), "\n") {
			if line == "" {
				continue
			}
			select {
			case lines <- line:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

// readMultiplexed demultiplexes Docker's log stream and returns all lines.
func readMultiplexed(r io.Reader) ([]string, error) {
	var lines []string
	hdr := make([]byte, 8)
	for {
		if _, err := io.ReadFull(r, hdr); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return lines, nil
			}
			return lines, err
		}
		size := binary.BigEndian.Uint32(hdr[4:8])
		buf := make([]byte, size)
		if _, err := io.ReadFull(r, buf); err != nil {
			return lines, err
		}
		for _, line := range strings.Split(strings.TrimRight(string(buf), "\n"), "\n") {
			if line != "" {
				lines = append(lines, line)
			}
		}
	}
}
