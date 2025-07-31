package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vgauthier/mrtdump/internal/mrtheader"
)

func TestMRTMessage(t *testing.T) {
	t.Run("Test MRTMessage Read", func(t *testing.T) {
		buf := append(SimpleTableDumpHeader, SimpleRIBEntry...) // Append a simple RIB entry
		header := &mrtheader.MRTHeader{
			Ts:      1622547800, // Example timestamp
			Type:    TABLE_DUMP_V2,
			Subtype: RIB_IPV4_UNICAST,
			Length:  uint32(len(buf)), // Example length
		}
		message, err := NewMRTMessage(header).Parse(buf)
		assert.NotNil(t, message)
		assert.NoError(t, err)

		m, err := message.GetMessage()
		assert.NotNil(t, m)
		assert.NoError(t, err)
	})

	t.Run("Test MRTMessage String", func(t *testing.T) {
		// Add your test cases here
	})
}
