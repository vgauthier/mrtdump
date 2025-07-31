package message

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
)

type TableDumpV2 struct {
	PeerIndex      *MRTPeerIndex // Peer index from the MRT header
	Timestamp      int32         // Unix timestamp in seconds
	Type           uint16
	Subtype        uint16
	Length         uint32
	SequenceNumber uint32 // Sequence number of the RIB entry
	PrefixLen      byte   // Length of the prefix
	Prefix         net.IP // network prefix
	EntryCount     uint16 // Number of entries in the RIB
	Entries        []RIBEntry
}

func (t *TableDumpV2) Read(buf []byte) (Message, error) {
	// Handle subtypes 2 RIB_IPV4_UNICAST
	if t.Subtype != RIB_IPV4_UNICAST {
		return nil, fmt.Errorf("unsupported subtype %d for TableDumpV2", t.Subtype)
	}
	o := 0
	t.SequenceNumber = binary.BigEndian.Uint32(buf[o : o+4])
	o += 4
	// extract Prefix length
	t.PrefixLen = buf[o : o+1][0]
	o += 1
	// Compute Prefix length in bytes
	prefixLenBytes := int((t.PrefixLen + 7) / 8)
	// extract the prefix
	PrefixBytes := make([]byte, prefixLenBytes)
	copy(PrefixBytes, buf[o:o+prefixLenBytes])
	// add 0 padding if necessary
	for i := 0; i < 4-prefixLenBytes; i++ {
		PrefixBytes = append(PrefixBytes, 0)
	}
	// format the prefix
	t.Prefix = net.IP(PrefixBytes)
	o += prefixLenBytes
	// Entry count
	t.EntryCount = binary.BigEndian.Uint16(buf[o : o+2])
	o += 2
	// Extract the entries
	_, err := t.ReadEntries(buf[o:])
	if err != nil {
		return nil, fmt.Errorf("error reading entries: %v", err)
	}
	return t, nil
}

func (t *TableDumpV2) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("TableDumpV2 SeqNumber: %d, PrefixLen: %d, Prefix: %s/%d, EntryCount: %d \n",
		t.SequenceNumber, t.PrefixLen, t.Prefix, t.PrefixLen, t.EntryCount))
	for _, entry := range t.Entries {
		sb.WriteString(fmt.Sprintf("%s\n", entry.String()))
	}
	return sb.String()
}

func (t *TableDumpV2) ReadEntries(buf []byte) (int, error) {
	o := 0
	t.Entries = make([]RIBEntry, t.EntryCount)
	for i := 0; i < int(t.EntryCount); i++ {
		if o >= len(buf) {
			return o, fmt.Errorf("buffer overflow while reading RIB entries")
		}
		if t.PeerIndex != nil {
			t.Entries[i].PeerIndex = t.PeerIndex
		}
		length, err := t.Entries[i].Read(buf[o:])
		if err != nil {
			fmt.Printf("Error reading RIB entry: %v\n", err)
		}
		o += length
	}
	return o, nil
}

func (t *TableDumpV2) WithPeerIndex(peerIndex *MRTPeerIndex) *TableDumpV2 {
	t.PeerIndex = peerIndex
	return t
}

func NewTableDumpV2(subType uint16) *TableDumpV2 {
	return &TableDumpV2{
		Subtype: subType,
		Type:    TABLE_DUMP_V2,
	}
}
