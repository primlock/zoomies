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
	Name          string `json:"name"`
	URL           string `json:"url"`
	RangeBasedURL string `json:"rburl"`
	Location      struct {
		City    string `json:"city"`
		Country string `json:"country"`
	} `json:"location"`
}

func (s *Server) Download(requests int, chunk int64, duration time.Duration) (float64, error) {
	var total uint64
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	fmt.Printf("endpoint: %s\n", s.RangeBasedURL)

	// Create a default request for downloading the data
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.RangeBasedURL, nil)
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
			if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
				fmt.Printf("failed when making http request: %s", err)
			}
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
			return CalculateMbps(float64(total), time.Since(start).Seconds()), nil
		case <-downloadChannel:
			// Begin another download while not timed out
			go downloadData()
		}
	}
}

func (s *Server) Upload(requests int, duration time.Duration, payload []byte) (float64, error) {
	var total uint64
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	// Create a channel for tracking uploads
	uploadChannel := make(chan struct{}, requests)

	uploadData := func() {
		// Generate a request for the URL
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.URL, bytes.NewReader(payload))
		if err != nil {
			fmt.Printf("failed to generate http request: %s", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
				fmt.Printf("failed when making http request: %s", err)
			}
		} else {
			defer resp.Body.Close()

			atomic.AddUint64(&total, uint64(len(payload)))

			// Signal the channel that the upload finished
			uploadChannel <- struct{}{}
		}
	}

	// Begin the upload goroutines
	start := time.Now()
	for i := 0; i < requests; i++ {
		go uploadData()
	}

	// Main loop for orchastrating the downloads
	for {
		select {
		case <-ctx.Done():
			return CalculateMbps(float64(total), time.Since(start).Seconds()), nil
		case <-uploadChannel:
			// Begin another upload while not timed out
			go uploadData()
		}
	}
}

func (s *Server) SetChunkSize(size int64) error {
	u, err := url.Parse(s.URL)
	if err != nil {
		return err
	}

	p := fmt.Sprintf("/range/0-%d", size)
	s.RangeBasedURL = u.JoinPath(p).String()

	return nil
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
