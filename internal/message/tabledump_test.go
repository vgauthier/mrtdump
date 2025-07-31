package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var SimpleTableDumpHeader = []byte{
	0x00, 0x00, 0x00, 0x01, // SequenceNumber
	0x18,             // PrefixLen
	0xc0, 0xa8, 0x01, // Prefix (192.168.1.1)
	0x00, 0x01, // EntryCount
}

func TestTableDumpV2(t *testing.T) {
	t.Run("Test TableDump Read", func(t *testing.T) {
		tableDump := NewTableDumpV2(RIB_IPV4_UNICAST)
		buf := append(SimpleTableDumpHeader, SimpleRIBEntry...) // Append a simple RIB entry
		m, err := tableDump.Read(buf)
		assert.NoError(t, err, "TableDump Read should not return an error")
		assert.EqualValues(t, 1, m.(*TableDumpV2).SequenceNumber)
	})

	t.Run("Test TableDump Read with overflow", func(t *testing.T) {
		// Add your test cases here
		tableDump := NewTableDumpV2(RIB_IPV4_UNICAST)
		buf := []byte{
			0x00, 0x00, 0x00, 0x01, // SequenceNumber
			0x18,             // PrefixLen
			0xc0, 0xa8, 0x01, // Prefix (192.168.1.1)
			0x00, 0x06, // EntryCount
		}
		buf = append(buf, SimpleRIBEntry...) // Append a simple RIB entry
		_, err := tableDump.Read(buf)
		assert.Error(t, err, "TableDump Read should return an error due to buffer overflow")
	})

	t.Run("Test TableDump String", func(t *testing.T) {
		tableDump := NewTableDumpV2(RIB_IPV4_UNICAST)
		buf := append(SimpleTableDumpHeader, SimpleRIBEntry...) // Append a simple RIB entry
		_, err := tableDump.Read(buf)
		assert.NoError(t, err, "TableDump Read should not return an error")
		assert.Equal(t, 179, len(tableDump.String())) // Call String method to ensure it works without error
	})
}
