package main

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"ipc-recorder/schema"
	"os"
	"time"

	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
	"github.com/bluenviron/mediacommon/pkg/formats/mpegts"
)

func durationGoToMPEGTS(v time.Duration) int64 {
	return int64(v.Seconds() * 90000)
}

// mpegtsMuxer allows to save a H264 stream into a MPEG-TS file.
type mpegtsMuxer struct {
	fileName string
	sps      []byte
	pps      []byte

	f            *os.File
	b            *bufio.Writer
	w            *mpegts.Writer
	track        *mpegts.Track
	dtsExtractor *h264.DTSExtractor

	lastEpoch      int
	sliceStartTime time.Time
}

// initialize initializes a mpegtsMuxer.
func (e *mpegtsMuxer) initialize(filename string) error {
	e.close()

	e.fileName = filename

	var err error
	e.f, err = os.Create(e.fileName)
	if err != nil {
		return err
	}

	if e.b == nil {
		e.b = bufio.NewWriter(e.f)
	} else {
		e.b.Reset(e.f)
	}

	if e.track == nil {
		e.track = &mpegts.Track{
			Codec: &mpegts.CodecH264{},
		}
	}

	e.w = mpegts.NewWriter(e.b, []*mpegts.Track{e.track})
	return nil
}

// close closes all the mpegtsMuxer resources.
func (e *mpegtsMuxer) close() {
	if e.b != nil {
		e.b.Flush()
		e.b = nil
	}
	if e.f != nil {
		e.f.Close()
		e.f = nil
	}
}

// writeH264 writes a H264 access unit into MPEG-TS.
func (e *mpegtsMuxer) writeH264(au [][]byte, pts time.Duration, currentTime time.Time) error {
	// prepend an AUD. This is required by some players
	filteredAU := [][]byte{
		{byte(h264.NALUTypeAccessUnitDelimiter), 240},
	}

	nonIDRPresent := false
	idrPresent := false

	for _, nalu := range au {
		typ := h264.NALUType(nalu[0] & 0x1F)
		switch typ {
		case h264.NALUTypeSPS:
			e.sps = nalu
			continue

		case h264.NALUTypePPS:
			e.pps = nalu
			continue

		case h264.NALUTypeAccessUnitDelimiter:
			continue

		case h264.NALUTypeIDR:
			idrPresent = true

		case h264.NALUTypeNonIDR:
			nonIDRPresent = true
		}

		filteredAU = append(filteredAU, nalu)
	}

	au = filteredAU

	if len(au) <= 1 || (!nonIDRPresent && !idrPresent) {
		return nil
	}

	if e.w == nil && !idrPresent {
		// drop non-IDR frames until we find one
		return nil
	}

	// add SPS and PPS before access unit that contains an IDR
	if idrPresent {
		au = append([][]byte{e.sps, e.pps}, au...)
	}

	var dts time.Duration

	if e.dtsExtractor == nil {
		// skip samples silently until we find one with a IDR
		if !idrPresent {
			return nil
		}
		e.dtsExtractor = h264.NewDTSExtractor()
	}

	var err error
	dts, err = e.dtsExtractor.Extract(au, pts)
	if err != nil {
		return err
	}

	if idrPresent {
		epoch := int(currentTime.Unix()) / 10
		if e.lastEpoch != epoch {
			if e.fileName != "" {
				e.close()
				fi, _ := os.Stat(e.fileName)
				// create record
				CreateRecord(&schema.Record{
					Device:   "ipc01",
					Stream:   "1",
					FileURL:  e.fileName,
					StartAt:  e.sliceStartTime,
					EndAt:    currentTime,
					Interval: currentTime.Sub(e.sliceStartTime).Seconds(),
					FileSize: fi.Size(),
				})
			}

			hash := md5.Sum([]byte(fmt.Sprintf("%s-%s-%d", "ipc01", "1", epoch)))
			hashHex := fmt.Sprintf("%x", hash)
			filename := fmt.Sprintf("data/%s.ts", hashHex)
			if err := e.initialize(filename); err != nil {
				fmt.Println("error initializing mpegtsMuxer")
				return err
			}
			e.lastEpoch = epoch
			e.sliceStartTime = currentTime

			fmt.Println("create new file: ", filename)
		}
	}

	// encode into MPEG-TS
	return e.w.WriteH26x(e.track, durationGoToMPEGTS(pts), durationGoToMPEGTS(dts), idrPresent, au)
}
