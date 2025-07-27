package cmd

import (
	"fmt"
	"log"
	"os"

	me "github.com/vgauthier/mrtdump/internal/message"
)

func ReadFile(filepath string) {
	f, err := os.Open(filepath)
	if err != nil {
		log.Fatalln(err)
		return
	}
	defer f.Close()

	// first message is PeerIndex
	peerIndex, err := me.NewMRTMessage().Parse(f)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("%s\n", peerIndex.Message.String())
	message := peerIndex.GetMessage()
	p, ok := message.(*me.MRTPeerIndex)
	if !ok {
		log.Fatalln("Failed to cast message to MRTPeerIndex")
	}

	for {
		// second packet is a RIB
		message, err := me.NewMRTMessage().WithPeerIndex(p).Parse(f)
		if err != nil {
			log.Fatalln(fmt.Errorf("failed to parse MRT: %w", err))
			return
		}
		if message.Err != nil {
			log.Fatalln(fmt.Errorf("error parsing MRT message: %w", message.Err))
		} else {
			fmt.Printf("%s\n", message.Message.String())
		}
	}
}
