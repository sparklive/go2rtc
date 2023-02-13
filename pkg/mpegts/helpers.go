package mpegts

import (
	"github.com/AlexxIT/go2rtc/pkg/h264"
	"github.com/AlexxIT/go2rtc/pkg/streamer"
	"time"
)

const (
	PacketSize = 188
	SyncByte   = 0x47
)

const (
	StreamTypeH264 = 0x1B
	StreamTypePCMA = 0x90
)

type Packet struct {
	StreamType byte
	PTS        time.Duration
	DTS        time.Duration
	Payload    []byte
}

// PES - Packetized Elementary Stream
type PES struct {
	StreamType byte
	StreamID   byte
	Payload    []byte
}

func (p *PES) Packet() *Packet {
	// parse Optional PES header
	const minHeaderSize = 3

	pkt := &Packet{StreamType: p.StreamType}

	// fist byte also flags
	flags := p.Payload[1]
	hSize := p.Payload[2] // optional fields

	const hasPTS = 0b1000_0000
	if flags&hasPTS != 0 {
		pkt.PTS = ParseTime(p.Payload[minHeaderSize:])

		const hasDTS = 0b0100_0000
		if flags&hasDTS != 0 {
			pkt.DTS = ParseTime(p.Payload[minHeaderSize+5:])
		}
	}

	pkt.Payload = p.Payload[minHeaderSize+hSize:]

	return pkt
}

func ParseTime(b []byte) time.Duration {
	ts := (uint64(b[0]) >> 1 & 0x7 << 30) | (uint64(b[1]) << 22) | (uint64(b[2]) >> 1 & 0x7F << 15) | (uint64(b[3]) << 7) | (uint64(b[4]) >> 1 & 0x7F)
	return time.Duration(ts)
}

func GetMedia(pkt *Packet) *streamer.Media {
	var codec *streamer.Codec
	var kind string

	switch pkt.StreamType {
	case StreamTypeH264:
		codec = &streamer.Codec{
			Name:        streamer.CodecH264,
			ClockRate:   90000,
			PayloadType: streamer.PayloadTypeRAW,
			FmtpLine:    h264.GetFmtpLine(pkt.Payload),
		}
		kind = streamer.KindVideo

	case StreamTypePCMA:
		codec = &streamer.Codec{
			Name:      streamer.CodecPCMA,
			ClockRate: 8000,
		}
		kind = streamer.KindAudio

	default:
		return nil
	}

	return &streamer.Media{
		Kind:      kind,
		Direction: streamer.DirectionSendonly,
		Codecs:    []*streamer.Codec{codec},
	}
}
