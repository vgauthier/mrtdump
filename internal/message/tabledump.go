package message

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"time"
)

type RIBEntry struct {
	PeerIndex        *MRTPeerIndex
	PeerIndexId      uint16 // Peer index from the MRT header
	OriginatedTime   uint32 // Unix timestamp in seconds
	AttributeLength  uint16
	LargeCommunities []string
	Communities      []string
	NextHop          net.IP
	Origin           string
	ASPath           []uint32
	MultiExitDisc    int32
	Aggregator       string // Aggregator information
}

func (r *RIBEntry) String() string {
	var sb strings.Builder

	if r.PeerIndex != nil {
		sb.WriteString(fmt.Sprintf("Peer AS: %v, ", r.PeerIndex.Entries[r.PeerIndexId].PeerAS))
		sb.WriteString(fmt.Sprintf("Peer IP: %s, ", r.PeerIndex.Entries[r.PeerIndexId].PeerIP.String()))
	} else {
		sb.WriteString(fmt.Sprintf("PeerIndexId: %d, ", r.PeerIndexId))
	}
	sb.WriteString(fmt.Sprintf("OriginatedTime: %s, ", time.Unix(int64(r.OriginatedTime), 0).Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("NextHop: %s, ", r.NextHop))
	sb.WriteString(fmt.Sprintf("Origin: %s, ", r.Origin))
	sb.WriteString(fmt.Sprintf("ASPath: %v", r.ASPath))
	if r.MultiExitDisc != -1 {
		sb.WriteString(fmt.Sprintf(", MultiExitDisc: %d", r.MultiExitDisc))
	}
	if len(r.Communities) > 0 {
		sb.WriteString(fmt.Sprintf(", Communities: %v", r.Communities))
	}
	if len(r.LargeCommunities) > 0 {
		sb.WriteString(fmt.Sprintf(", LargeCommunities: %v", r.LargeCommunities))
	}
	if len(r.Aggregator) > 0 {
		sb.WriteString(fmt.Sprintf(", Aggregator: %s", r.Aggregator))
	}
	return sb.String()
}

func (r *RIBEntry) ReadMultiExitDisc(bgpPathAttribute []byte) error {
	// MULTI_EXIT_DISC is a 4-byte field
	// It contains the Multi-Exit Discriminator value
	if len(bgpPathAttribute) < 4 {
		return fmt.Errorf("MULTI_EXIT_DISC: Invalid length")
	}
	r.MultiExitDisc = int32(binary.BigEndian.Uint32(bgpPathAttribute[:4]))
	return nil
}

func (r *RIBEntry) ReadLargeCommunities(bgpPathAttribute []byte) error {
	// LARGE_COMMUNITIES is a variable-length field
	// The first four bytes are the AS number
	// The next four bytes are the community value
	if len(bgpPathAttribute) < 12 {
		return fmt.Errorf("LARGE_COMMUNITIES: Invalid length")
	}
	r.LargeCommunities = []string{}
	communityCount := len(bgpPathAttribute) / 12 // Each community is 12 bytes
	o := 0
	for i := 0; i < int(communityCount); i++ {
		asNumber := binary.BigEndian.Uint32(bgpPathAttribute[o : o+4])
		localPart1Value := binary.BigEndian.Uint32(bgpPathAttribute[o+4 : o+8])
		localPart2Value := binary.BigEndian.Uint32(bgpPathAttribute[o+8 : o+12])
		r.LargeCommunities = append(r.LargeCommunities, fmt.Sprintf("%d:%d:%d", asNumber, localPart1Value, localPart2Value))
		o += 12
	}
	return nil
}

func (r *RIBEntry) ReadCommunities(bgpPathAttribute []byte) error {
	// COMMUNITY is a variable-length field
	// The first two bytes are the AS number
	// The next two bytes are the community value
	r.Communities = []string{}
	if len(bgpPathAttribute) < 4 {
		return fmt.Errorf("COMMUNITIES: Invalid length")
	}
	communityCount := len(bgpPathAttribute) / 4 // Each community is 4 bytes
	o := 0
	for i := 0; i < int(communityCount); i++ {
		asn := binary.BigEndian.Uint16(bgpPathAttribute[o : o+2])
		community := binary.BigEndian.Uint16(bgpPathAttribute[o+2 : o+4])
		r.Communities = append(r.Communities, fmt.Sprintf("%d:%d", asn, community))
		o += 4
	}
	return nil
}

func (r *RIBEntry) ReadNextHopV4(bgpPathAttribute []byte) error {
	// NEXT_HOP is a 4-byte field
	// It contains the IPv4 address of the next hop
	if len(bgpPathAttribute) < 4 {
		return fmt.Errorf("NEXT_HOP: Invalid length")
	}
	r.NextHop = net.IP(bgpPathAttribute[0:4])
	return nil
}

func (r *RIBEntry) ReadOrigin(bgpPathAttribute []byte) error {
	const (
		IGP = iota
		EGP
		Incomplete
	)
	originValue := bgpPathAttribute[0]
	switch originValue {
	case IGP:
		r.Origin = "IGP"
	case EGP:
		r.Origin = "EGP"
	case Incomplete:
		r.Origin = "INCOMPLETE"
	default:
		return fmt.Errorf("unknown ORIGIN value: %d", originValue)
	}
	return nil
}

func (r *RIBEntry) ReadASPath(bgpPathAttribute []byte) error {
	// AS_PATH is a variable-length field
	// The first byte is the segment type
	// The second byte is the segment length
	// The remaining bytes are the AS numbers
	//segmentType := bgpPathAttribute[0]
	segmentLength := bgpPathAttribute[1]
	o := 2 // Start after segment type and segment length
	for i := 0; i < int(segmentLength); i++ {
		r.ASPath = append(r.ASPath, binary.BigEndian.Uint32(bgpPathAttribute[o:o+4]))
		o += 4
	}
	return nil
}

func (r *RIBEntry) ReadAggregator(bgpPathAttribute []byte) error {
	// AGGREGATOR is a variable-length field
	// The first four bytes are the AS number
	// The next four bytes are the IP address
	if len(bgpPathAttribute) < 8 {
		return fmt.Errorf("AGGREGATOR: Invalid length")
	}
	asNumber := binary.BigEndian.Uint32(bgpPathAttribute[:4])
	ipAddress := net.IP(bgpPathAttribute[4:8])
	r.Aggregator = fmt.Sprintf("%d %s", asNumber, ipAddress.String())
	return nil
}

func (r *RIBEntry) ReadAttributeByType(attrType byte, attributeBuff []byte) error {
	// This function will read a BGP attribute entry by its type.
	switch attrType {
	case 1: // ORIGIN = 1
		return r.ReadOrigin(attributeBuff)
	case 2: // AS_PATH = 2
		return r.ReadASPath(attributeBuff)
	case 3: // NEXT_HOP = 3
		return r.ReadNextHopV4(attributeBuff)
	case 4: // MULTI_EXIT_DISC = 4
		return r.ReadMultiExitDisc(attributeBuff)
	case 7: // AGGREGATOR = 7
		return r.ReadAggregator(attributeBuff)
	case 8: // COMMUNITY = 8
		return r.ReadCommunities(attributeBuff)
	case 32: // LARGE_COMMUNITIES = 32
		return r.ReadLargeCommunities(attributeBuff)
	}
	return fmt.Errorf("unknown or not implemented BGP attribute type: %d, buffer: %v with length: %d", attrType, attributeBuff, len(attributeBuff))
}

func (r *RIBEntry) Read(buf []byte) (int, error) {
	o := 0
	// PeerIndex is a 2-byte field
	r.PeerIndexId = binary.BigEndian.Uint16(buf[o : o+2])
	o += 2
	// Originated Time
	r.OriginatedTime = binary.BigEndian.Uint32(buf[o : o+4])
	o += 4
	// Attribute Length
	attributeLength := binary.BigEndian.Uint16(buf[o : o+2])
	o += 2
	// BGP Path Attributes
	bgpPathAttributes := buf[o : o+int(attributeLength)]
	var attributesFlag byte
	var attributesType byte
	isExtLength := byte(0x10) // Extended Length flag
	r.MultiExitDisc = -1
	i := 0
	for i < len(bgpPathAttributes) {
		attributesFlag = bgpPathAttributes[i]
		attributesType = bgpPathAttributes[i+1]
		if attributesFlag&isExtLength != 0 {
			r.AttributeLength = binary.BigEndian.Uint16(bgpPathAttributes[i+2 : i+4])
			i += 4
		} else {
			r.AttributeLength = uint16(bgpPathAttributes[i+2])
			i += 3
		}
		err := r.ReadAttributeByType(attributesType, bgpPathAttributes[i:i+int(r.AttributeLength)])
		i += int(r.AttributeLength)
		if err != nil {
			return 0, err
		}
	}
	return o + int(attributeLength), nil
}

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
	t.ReadEntries(buf[o:])
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
