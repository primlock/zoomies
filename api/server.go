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

	probing "github.com/prometheus-community/pro-bing"
	"github.com/pterm/pterm"
)

const (
	FastSpeedTestServerURL = "https://api.fast.com/netflix/speedtest/v2"
	FastBaseURL            = "https://fast.com"
)

type ProbeFunc func(server Server, count int) (time.Duration, error)

type Server struct {
	Name          string `json:"name"`
	URL           string `json:"url"`
	RangeBasedURL string `json:"rburl"`
	Location      struct {
		City    string `json:"city"`
		Country string `json:"country"`
	} `json:"location"`
}

var (
	DownloadPrinter = pterm.PrefixPrinter{
		MessageStyle: &pterm.Style{pterm.FgLightBlue},
		Prefix: pterm.Prefix{
			Style: &pterm.Style{pterm.FgBlack, pterm.BgLightBlue},
			Text:  "DOWNLOAD",
		},
	}

	UploadPrinter = pterm.PrefixPrinter{
		MessageStyle: &pterm.Style{pterm.FgLightMagenta},
		Prefix: pterm.Prefix{
			Style: &pterm.Style{pterm.FgBlack, pterm.BgLightMagenta},
			Text:  " UPLOAD ",
		},
	}

	LatencyPrinter = pterm.PrefixPrinter{
		MessageStyle: &pterm.Style{pterm.FgLightYellow},
		Prefix: pterm.Prefix{
			Style: &pterm.Style{pterm.FgBlack, pterm.BgLightYellow},
			Text:  "  PING  ",
		},
	}
)

func (s *Server) Download(requests int, chunk int64, duration time.Duration) (float64, error) {
	var total uint64
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

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

	// Create a channel to update the display
	displayChannel := make(chan bool)

	ticker := time.NewTicker(200 * time.Millisecond)

	spinner := &pterm.DefaultSpinner
	spinner = spinner.WithShowTimer(false)
	spinner.Style = &pterm.Style{pterm.FgLightBlue}
	spinner.InfoPrinter = &DownloadPrinter

	spinner, err = spinner.Start()
	if err != nil {
		return 0, err
	}

	updateDisplay := func(start time.Time) {
		for {
			select {
			case <-displayChannel:
				return
			case <-ticker.C:
				mbps := CalculateMbps(float64(total), time.Since(start).Seconds())
				spinner.UpdateText(pterm.Sprintf("Download: %.2f Mbps", mbps))
			}
		}
	}

	// Begin the concurrent downloads
	start := time.Now()
	for i := 0; i < requests; i++ {
		go downloadData()
	}

	go updateDisplay(start)

	// Main loop for orchastrating goroutines
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			displayChannel <- true

			avg := CalculateMbps(float64(total), time.Since(start).Seconds()) // TODO: calculate the average
			spinner.Info(pterm.Sprintf("%.2f Mbps", avg))

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

	// Create a channel to update the display
	displayChannel := make(chan bool)

	ticker := time.NewTicker(200 * time.Millisecond)

	spinner := &pterm.DefaultSpinner
	spinner = spinner.WithShowTimer(false)
	spinner.Style = &pterm.Style{pterm.FgLightMagenta}
	spinner.InfoPrinter = &UploadPrinter

	spinner, err := spinner.Start()
	if err != nil {
		return 0, err
	}

	updateDisplay := func(start time.Time) {
		for {
			select {
			case <-displayChannel:
				return
			case <-ticker.C:
				mbps := CalculateMbps(float64(total), time.Since(start).Seconds())
				spinner.UpdateText(pterm.Sprintf("Upload: %.2f Mbps", mbps))
			}
		}
	}

	// Begin the upload goroutines
	start := time.Now()
	for i := 0; i < requests; i++ {
		go uploadData()
	}

	go updateDisplay(start)

	// Main loop for orchastrating the downloads
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			displayChannel <- true

			avg := CalculateMbps(float64(total), time.Since(start).Seconds()) // TODO: calculate the average
			spinner.Info(pterm.Sprintf("%.2f Mbps", avg))

			return CalculateMbps(float64(total), time.Since(start).Seconds()), nil
		case <-uploadChannel:
			// Begin another upload while not timed out
			go uploadData()
		}
	}
}

