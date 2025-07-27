package mrtheader

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"time"
)

type MRTHeader struct {
	Ts      uint32 // "Timestamp" in seconds since epoch"
	Type    uint16 // Type of the MRT header message
	Subtype uint16 // Subtype of the MRT header message
	Length  uint32 // Length of the MRT header message
}

func NewMRTHeader() *MRTHeader {
	return &MRTHeader{}
}

func (h *MRTHeader) Parse(f *os.File) (*MRTHeader, []byte, error) {
	// MRT Types
	const (
		OSPFv2        = 11
		TABLE_DUMP    = 12
		TABLE_DUMP_V2 = 13
		BGP4MP        = 16
		BGP4MP_ET     = 17
		ISIS          = 32
		ISIS_ET       = 33
		OSPFv3        = 48
		OSPFv3_ET     = 49
	)
	var mrtTypes = []uint16{OSPFv2, TABLE_DUMP, TABLE_DUMP_V2, BGP4MP, BGP4MP_ET, ISIS, ISIS_ET, OSPFv3, OSPFv3_ET}
	var mrtTypesSupported = []uint16{TABLE_DUMP_V2}
	err := binary.Read(f, binary.BigEndian, h)
	if err != nil {
		if errors.Is(err, io.EOF) {
			os.Exit(0) // End of file reached, exit gracefully
		}
		return nil, nil, fmt.Errorf("failed to read header: %w", err)
	}
	if !slices.Contains(mrtTypes, h.Type) {
		return nil, nil, fmt.Errorf("unknown MRT type: %d", h.Type)
	}
	if !slices.Contains(mrtTypesSupported, h.Type) {
		return nil, nil, fmt.Errorf("MRT type %d not implemented", h.Type)
	}
	buf := make([]byte, h.Length)
	if err = binary.Read(f, binary.BigEndian, &buf); err != nil {
		return nil, nil, fmt.Errorf("failed to read packet: %w", err)
	}
	return h, buf, nil
}

func (h *MRTHeader) String() string {
	return fmt.Sprintf("MRTHeader Ts: %v, Type: %d, Subtype: %d, Length: %d", time.Unix(int64(h.Ts), 0), h.Type, h.Subtype, h.Length)
}
