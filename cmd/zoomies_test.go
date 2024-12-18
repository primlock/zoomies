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
			cmd.SetOutput(&bytes.Buffer{})
			cmd.SetArgs([]string{
				fmt.Sprintf("--count=%d", tt.count),
			})

			got := cmd.Execute()

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
			cmd.SetOutput(&bytes.Buffer{})
			cmd.SetArgs([]string{
				fmt.Sprintf("--count=%d", 1),
				fmt.Sprintf("--token=%s", tt.token),
			})

			got := cmd.Execute()

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
			cmd.SetOutput(&bytes.Buffer{})
			cmd.SetArgs([]string{
				fmt.Sprintf("--chunk=%d", tt.n),
			})

			got := cmd.Execute()

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
	}{
		{
			name: "Top 2 Servers with the Lowest RTT",
			candidates: []api.Server{
				{Name: "server1"},
				{Name: "server2"},
				{Name: "server3"},
			},
			count:     2,
			probeFunc: mockProbeFunc,
			expected:  []string{"server1", "server2"},
		},
		{
			name: "Request More Servers than Available",
			candidates: []api.Server{
				{Name: "server1"},
				{Name: "server2"},
			},
			count:     4,
			probeFunc: mockProbeFunc,
			expected:  []string{"server1", "server2"},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getLowestRTTServers(tt.candidates, tt.count, tt.probeFunc)
			if err != nil {
				t.Errorf("got unexpected error: %v", err)
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
