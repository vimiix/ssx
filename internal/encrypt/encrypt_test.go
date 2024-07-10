package encrypt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptDecrypt(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{"empty", ""},
		{"regular", "abc123"},
		{"symbol", "!*#$)@>?"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := Decrypt(Encrypt(tt.text))
			assert.Equal(t, tt.text, actual)
		})
	}
}

func TestDecrypt(t *testing.T) {
	tests := []struct {
		name   string
		cipher string
		expect string
	}{
		{"empty", "", ""},
		{"regular", "NmUxODZmYWM8PTxFPUQ9QENIQUc2eGl4T2pEWnQtQ0I2YkE0RkRxRUI0ei1fLUlNMmZKYi1lTFlnQk0=", "abc123"},
		{"plaintext", "abc123", "abc123"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := Decrypt(tt.cipher)
			assert.Equal(t, tt.expect, actual)
		})
	}
}
