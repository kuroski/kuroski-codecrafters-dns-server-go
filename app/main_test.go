package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseDnsHeader(t *testing.T) {
	testCases := []struct {
		name     string
		input    []byte
		expected *Header
		hasError bool
	}{
		{
			name:     "short_buffer",
			input:    []byte{0x01},
			expected: nil,
			hasError: true,
		},
		{
			name: "valid_case",
			input: append(
				make([]byte, 2),
				0x01,
				0x02,
				0x01, 0x02,
				0x03, 0x04,
				0x05, 0x06,
				0x07, 0x08,
			),
			expected: &Header{
				ID:      0,
				QR:      false,
				OPCODE:  0,
				AA:      false,
				TC:      false,
				RD:      true,
				RA:      false,
				Z:       0,
				RCODE:   2,
				QDCOUNT: 0x0102,
				ANCOUNT: 0x0304,
				NSCOUNT: 0x0506,
				ARCOUNT: 0x0708,
			},
			hasError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseDnsHeader(tc.input)

			assert.Equal(t, err != nil, tc.hasError)

			if !tc.hasError {
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestBoolToUint16(t *testing.T) {
	testCases := []struct {
		name     string
		input    bool
		expected uint16
	}{
		{
			name:     "true_input",
			input:    true,
			expected: 1,
		},
		{
			name:     "false_input",
			input:    false,
			expected: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := boolToUint16(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
