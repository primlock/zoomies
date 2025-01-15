package cmd

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/primlock/zoomies/api"
	"gotest.tools/v3/assert"
)

func TestURLCountOutOfBounds(t *testing.T) {
	testCases := []struct {
		name     string
		count    int
		expected error
	}{
		{name: "1", count: 0, expected: ErrURLCountOutOfBounds},
		{name: "2", count: 6, expected: ErrURLCountOutOfBounds},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCmd()

			c.SetOutput(&bytes.Buffer{})
			c.SetArgs([]string{
				fmt.Sprintf("--count=%d", tt.count),
			})

			got := c.Execute()

			assert.Error(t, got, tt.expected.Error())
		})
	}
}

func TestBadToken(t *testing.T) {
	testCases := []struct {
		name     string
		token    string
		expected error
	}{
		{name: "1", token: "invalid", expected: ErrUnknownAppToken},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCmd()

			c.SetOutput(&bytes.Buffer{})
			c.SetArgs([]string{
				fmt.Sprintf("--token=%s", tt.token),
			})

			got := c.Execute()

			assert.Error(t, got, tt.expected.Error())
		})
	}
}

func TestChunkSizeOutOfBounds(t *testing.T) {
	testCases := []struct {
		name     string
		n        int64
		expected error
	}{
		{name: "1", n: 0, expected: ErrChunkSizeOutOfBounds},
		{name: "2", n: 52428800, expected: ErrChunkSizeOutOfBounds},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCmd()

			c.SetOutput(&bytes.Buffer{})
			c.SetArgs([]string{
				fmt.Sprintf("--chunk=%d", tt.n),
			})

			got := c.Execute()

			assert.Error(t, got, tt.expected.Error())
		})
	}
}

var mockRTT time.Duration = 20

func mockProbeFunc(server api.Server, count int) (time.Duration, error) {
	mockRTT += 20
	return mockRTT, nil
}

func TestGetLowestRTTServers(t *testing.T) {
	testCases := []struct {
		name       string
		candidates []api.Server
		count      int
		probeFunc  api.ProbeFunc
		expected   []string
		err        error
	}{
		{
			name: "Top 2 servers with the lowest rtt",
			candidates: []api.Server{
				{Name: "server1"},
				{Name: "server2"},
				{Name: "server3"},
			},
			count:     2,
			probeFunc: mockProbeFunc,
			expected:  []string{"server1", "server2"},
			err:       nil,
		},
		{
			name: "Request more servers than available",
			candidates: []api.Server{
				{Name: "server1"},
				{Name: "server2"},
			},
			count:     4,
			probeFunc: mockProbeFunc,
			expected:  []string{"server1", "server2"},
			err:       nil,
		},
		{
			name:       "Empty list of servers passed",
			candidates: []api.Server{},
			count:      4,
			probeFunc:  mockProbeFunc,
			expected:   []string{},
			err:        ErrNoCandidatesToRank,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getLowestRTTServers(tt.candidates, tt.count, tt.probeFunc)
			if err != nil {
				assert.Error(t, err, tt.err.Error())
			}

			names := make([]string, len(got))
			for i, server := range got {
				names[i] = server.Name
			}

			if !reflect.DeepEqual(names, tt.expected) {
				t.Errorf("got %v, want %v", names, tt.expected)
			}
		})
	}
}

func TestDurationOutOfBounds(t *testing.T) {
	testCases := []struct {
		name     string
		seconds  int
		expected error
	}{
		{name: "Below the lower boundary", seconds: 2, expected: ErrDurationOutOfBounds},
		{name: "Above the upper boundary", seconds: 35, expected: ErrDurationOutOfBounds},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCmd()

			c.SetOutput(&bytes.Buffer{})
			c.SetArgs([]string{
				fmt.Sprintf("--duration=%d", tt.seconds),
			})

			got := c.Execute()

			assert.Error(t, got, tt.expected.Error())
		})
	}
}

func TestPingCountOutOfBounds(t *testing.T) {
	testCases := []struct {
		name     string
		count    int
		expected error
	}{
		{name: "Below the lower boundary", count: -1, expected: ErrPingCountOutOfBounds},
		{name: "Above the upper boundary", count: 7, expected: ErrPingCountOutOfBounds},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCmd()

			c.SetOutput(&bytes.Buffer{})
			c.SetArgs([]string{
				fmt.Sprintf("--pcount=%d", tt.count),
			})

			got := c.Execute()

			assert.Error(t, got, tt.expected.Error())
		})
	}
}
