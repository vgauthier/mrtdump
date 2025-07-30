package message

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

var SimpleRIBEntry = []byte{
	0x00, 0x00, // PeerIndexId
	0x00, 0x00, 0x00, 0x01, // Originated Time
	0x00, 0x05, // Attributes Length
	0x10,       // attributesFlag BEGIN Attribute 1
	0x01,       // attributesType (ORIGIN)
	0x00, 0x01, // Attribute Length
	0x01, // BGP ORIGIN attribute (IGP)
}

func TestRIBEntry(t *testing.T) {
	t.Run("Test attribute ReadOrigin", func(t *testing.T) {
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

	t.Run("Test attribute NextHopV4", func(t *testing.T) {
		ribEntry := &RIBEntry{}
		err := ribEntry.ReadAttributeByType(NEXT_HOP, []byte{0x7f, 0x00, 0x00, 0x01}) // 127.0.0.1
		assert.NoError(t, err, "ReadNextHopV4 should not return an error")
		assert.Equal(t, net.IP{0x7f, 0x00, 0x00, 0x01}, ribEntry.NextHop, "NextHop should be 127.0.0.1")

		// Test with a bad buffer
		ribEntry = &RIBEntry{}
		err = ribEntry.ReadAttributeByType(NEXT_HOP, []byte{0x7f, 0x00, 0x00})
		assert.Error(t, err, "bad ReadNextHopV4 should return an error")
	})

	t.Run("Test attribute ASPath", func(t *testing.T) {
		ribEntry := &RIBEntry{}
		err := ribEntry.ReadAttributeByType(AS_PATH, []byte{0x2, 0x2, 0x0, 0x0, 0x8, 0x68, 0x0, 0x0, 0x34, 0x17})
		assert.NoError(t, err, "ReadASPath should not return an error")
		assert.Equal(t, []uint32{2152, 13335}, ribEntry.ASPath, "ASPath should contain AS number [2152, 13335]")

		// Test with a bad buffer
		ribEntry = &RIBEntry{}
		err = ribEntry.ReadAttributeByType(AS_PATH, []byte{0x02, 0x01})
		assert.Error(t, err, "bad ReadASPath should return an error")
	})

	t.Run("Test attribute Community", func(t *testing.T) {
		ribEntry := &RIBEntry{}
		err := ribEntry.ReadAttributeByType(COMMUNITY, []byte{0x00, 0x01, 0x00, 0x01})
		assert.NoError(t, err, "ReadCommunity should not return an error")
		assert.Equal(t, []string{"1:1"}, ribEntry.Communities, "Community should contain [1:1]")

		// Test with a bad buffer
		ribEntry = &RIBEntry{}
		err = ribEntry.ReadAttributeByType(COMMUNITY, []byte{0x00, 0x00})
		assert.Error(t, err, "bad ReadCommunity should return an error")

		// Test with a bad buffer
		ribEntry = &RIBEntry{}
		err = ribEntry.ReadAttributeByType(COMMUNITY, []byte{0x00, 0x01, 0x00, 0x01, 0x00, 0x01})
		assert.Error(t, err, "bad ReadCommunity should return an error")
	})

	t.Run("Test attribute LargeCommunities", func(t *testing.T) {
		ribEntry := &RIBEntry{}
		err := ribEntry.ReadAttributeByType(LARGE_COMMUNITIES, []byte{
			0x00, 0x00, 0x00, 0x01,
			0x00, 0x00, 0x00, 0x01,
			0x00, 0x00, 0x00, 0x01})
		assert.NoError(t, err, "ReadLargeCommunities should not return an error")
		assert.Equal(t, []string{"1:1:1"}, ribEntry.LargeCommunities, "LargeCommunities should contain [1:1:1]")

		// Test with a bad buffer
		ribEntry = &RIBEntry{}
		err = ribEntry.ReadAttributeByType(LARGE_COMMUNITIES, []byte{0x00, 0x01})
		assert.Error(t, err, "bad ReadLargeCommunities should return an error")
	})

	t.Run("Test attribute ReadAttributeByType with unknown type", func(t *testing.T) {
		ribEntry := &RIBEntry{}
		err := ribEntry.ReadAttributeByType(99, []byte{0x00, 0x01})
		assert.Error(t, err, "ReadAttributeByType should return an error for unknown type")
	})

	// Test reading a complete RIB entry
	t.Run("Test RIBEntry Read", func(t *testing.T) {
		ribEntry := &RIBEntry{}
		_, err := ribEntry.Read(SimpleRIBEntry)
		assert.NoError(t, err, "RIBEntry Read should not return an error")
		expected := "PeerIndexId: 0, OriginatedTime: 1970-01-01T01:00:01+01:00, NextHop: <nil>, Origin: EGP, ASPath: []"
		assert.Equal(t, expected, ribEntry.String(), "RIBEntry String() should return the expected value")
	})
}
