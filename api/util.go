package api

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"regexp"

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

func CalculateMbps(contentLength, duration float64) float64 {
	return (contentLength * 8) / (duration * 1_000_000)
}

func CalculateAverage(speeds []float64) float64 {
	sum := 0.0
	for _, speed := range speeds {
		sum += speed
	}
	return sum / float64(len(speeds))
}

func CalculateStdDev(speeds []float64, mean float64) float64 {
	varianceSum := 0.0
	for _, speed := range speeds {
		varianceSum += math.Pow(speed-mean, 2)
	}
	variance := varianceSum / float64(len(speeds))
	return math.Sqrt(variance)
}

func DisplayTestResults(speeds []float64) {
	avg := CalculateAverage(speeds)
	stdDev := CalculateStdDev(speeds, avg)

	fmt.Printf("\nAverage Speed: %.2f Mbps - %v\n", avg, speeds)
	fmt.Printf("Standard Deviation: %.2f Mbps\n\n", stdDev)
}
