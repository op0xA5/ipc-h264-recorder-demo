package main

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/bluenviron/gortsplib/v4"
	"github.com/bluenviron/gortsplib/v4/pkg/base"
	"github.com/bluenviron/gortsplib/v4/pkg/format"
	"github.com/bluenviron/gortsplib/v4/pkg/format/rtph264"
	"github.com/pion/rtp"
)

// This example shows how to
// 1. connect to a RTSP server
// 2. check if there's a H264 format
// 3. save the content of the format into a file in MPEG-TS format

func main() {
	if err := initStorage(); err != nil {
		panic(err)
	}

	c := gortsplib.Client{}
	transport := gortsplib.TransportTCP
	c.Transport = &transport

	// parse URL
	u, err := base.ParseURL("rtsp://admin:123456@192.168.1.191/Stream/Channel/102")
	if err != nil {
		panic(err)
	}

	// connect to the server
	err = c.Start(u.Scheme, u.Host)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	// find available medias
	desc, _, err := c.Describe(u)
	if err != nil {
		panic(err)
	}

	for _, media := range desc.Medias {
		fmt.Println(media)
		for _, f := range media.Formats {
			fmt.Println("format: ", reflect.TypeOf(f))
		}
	}

	// find the H264 media and format
	var forma *format.H264
	medi := desc.FindFormat(&forma)
	if medi == nil {
		panic("media not found")
	}

	// setup RTP/H264 -> H264 decoder
	rtpDec, err := forma.CreateDecoder()
	if err != nil {
		panic(err)
	}

	// setup H264 -> MPEG-TS muxer
	mpegtsMuxer := &mpegtsMuxer{
		sps: forma.SPS,
		pps: forma.PPS,
	}
	if err != nil {
		panic(err)
	}

	// setup a single media
	_, err = c.Setup(desc.BaseURL, medi, 0, 0)
	if err != nil {
		panic(err)
	}

	startTime := time.Now()

	// called when a RTP packet arrives
	c.OnPacketRTP(medi, forma, func(pkt *rtp.Packet) {
		// decode timestamp
		pts, ok := c.PacketPTS(medi, pkt)
		if !ok {
			log.Printf("waiting for timestamp")
			return
		}

		// extract access unit from RTP packets
		au, err := rtpDec.Decode(pkt)
		if err != nil {
			if err != rtph264.ErrNonStartingPacketAndNoPrevious && err != rtph264.ErrMorePacketsNeeded {
				log.Printf("ERR: %v", err)
			}
			return
		}

		// encode the access unit into MPEG-TS
		err = mpegtsMuxer.writeH264(au, pts, startTime.Add(pts))
		if err != nil {
			log.Printf("ERR: %v", err)
			return
		}
	})

	// start playing
	_, err = c.Play(nil)
	if err != nil {
		panic(err)
	}

	// wait until a fatal error
	panic(c.Wait())
}
