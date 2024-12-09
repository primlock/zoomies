package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
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

func (s *Server) Download(timeout, requests int) ([]float64, error) {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	speeds := make([]float64, 0, requests)

	for i := 0; i < requests; i++ {
		start := time.Now()
		resp, err := client.Get(s.URL)
		if err != nil {
			return nil, fmt.Errorf("download request %d failed: %w", i+1, err)
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		speed := CalculateMbps(float64(resp.ContentLength), time.Since(start).Seconds())
		speeds = append(speeds, speed)
	}

	return speeds, nil
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
