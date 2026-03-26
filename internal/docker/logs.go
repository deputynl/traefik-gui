package docker

import (
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
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

// ContainerLogLinesSince fetches log lines from a container written after the given time.
func ContainerLogLinesSince(containerName string, since time.Time) ([]string, error) {
	url := fmt.Sprintf(
		"http://docker/v1.41/containers/%s/logs?stdout=true&stderr=true&since=%d",
		containerName, since.Unix(),
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
