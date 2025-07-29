package mrtheader

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMRTHeader(t *testing.T) {
	header := NewMRTHeader()
	// Example buffer with valid MRT header data
	// 68 63 25 00
	// 00 0d
	// 00 01
	// 00 00 01 7f
	testCase := []byte{
		0x68, 0x63, 0x25, 0x00, // Timestamp
		0x00, 0x0d, // Type
		0x00, 0x01, // Subtype
		0x00, 0x00, 0x01, 0x7f, // Length
	}
	// Expected values after parsing
	expected := &MRTHeader{
		Ts:      1751328000,
		Type:    13,
		Subtype: 1,
		Length:  383,
	}

	t.Run("Test MRT Header", func(t *testing.T) {
		// Parse the buffer
		got, err := header.Parse(testCase)
		assert.NoError(t, err, "Parse should not return an error")
		// Check if the parsed values match the expected values
		assert.Equal(t, expected, got)
	})
}
