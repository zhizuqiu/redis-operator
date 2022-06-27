package sm4

import (
	"testing"
)

func TestSM4Encrypt(t *testing.T) {
	var tests = []struct {
		in       string
		expected string
	}{
		{"pass", "fbd297723eb1d4a925b69d1437bb91ae"},
		{"", "830cb76d59c88d6c3a0e60d03faa5f34"},
	}

	for _, tt := range tests {
		actual, err := EncryptSm4([]byte(tt.in), Sm4Key)
		if err != nil {
			t.Fatalf("EncryptSm4 error: %s", err)
		}
		if actual != tt.expected {
			t.Fatalf("EncryptSm4(%s) = %s; expected %s", tt.in, actual, tt.expected)
		}
	}
}

func TestSM4Decrypt(t *testing.T) {
	var tests = []struct {
		in       string
		expected string
	}{
		{"fbd297723eb1d4a925b69d1437bb91ae", "pass"},
		{"830cb76d59c88d6c3a0e60d03faa5f34", ""},
	}

	for _, tt := range tests {
		actual, err := DecryptSm4([]byte(tt.in), Sm4Key)
		if err != nil {
			t.Fatalf("DecryptSm4 error: %s", err)
		}
		if actual != tt.expected {
			t.Fatalf("DecryptSm4(%s) = %s; expected %s", tt.in, actual, tt.expected)
		}
	}
}
