package cmd

import (
	"fmt"
	"io/fs"
	"log"

	me "github.com/vgauthier/mrtdump/internal/message"
	he "github.com/vgauthier/mrtdump/internal/mrtheader"
)

func NewReadFileOptions(fileSystem fs.FS, fileName string) *ReadFileOptions {
	return &ReadFileOptions{
		FileSystem: fileSystem,
		FileName:   fileName,
		PeerIndex:  nil, // Peer index can be set later if needed
	}
}

type ReadFileOptions struct {
	FileSystem     fs.FS
	FileName       string
	PeerIndex      *me.MRTPeerIndex
	FileDescriptor fs.File
}

func (rf *ReadFileOptions) parseMessage(f fs.File) (*me.MRTMessage, error) {
	// Read the MRT header
	headerBuf, err := rf.readHeader(f)
	if err != nil {
		return nil, err
	}

	// Parse the MRT header
	header, err := he.NewMRTHeader().Parse(headerBuf)
	if err != nil {
		return nil, err
	}

	// Read the MRT message
	messageBuf, err := readMessage(f, header.Length)
	if err != nil {
		return nil, err
	}
	// Parse the MRT message
	var message *me.MRTMessage
	if rf.PeerIndex == nil {
		// If no peer index is provided, we assume the first message is a peer index
		message, err = me.NewMRTMessage(header).Parse(messageBuf)
	} else {
		message, err = me.NewMRTMessage(header).WithPeerIndex(rf.PeerIndex).Parse(messageBuf)
	}
	if err != nil {
		return nil, err
	}

	return message, nil
}

// ReadFile reads an MRT file and return a raw bytes buffer that can be used to parse the MRT Message.
func readMessage(f fs.File, size uint32) ([]byte, error) {
	buf := make([]byte, size)
	_, err := f.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// ReadFile reads an MRT file and return a raw bytes buffer that can be used to parse the MRT header.
func (rf *ReadFileOptions) readHeader(f fs.File) ([]byte, error) {
	const headerSize = 12 // Size of the MRT header in bytes

	buf := make([]byte, headerSize)
	// Read the first 10 bytes of the file to get the MRT header
	_, err := f.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (rf *ReadFileOptions) ReadFile() {
	// Open the file
	var err error
	rf.FileDescriptor, err = rf.FileSystem.Open(rf.FileName)
	if err != nil {
		log.Fatalln(err)
		return
	}
	defer rf.FileDescriptor.Close()

	// Read the first message (peer index)
	peerIndex, err := rf.parseMessage(rf.FileDescriptor)
	if err != nil {
		log.Fatalln(fmt.Errorf("failed to parse MRT message: %w", err))
	}
	m, err := peerIndex.GetMessage()
	if err != nil {
		log.Fatalln(fmt.Errorf("failed to get MRT message: %w", err))
	}
	rf.PeerIndex = m.(*me.MRTPeerIndex)
	fmt.Printf("%s\n", rf.PeerIndex.String())
	// If the peer index is not nil, we can use it to parse subsequent messages
	for i := 0; i < 2; i++ {
		rib, err := rf.parseMessage(rf.FileDescriptor)
		if err != nil {
			log.Fatalln(fmt.Errorf("failed to parse MRT: %w", err))
			return
		}
		if rib.Err != nil {
			log.Fatalln(fmt.Errorf("error parsing MRT message: %w", rib.Err))
		} else {
			fmt.Printf("%s\n", rib.Message.String())
		}
	}
}
