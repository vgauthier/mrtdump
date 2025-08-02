package message

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"
)

type From struct {
	PeerIP string // IP address of the peer
	PeerAS uint32 // AS number of the peer
}

type TableDumpV2Json struct {
	Datetime         string   `json:"datetime"`                   // Timestamp in RFC3339 format
	Type             uint16   `json:"type"`                       // Type of the message
	Subtype          uint16   `json:"subtype"`                    // Subtype of the message
	Prefix           string   `json:"prefix"`                     // network prefix
	PrefixLen        uint8    `json:"prefixLen"`                  // Length of the prefix
	Sequence         uint32   `json:"sequence"`                   // Sequence number of the RIB entry
	Originated       string   `json:"originated"`                 // Time when the entry was originated in RFC3339 format
	Origin           string   `json:"origin"`                     // Origin of the entry
	ASPath           []uint32 `json:"asPath"`                     // AS Path
	From             From     `json:"from"`                       // Information about the peer
	NextHop          string   `json:"nextHop"`                    // Next hop IP address
	Communities      []string `json:"communities,omitempty"`      // Communities associated with the entry
	LargeCommunities []string `json:"largeCommunities,omitempty"` // Large communities associated with the entry
	Aggregator       string   `json:"aggregator,omitempty"`       // Aggregator information
	MultiExitDisc    *int32   `json:"multiExitDisc,omitempty"`    // Multi Exit Discriminator
}

type TableDumpV2 struct {
	PeerIndex      *MRTPeerIndex // Peer index from the MRT header
	Timestamp      int32         // Unix timestamp in seconds
	Type           uint16
	Subtype        uint16
	TypeName       string // Name of the type
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
	for _, entry := range t.Entries {
		sb.WriteString(fmt.Sprintf("TIME: %s\n", time.Unix(int64(t.Timestamp), 0).Format(time.RFC3339)))
		sb.WriteString(fmt.Sprintf("TYPE: %s\n", t.TypeName))
		sb.WriteString(fmt.Sprintf("PREFIX: %s/%d\n", t.Prefix, t.PrefixLen))
		sb.WriteString(fmt.Sprintf("SEQUENCE: %d\n", t.SequenceNumber))

		if t.PeerIndex != nil {
			sb.WriteString(fmt.Sprintf("FROM: %s AS%d\n",
				t.PeerIndex.Entries[entry.PeerIndexId].PeerIP,
				t.PeerIndex.Entries[entry.PeerIndexId].PeerAS))
		} else {
			sb.WriteString(fmt.Sprintf("PeerIndexId: %d\n", entry.PeerIndexId))
		}
		sb.WriteString(fmt.Sprintf("ORIGINATED: %s\n", time.Unix(int64(entry.OriginatedTime), 0).Format(time.RFC3339)))
		sb.WriteString(fmt.Sprintf("ORIGIN: %s\n", entry.Origin))

		// AS Path
		sb.WriteString("ASPATH: ")
		for _, asn := range entry.ASPath {
			sb.WriteString(fmt.Sprintf("AS%d ", asn))
		}
		sb.WriteString("\n")

		// Next Hop
		sb.WriteString(fmt.Sprintf("NEXT_HOP: %s\n", entry.NextHop.String()))

		// Multi Exit Disc
		if entry.MultiExitDisc != nil {
			sb.WriteString(fmt.Sprintf("MULTI_EXIT_DISC: %d\n", *entry.MultiExitDisc))
		}

		// Communities
		if len(entry.Communities) > 0 {
			sb.WriteString(fmt.Sprintf("COMMUNITIES: %v\n", strings.Trim(fmt.Sprintf("%v", entry.Communities), "[]")))
		}

		// Large Communities
		if len(entry.LargeCommunities) > 0 {
			sb.WriteString(fmt.Sprintf("LARGE_COMMUNITIES: %v\n", strings.Trim(fmt.Sprintf("%v", entry.LargeCommunities), "[]")))
		}
		// Aggregator
		if len(entry.Aggregator) > 0 {
			sb.WriteString(fmt.Sprintf("AGGREGATOR: %s\n", entry.Aggregator))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func (t *TableDumpV2) ToJSON() string {
	var jsonEntry bytes.Buffer
	for _, entry := range t.Entries {
		b, err := json.MarshalIndent(TableDumpV2Json{
			Datetime:   time.Unix(int64(t.Timestamp), 0).Format(time.RFC3339),
			Type:       t.Type,
			Subtype:    t.Subtype,
			Prefix:     t.Prefix.String(),
			PrefixLen:  t.PrefixLen,
			Sequence:   t.SequenceNumber,
			Originated: time.Unix(int64(entry.OriginatedTime), 0).Format(time.RFC3339),
			ASPath:     entry.ASPath,
			From: From{
				PeerIP: t.PeerIndex.Entries[entry.PeerIndexId].PeerIP.String(),
				PeerAS: t.PeerIndex.Entries[entry.PeerIndexId].PeerAS,
			},
			NextHop:          entry.NextHop.String(),
			Communities:      entry.Communities,
			Origin:           entry.Origin,
			LargeCommunities: entry.LargeCommunities,
			Aggregator:       entry.Aggregator,
			MultiExitDisc:    entry.MultiExitDisc,
		}, "    ", "    ")
		if err == nil {
			jsonEntry.Write(b)
			jsonEntry.WriteString(",\n    ")
		}
	}
	return jsonEntry.String()
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
	if subType != RIB_IPV4_UNICAST {
		return nil
	}
	return &TableDumpV2{
		Subtype:  subType,
		Type:     TABLE_DUMP_V2,
		TypeName: "TABLE_DUMP_V2/IPV4_UNICAST",
	}
}
