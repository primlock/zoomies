package api

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"golang.org/x/net/html"
)

var (
	ErrScriptSrcAttrNotFound = errors.New("no src attribute found within the script tag")
	ErrTokenNotFound         = errors.New("token not found in provided string")
)

const (
	tokenRegex = `token:\s*"([^"]+)"`
)

func GeneratePayload(size int) ([]byte, error) {
	payload := make([]byte, size)
	_, err := rand.Read(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random payload: %w", err)
	}
	return payload, nil
}

func GetAPIEndpointToken() (string, error) {
	// Request for the HTML template where the .js script name lives.
	resp, err := http.Get(FastBaseURL)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	// Break down the request body into a tree of nodes which we can parse through
	script, err := getScriptName(resp.Body)
	if err != nil {
		return "", err
	}

	// Make a request to the server for the .js file.
	scriptURL := fmt.Sprintf("%s%s", FastBaseURL, script)
	resp, err = http.Get(scriptURL)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Pull the token out of the script
	token, err := extractToken(string(body))
	if err != nil {
		return "", err
	}

	return token, nil
}

func getScriptName(r io.Reader) (string, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return "", err
	}

	// Parse through the doc until you find the script tag containing the file name.
	for n := range doc.Descendants() {
		if n.Type == html.ElementNode && n.Data == "script" {
			for _, attr := range n.Attr {
				if attr.Key == "src" {
					return attr.Val, nil
				}
			}
		}
	}

	return "", ErrScriptSrcAttrNotFound
}

func extractToken(response string) (string, error) {
	re := regexp.MustCompile(tokenRegex)

	match := re.FindStringSubmatch(response)
	if len(match) < 2 {
		return "", ErrTokenNotFound
	}

	return match[1], nil
}

// BytesConsumed provides a human readable string that describes how much data was read or written.
func BytesConsumed(B uint64, binary bool) string {
	var val float64 = float64(B)
	var base float64 = 1000
	units := []string{"B", "KB", "MB", "GB"}

	if binary {
		base = 1024
		units = []string{"B", "KiB", "MiB", "GiB"}
	}

	var i int
	for val >= base && i < len(units)-1 {
		val /= base
		i++
	}

	return fmt.Sprintf("%.2f %s", val, units[i])
}

// CurrentBitRate provides a human readable string that describes the rate of the data transfer.
func CurrentBitRate(B uint64, start time.Time, binary bool) string {
	bps := float64(B*8) / time.Since(start).Seconds()
	var base float64 = 1000
	units := []string{"bps", "Kbps", "Mbps", "Gbps"}

	if binary {
		base = 1024
		units = []string{"bit/s", "Kibit/s", "Mibit/s", "Gibit/s"}
	}

	var i int
	for bps >= base && i < len(units)-1 {
		bps /= base
		i++
	}

	return fmt.Sprintf("%.2f %s", bps, units[i])
}
