package message

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	ORIGIN            = 1  // ORIGIN
	AS_PATH           = 2  // AS_PATH
	NEXT_HOP          = 3  // NEXT_HOP
	MULTI_EXIT_DISC   = 4  // MULTI_EXIT_DISC
	AGGREGATOR        = 7  // AGGREGATOR
	COMMUNITY         = 8  // COMMUNITY
	LARGE_COMMUNITIES = 32 // LARGE_COMMUNITIES
)

type RIBEntry struct {
	PeerIndex        *MRTPeerIndex
	PeerIndexId      uint16 // Peer index from the MRT header
	OriginatedTime   uint32 // Unix timestamp in seconds
	AttributeLength  uint16
	LargeCommunities []string // Optional: Large Communities
	Communities      []string // Optional Communities associated with the entry
	NextHop          net.IP
	Origin           string
	ASPath           []uint32
	MultiExitDisc    *int32 // Optional: Multi Exit Discriminator nil if not set
	Aggregator       string // Aggregator information
}

func (r *RIBEntry) String() string {
	var sb strings.Builder

	if r.PeerIndex != nil {
		sb.WriteString(fmt.Sprintf("FROM: %s AS%d\n",
			r.PeerIndex.Entries[r.PeerIndexId].PeerIP,
			r.PeerIndex.Entries[r.PeerIndexId].PeerAS))
	} else {
		sb.WriteString(fmt.Sprintf("PeerIndexId: %d\n", r.PeerIndexId))
	}
	sb.WriteString(fmt.Sprintf("ORIGINATED: %s\n", time.Unix(int64(r.OriginatedTime), 0).Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("ORIGIN: %s\n", r.Origin))

	sb.WriteString("ASPATH: ")
	for _, asn := range r.ASPath {
		sb.WriteString(fmt.Sprintf("AS%d ", asn))
	}
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("NEXT_HOP: %s\n", r.NextHop.String()))
	if r.MultiExitDisc != nil {
		sb.WriteString(fmt.Sprintf("MULTI_EXIT_DISC: %d\n", *r.MultiExitDisc))
	}
	if len(r.Communities) > 0 {
		sb.WriteString(fmt.Sprintf("COMMUNITIES: %v\n", strings.Trim(fmt.Sprintf("%v", r.Communities), "[]")))
	}
	if len(r.LargeCommunities) > 0 {
		sb.WriteString(fmt.Sprintf("LARGE_COMMUNITIES: %v\n", strings.Trim(fmt.Sprintf("%v", r.LargeCommunities), "[]")))
	}
	if len(r.Aggregator) > 0 {
		sb.WriteString(fmt.Sprintf("AGGREGATOR: %s\n", r.Aggregator))
	}
	return sb.String()
}

func (r *RIBEntry) ReadMultiExitDisc(bgpPathAttribute []byte) error {
	// MULTI_EXIT_DISC is a 4-byte field
	// It contains the Multi-Exit Discriminator value
	if len(bgpPathAttribute) < 4 {
		return fmt.Errorf("MULTI_EXIT_DISC: Invalid length")
	}
	r.MultiExitDisc = new(int32)
	*r.MultiExitDisc = int32(binary.BigEndian.Uint32(bgpPathAttribute[:4]))
	return nil
}

func (r *RIBEntry) ReadLargeCommunities(bgpPathAttribute []byte) error {
	// LARGE_COMMUNITIES is a variable-length field
	// The first four bytes are the AS number
	// The next four bytes are the community value
	if len(bgpPathAttribute) < 12 || len(bgpPathAttribute)%12 != 0 {
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
	if len(bgpPathAttribute) < 4 || len(bgpPathAttribute)%4 != 0 {
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
	if len(bgpPathAttribute) < 6 {
		return fmt.Errorf("AS_PATH: Invalid length")
	}
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
	case ORIGIN: // ORIGIN = 1
		return r.ReadOrigin(attributeBuff)
	case AS_PATH: // AS_PATH = 2
		return r.ReadASPath(attributeBuff)
	case NEXT_HOP: // NEXT_HOP = 3
		return r.ReadNextHopV4(attributeBuff)
	case MULTI_EXIT_DISC: // MULTI_EXIT_DISC = 4
		return r.ReadMultiExitDisc(attributeBuff)
	case AGGREGATOR: // AGGREGATOR = 7
		return r.ReadAggregator(attributeBuff)
	case COMMUNITY: // COMMUNITY = 8
		return r.ReadCommunities(attributeBuff)
	case LARGE_COMMUNITIES: // LARGE_COMMUNITIES = 32
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
	// BGP Path Attributes buffer
	bgpPathAttributes := buf[o : o+int(attributeLength)]
	var attributesFlag byte
	var attributesType byte
	isExtLength := byte(0x10) // Extended Length flag
	i := 0
	// while loop to read BGP Path Attributes buffer
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
			// If an error occurs while reading an attribute, skip the rest of the attributes
			return i, err
		}
	}
	return o + int(attributeLength), nil
}
