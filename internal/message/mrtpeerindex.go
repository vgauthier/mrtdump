package message

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
)

type MRTPeerEntry struct {
	BGPId  net.IP // BGP ID of the peer
	PeerIP net.IP // IP address of the peer
	PeerAS uint32 // Autonomous System Number of the peer
}

func (e *MRTPeerEntry) String() string {
	return fmt.Sprintf("BGPID: %s, PeerIP: %s, PeerAS: %d",
		e.BGPId, e.PeerIP, e.PeerAS)
}

type MRTPeerIndex struct {
	CollectorBGPID uint32         // BGP ID of the collector
	ViewNameLen    uint16         // Length of the view name
	ViewName       string         // Name of the view
	Nentries       uint16         // Number of entries in the peer index
	Entries        []MRTPeerEntry // Entries for the peers in the index
}

func (h *MRTPeerIndex) Read(buf []byte) (Message, error) {
	if len(buf) < 4 {
		return nil, fmt.Errorf("buffer too short")
	}
	var o uint16 = 0
	var isIPv6 byte = 0x01
	h.CollectorBGPID = binary.BigEndian.Uint32(buf[o : o+4])
	o += 4
	h.ViewNameLen = binary.BigEndian.Uint16(buf[o : o+2])
	o += 2
	h.ViewName = string(buf[o : o+h.ViewNameLen])
	o += h.ViewNameLen
	h.Nentries = binary.BigEndian.Uint16(buf[o : o+2])
	o += 2
	h.Entries = make([]MRTPeerEntry, 0)
	for i := 0; i < int(h.Nentries); i++ {
		entry := MRTPeerEntry{}
		if o >= uint16(len(buf)) {
			return nil, fmt.Errorf("buffer too short for peer entry")
		}
		peerType := buf[o : o+1][0]
		o += 1
		entry.BGPId = net.IP(buf[o : o+4])
		o += 4
		if peerType&isIPv6 == isIPv6 {
			entry.PeerIP = net.IP(buf[o : o+16])
			o += 16
		} else {
			entry.PeerIP = net.IP(buf[o : o+4])
			o += 4
		}

		if peerType&0x2 == 0x2 {
			entry.PeerAS = binary.BigEndian.Uint32(buf[o : o+4])
			o += 4
		} else {
			entry.PeerAS = uint32(binary.BigEndian.Uint16(buf[o : o+2]))
			o += 2
		}
		h.Entries = append(h.Entries, entry)
	}
	return h, nil
}

func (h *MRTPeerIndex) String() string {
	var output strings.Builder
	output.WriteString("MRTPeerIndexHeader ")
	fmt.Fprintf(&output, "CollectorBGPId: %d, ViewNameLen: %d, ViewName: %s, Nentries: %d",
		h.CollectorBGPID, h.ViewNameLen, h.ViewName, h.Nentries)
	for i, entry := range h.Entries {
		fmt.Fprintf(&output, "\nPeerID: %d, %s", i, entry.String())
	}
	return output.String()
}

func (h *MRTPeerIndex) ToJSON() string {
	return "To JSON not implemented for MRTPeerIndex"
}

func NewMRTPeerIndex() *MRTPeerIndex {
	return &MRTPeerIndex{}
}
