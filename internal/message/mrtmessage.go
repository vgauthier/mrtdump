package message

import (
	"fmt"

	"github.com/vgauthier/mrtdump/internal/mrtheader"
)

const (
	PEER_INDEX_TABLE = iota + 1 // Peer index type
	RIB_IPV4_UNICAST            // RIB IPv4 Unicast subtype
)

const (
	TABLE_DUMP    = 12 // Type for Table Dump messages
	TABLE_DUMP_V2 = 13 // Type for Table Dump V2 messages
)

type typeTuple struct {
	Type    uint16
	Subtype uint16
}

func NewMRTMessage(header *mrtheader.MRTHeader) *MRTMessage {
	return &MRTMessage{Type: header.Type, SubType: header.Subtype}
}

type MRTMessage struct {
	Type      uint16        // Type of the MRT message
	SubType   uint16        // Subtype of the MRT message
	Message   Message       // Generic message interface
	Err       error         // Error if any occurred during parsing
	PeerIndex *MRTPeerIndex // Peer index if applicable
}

func (m *MRTMessage) Parse(buf []byte) (*MRTMessage, error) {
	switch (typeTuple{Type: m.Type, Subtype: m.SubType}) {
	case typeTuple{TABLE_DUMP_V2, PEER_INDEX_TABLE}:
		m.Message, m.Err = NewMRTPeerIndex().Read(buf)
		return m, nil
	case typeTuple{TABLE_DUMP_V2, RIB_IPV4_UNICAST}:
		// Handle subtypes 2 RIB_IPV4_UNICAST
		if m.PeerIndex == nil {
			m.Message, m.Err = NewTableDumpV2(RIB_IPV4_UNICAST).Read(buf)
		} else {
			m.Message, m.Err = NewTableDumpV2(RIB_IPV4_UNICAST).WithPeerIndex(m.PeerIndex).Read(buf)
		}
		return m, nil
	}
	return nil, fmt.Errorf("unsupported MRT type %d subtype %d", m.Type, m.SubType)
}

func (m *MRTMessage) GetMessage() Message {
	if m.Message == nil {
		return nil
	}
	return m.Message
}

func (m *MRTMessage) WithPeerIndex(peerIndex *MRTPeerIndex) *MRTMessage {
	if peerIndex != nil {
		m.PeerIndex = peerIndex
	}
	return m
}

func (m *MRTMessage) String() string {
	return "MRTMessage"
}
