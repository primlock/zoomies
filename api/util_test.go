package api

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestGetScriptName(t *testing.T) {
	testCases := []struct {
		name         string
		html         string
		expected_val string
		expected_err error
	}{
		{
			name: "1",
			html: `
					<!DOCTYPE html>
					<html>
					<head></head>
					<body>
						<div class="container">
							<p>Hello, World!<p>
						</div>
						<script src="target.js"></script>
					</body>
					</html>`,
			expected_val: "target.js",
			expected_err: nil,
		},
		{
			name: "2",
			html: `
					<!DOCTYPE html>
					<html>
					<head></head>
					<body>
						<div class="container">
							<p>Hello, World!<p>
						</div>
					</body>
					</html>`,
			expected_val: "",
			expected_err: ErrScriptSrcAttrNotFound,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			got_val, got_err := getScriptName(strings.NewReader(tt.html))

			assert.Equal(t, got_val, tt.expected_val)
			assert.Equal(t, got_err, tt.expected_err)
		})
	}
}

func TestExtractToken(t *testing.T) {
	testCases := []struct {
		name         string
		s            string
		expected_val string
		expected_err error
	}{
		{
			name:         "1",
			s:            `object:{isEnabled:false,endpoint:auth,token:"FnmAejbbyAYbmMUpMj",n:5}`,
			expected_val: "FnmAejbbyAYbmMUpMj",
			expected_err: nil,
		},
		{
			name:         "2",
			s:            `object:{isEnabled:false,n:5}`,
			expected_val: "",
			expected_err: ErrTokenNotFound,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			got_val, got_err := extractToken(tt.s)

			assert.Equal(t, got_val, tt.expected_val)
			assert.Equal(t, got_err, tt.expected_err)
		})
	}
}

func TestBytesConsumed(t *testing.T) {
	testCases := []struct {
		name     string
		bytes    uint64
		binary   bool
		expected string
	}{
		{name: "Bytes conversion", bytes: 1, binary: true, expected: "1.00 B"},
		{name: "Kibibytes conversion", bytes: 1024, binary: true, expected: "1.00 KiB"},
		{name: "Kilobytes conversion", bytes: 1000, binary: false, expected: "1.00 KB"},
		{name: "Mebibytes conversion", bytes: 1048576, binary: true, expected: "1.00 MiB"},
		{name: "Megabytes conversion", bytes: 1000000, binary: false, expected: "1.00 MB"},
		{name: "Gibibytes conversion", bytes: 1073741824, binary: true, expected: "1.00 GiB"},
		{name: "Gigabytes conversion", bytes: 1000000000, binary: false, expected: "1.00 GB"},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			got := BytesConsumed(tt.bytes, tt.binary)

			if got != tt.expected {
				t.Errorf("got %v, want %v", got, tt.expected)
			}
		})
	}
}

func parseBitRate(bitRate string) (float64, string, error) {
	parts := strings.Fields(bitRate)
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("invalid format: %v", bitRate)
	}
	value, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, "", err
	}
	return value, parts[1], nil
}

func convertToBaseUnit(value float64, unit string, binary bool) float64 {
	base := 1000.0
	if binary {
		base = 1024.0
	}
	unitMap := map[string]int{
		"bps":     0,
		"Kbps":    1,
		"Mbps":    2,
		"Gbps":    3,
		"bit/s":   0,
		"Kibit/s": 1,
		"Mibit/s": 2,
		"Gibit/s": 3,
	}
	power := unitMap[unit]
	return value * math.Pow(base, float64(power))
}

func TestCurrentBitRate(t *testing.T) {
	tolerance := 0.05 // 5%
	testCases := []struct {
		name     string
		bytes    uint64
		duration time.Duration
		binary   bool
		expected string
	}{
		{name: "500 Kbps rate (decimal)", bytes: 62500, duration: 1 * time.Second, binary: false, expected: "500.00 Kbps"},
		{name: "512 Kibps rate (binary)", bytes: 65536, duration: 1 * time.Second, binary: true, expected: "512.00 Kibit/s"},
		{name: "1 Mbps rate (decimal)", bytes: 125000, duration: 1 * time.Second, binary: false, expected: "1.00 Mbps"},
		{name: "1 Mibps rate (binary)", bytes: 131072, duration: 1 * time.Second, binary: true, expected: "1.00 Mibit/s"},
		{name: "8 Gbps rate (decimal)", bytes: 1000000000, duration: 1 * time.Second, binary: false, expected: "8.00 Gbps"},
		{name: "2 Gibit/s rate (binary)", bytes: 268435456, duration: 1 * time.Second, binary: true, expected: "2.00 Gibit/s"},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			time.Sleep(tt.duration)

			got := CurrentBitRate(tt.bytes, start, tt.binary)

			// Parse the actual and expected results
			gotValue, gotUnit, err := parseBitRate(got)
			if err != nil {
				t.Fatalf("failed to parse output: %v", err)
			}

			expectedValue, expectedUnit, err := parseBitRate(tt.expected)
			if err != nil {
				t.Fatalf("failed to parse expected value: %v", err)
			}

			// Convert both values to base unit (bps)
			gotBaseValue := convertToBaseUnit(gotValue, gotUnit, tt.binary)
			expectedBaseValue := convertToBaseUnit(expectedValue, expectedUnit, tt.binary)

			// Assert the values are within tolerance
			diff := math.Abs(gotBaseValue - expectedBaseValue)
			if diff > expectedBaseValue*tolerance {
				t.Errorf("value out of tolerance: got %.2f %s, expected %.2f %s (difference: %.2f bps)", gotValue, gotUnit, expectedValue, expectedUnit, diff)
			}
		})
	}
}