func (s *Server) Latency(count int) error {
	// Create a channel for updating the display
	displayChannel := make(chan bool)

	ticker := time.NewTicker(400 * time.Millisecond)

	spinner := &pterm.DefaultSpinner
	spinner = spinner.WithShowTimer(false)
	spinner.Style = &pterm.Style{pterm.FgLightYellow}
	spinner.InfoPrinter = &LatencyPrinter

	spinner, err := spinner.Start()
	if err != nil {
		return err
	}

	dots := 0
	updateDisplay := func() {
		for {
			select {
			case <-displayChannel:
				return
			case <-ticker.C:
				switch dots {
				case 0:
					spinner.UpdateText("Ping:      ")
					dots++
				case 1:
					spinner.UpdateText("Ping: .    ")
					dots++
				case 2:
					spinner.UpdateText("Ping: . .  ")
					dots++
				case 3:
					spinner.UpdateText("Ping: . . .")
					dots = 0
				}
			}
		}
	}

	go updateDisplay()

	rtt, err := s.ICMPProbe(count)
	if err != nil {
		ticker.Stop()
		displayChannel <- true
		return err
	}

	ticker.Stop()
	displayChannel <- true

	t := rtt.Round(time.Millisecond)

	// Update the console with the results of the test
	spinner.Info(pterm.Sprintf("%s", t))

	spinner.Stop()

	return nil
}

func (s *Server) SetChunkSize(size int64) error {
	u, err := s.GetURL()
	if err != nil {
		return err
	}

	p := fmt.Sprintf("/range/0-%d", size)
	s.RangeBasedURL = u.JoinPath(p).String()

	return nil
}

// Get the IPv4 of the host URL.
func (s *Server) GetIPv4() (string, error) {
	u, err := s.GetURL()
	if err != nil {
		return "", err
	}

	ips, err := net.LookupIP(u.Host)
	if err != nil {
		return "", err
	}

	return ips[0].String(), nil
}

func (s *Server) GetURL() (*url.URL, error) {
	u, err := url.Parse(s.URL)
	if err != nil {
		return nil, fmt.Errorf("error parsing url for %s: %w", s.URL, err)
	}

	return u, nil
}

// Send a count number of ICMP pings to the server and return the average rtt.
func (s *Server) ICMPProbe(count int) (time.Duration, error) {
	u, err := s.GetURL()
	if err != nil {
		return 0, err
	}

	pinger, err := probing.NewPinger(u.Hostname())
	if err != nil {
		return 0, fmt.Errorf("error creating pinger for %s: %w", s.Name, err)
	}

	pinger.Count = count

	err = pinger.Run()
	if err != nil {
		return 0, fmt.Errorf("error probing server %s: %w", s.Name, err)
	}

	stats := pinger.Statistics()

	return stats.AvgRtt, nil
}

func (s *Server) HTTPProbe(count int) (time.Duration, error) {
	var rtt time.Duration
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest(http.MethodGet, s.URL, nil)
	if err != nil {
		return 0, fmt.Errorf("error creating request for %s: %w", s.URL, err)
	}

	for i := 0; i < count; i++ {
		start := time.Now()

		resp, err := client.Do(req)
		if err != nil {
			return 0, fmt.Errorf("error retrieving response for %s: %w", s.URL, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return 0, fmt.Errorf("unexpected status code %d for %s", resp.StatusCode, s.URL)
		}

		rtt += time.Since(start)
	}

	avg := rtt / time.Duration(count)
	return avg, nil
}

func ICMPProbe(server Server, count int) (time.Duration, error) {
	return server.ICMPProbe(count)
}

func HTTPProbe(server Server, count int) (time.Duration, error) {
	return server.HTTPProbe(count)
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
