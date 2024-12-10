package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"
)

const (
	FastSpeedTestServerURL = "https://api.fast.com/netflix/speedtest/v2"
	FastBaseURL            = "https://fast.com"
)

type Server struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Location struct {
		City    string `json:"city"`
		Country string `json:"country"`
	} `json:"location"`
}

func (s *Server) Download(requests int, duration time.Duration) (float64, error) {
	var total uint64
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	// Create a default request for downloading the data
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to generate http request: %s", err)
	}

	// Create a channel for tracking downloads
	downloadChannel := make(chan struct{}, requests)

	downloadData := func() {
		clone := req.Clone(ctx)

		// Send the request
		resp, err := http.DefaultClient.Do(clone)
		if err != nil {
			fmt.Printf("failed when making http request: %s", err)
		} else {
			defer resp.Body.Close()

			// Record the data
			n, err := io.Copy(io.Discard, resp.Body)
			if err != nil {
				if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
					fmt.Printf("failed to copy bytes: %s", err)
				}
			}

			atomic.AddUint64(&total, uint64(n))

			// Signal the channel that the download finished
			downloadChannel <- struct{}{}
		}
	}

	// Begin the concurrent downloads
	start := time.Now()
	for i := 0; i < requests; i++ {
		go downloadData()
	}

	// Main loop for orchastrating goroutines
	for {
		select {
		case <-ctx.Done():
			return CalculateMbps(float64(total), time.Since(start).Seconds()), err
		case <-downloadChannel:
			// Begin another download while not timed out
			go downloadData()
		}
	}
}

func (s *Server) Upload(timeout, requests int, payload []byte) ([]float64, error) {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	speeds := make([]float64, 0, requests)

	for i := 0; i < requests; i++ {
		start := time.Now()
		resp, err := client.Post(s.URL, "application/octet-stream", bytes.NewReader(payload))
		if err != nil {
			return nil, fmt.Errorf("upload request %d failed: %w", i+1, err)
		}
		resp.Body.Close()

		speed := CalculateMbps(float64(len(payload)), time.Since(start).Seconds())
		speeds = append(speeds, speed)
	}

	return speeds, nil
}

// Get the IPv4 of the host URL.
func (s *Server) GetIPv4() (string, error) {
	u, err := url.Parse(s.URL)
	if err != nil {
		return "", err
	}

	ips, err := net.LookupIP(u.Host)
	if err != nil {
		return "", err
	}

	return ips[0].String(), nil
}

// Pretty print the JSON object representing the server list.
func DisplayAllServers(servers []Server) error {
	// Create a custom JSON encoder to disable HTML escaping
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)

	// Encode the object with the custom encoder
	err := encoder.Encode(servers)
	if err != nil {
		log.Fatal(err)
	}

	// Pretty-print the JSON contents
	var pretty bytes.Buffer
	err = json.Indent(&pretty, buf.Bytes(), "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	// Print the pretty JSON
	fmt.Println(pretty.String())

	return nil
}
