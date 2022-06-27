package util

import (
	"encoding/base64"
	"testing"
)

func TestMD5(t *testing.T) {
	var decryptTests = []struct {
		in       string
		expected string
	}{
		{"pass", "1a1dc91c907325c69271ddf0c944bc72"},
		{"", "d41d8cd98f00b204e9800998ecf8427e"},
	}

	for _, tt := range decryptTests {
		actual := MD5(tt.in)
		if actual != tt.expected {
			t.Errorf("MD5(%s) = %s; expected %s", tt.in, actual, tt.expected)
		}
	}
}

func TestBase64Encrypt(t *testing.T) {
	var tests = []struct {
		in       string
		expected string
	}{
		{"pass", "cGFzcw=="},
		{"", ""},
	}

	for _, tt := range tests {
		actual := base64.StdEncoding.EncodeToString([]byte(tt.in))
		if actual != tt.expected {
			t.Fatalf("Base64Encrypt(%s) = %s; expected %s", tt.in, actual, tt.expected)
		}
	}
}

func TestBase64Decrypt(t *testing.T) {
	var tests = []struct {
		in       string
		expected string
	}{
		{"cGFzcw==", "pass"},
		{"", ""},
	}

	for _, tt := range tests {
		actual, err := base64.StdEncoding.DecodeString(tt.in)
		if err != nil {
			t.Fatalf("Base64Decrypt error: %s", err)
		}
		actualStr := string(actual)
		if string(actual) != tt.expected {
			t.Fatalf("Base64Decrypt(%s) = %s; expected %s", tt.in, actualStr, tt.expected)
		}
	}
}
