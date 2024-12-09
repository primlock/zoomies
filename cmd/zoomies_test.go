package cmd

import (
	"bytes"
	"fmt"
	"testing"

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
				fmt.Sprintf("--token=%s", tt.token),
			})

			got := cmd.Execute()

			assert.Error(t, got, tt.expected.Error())
		})
	}
}
