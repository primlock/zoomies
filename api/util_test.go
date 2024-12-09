package api

import (
	"strings"
	"testing"

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
