package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRIBEntry(t *testing.T) {
	t.Run("Test attribute ReadOrigin", func(t *testing.T) {
		// Example test case
		ribEntry := &RIBEntry{}
		err := ribEntry.ReadAttributeByType(1, []byte{0x00}) // IGP
		assert.NoError(t, err, "ReadOrigin should not return an error")
		assert.Equal(t, "IGP", ribEntry.Origin, "Origin should be IGP")

		err = ribEntry.ReadAttributeByType(1, []byte{0x01}) // EGP
		assert.NoError(t, err, "ReadOrigin should not return an error")
		assert.Equal(t, "EGP", ribEntry.Origin, "Origin should be EGP")

		err = ribEntry.ReadAttributeByType(1, []byte{0x02}) // INCOMPLETE
		assert.NoError(t, err, "ReadOrigin should not return an error")
		assert.Equal(t, "INCOMPLETE", ribEntry.Origin, "Origin should be INCOMPLETE")

		err = ribEntry.ReadAttributeByType(1, []byte{0x04}) // Invalid origin value
		assert.Error(t, err, "bad ReadOrigin should return an error")
	})

	t.Run("Test attribute Aggregator", func(t *testing.T) {
		ribEntry := &RIBEntry{}
		err := ribEntry.ReadAttributeByType(AGGREGATOR, []byte{0x00, 0x00, 0x00, 0x01, 0x7f, 0x00, 0x00, 0x01}) // AS number 1, IP 127.0.0.1
		assert.NoError(t, err, "ReadAggregator should not return an error")
		assert.Equal(t, "1 127.0.0.1", ribEntry.Aggregator, "Aggregator should be 1 127.0.0.1")

		ribEntry = &RIBEntry{}
		err = ribEntry.ReadAttributeByType(AGGREGATOR, []byte{0x00, 0x00, 0x00, 0x01}) // bad buffer
		assert.Error(t, err, "bad ReadAggregator should return an error")
	})

	t.Run("Test attribute MultiExitDisc", func(t *testing.T) {
		ribEntry := &RIBEntry{}
		err := ribEntry.ReadAttributeByType(MULTI_EXIT_DISC, []byte{0x00, 0x00, 0x00, 0x01})
		assert.NoError(t, err, "ReadMultiExitDisc should not return an error")
		assert.Equal(t, int32(1), ribEntry.MultiExitDisc, "MultiExitDisc should be 1")

		// Test with a bad buffer
		ribEntry = &RIBEntry{}
		err = ribEntry.ReadAttributeByType(MULTI_EXIT_DISC, []byte{0x00, 0x00, 0x00})
		assert.Error(t, err, "bad ReadMultiExitDisc should return an error")
	})

}
