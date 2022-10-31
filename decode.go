// © 2016 Steve McCoy under the MIT license. See LICENSE for details.

package ogg

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"strconv"
)

// A Decoder decodes an ogg stream page-by-page with its Decode method.
type Decoder struct {
	r   io.Reader
	buf [maxPageSize]byte
}

// NewDecoder creates an ogg Decoder.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// A Page represents a logical ogg page.
type Page struct {
	// Type is a bitmask of COP, BOS, and/or EOS.
	Type byte
	// Serial is the bitstream serial number.
	Serial uint32
	// Granule is the granule position, whose meaning is dependent on the encapsulated codec.
	Granule int64
	// Packets are the raw packet data.
	// If Type & COP != 0, the first element is
	// a continuation of the previous page's last packet.
	Packets [][]byte
}

// ErrBadSegs is the error used when trying to decode a page with a segment table size less than 1.
var ErrBadSegs = errors.New("invalid segment table size")

// ErrBadCrc is the error used when an ogg page's CRC field does not match the CRC calculated by the Decoder.
type ErrBadCrc struct {
	Found    uint32
	Expected uint32
}

func (bc ErrBadCrc) Error() string {
	return "invalid crc in packet: got " + strconv.FormatInt(int64(bc.Found), 16) +
		", expected " + strconv.FormatInt(int64(bc.Expected), 16)
}

var oggs = []byte{'O', 'g', 'g', 'S'}

// Decode reads from d's Reader to the next ogg page, then returns the decoded Page or an error.
// The error may be io.EOF if that's what the Reader returned.
//
// The buffer underlying the returned Page's Packets' bytes is owned by the Decoder.
// It may be overwritten by subsequent calls to Decode.
//
// It is safe to call Decode concurrently on distinct Decoders if their Readers are distinct.
// Otherwise, the behavior is undefined.
func (d *Decoder) Decode() (Page, error) {
	hbuf := d.buf[0:headsz]
	b := 0
	for {
		_, err := io.ReadFull(d.r, hbuf[b:])
		if err != nil {
			return Page{}, err
		}

		i := bytes.Index(hbuf, oggs)
		if i == 0 {
			break
		}

		if i < 0 {
			const n = headsz
			if hbuf[n-1] == 'O' {
				i = n - 1
			} else if hbuf[n-2] == 'O' && hbuf[n-1] == 'g' {
				i = n - 2
			} else if hbuf[n-3] == 'O' && hbuf[n-2] == 'g' && hbuf[n-1] == 'g' {
				i = n - 3
			}
		}

		if i > 0 {
			b = copy(hbuf, hbuf[i:])
		}
	}

	var h pageHeader
	_ = binary.Read(bytes.NewBuffer(hbuf), byteOrder, &h)

	if h.Nsegs < 1 {
		return Page{}, ErrBadSegs
	}

	nsegs := int(h.Nsegs)
	segtbl := d.buf[headsz : headsz+nsegs]
	_, err := io.ReadFull(d.r, segtbl)
	if err != nil {
		return Page{}, err
	}

	// A page can contain multiple packets; record their lengths from the table
	// now and slice up the payload after reading it.
	// I'm inclined to limit the Read calls this way,
	// but it's possible it isn't worth the annoyance of iterating twice
	var packetlens []int
	payloadlen := 0
	more := false
	for _, l := range segtbl {
		if more {
			packetlens[len(packetlens)-1] += int(l)
		} else {
			packetlens = append(packetlens, int(l))
		}

		more = l == mss
		payloadlen += int(l)
	}

	payload := d.buf[headsz+nsegs : headsz+nsegs+payloadlen]
	_, err = io.ReadFull(d.r, payload)
	if err != nil {
		return Page{}, err
	}

	page := d.buf[0 : headsz+nsegs+payloadlen]
	// Clear out existing crc before calculating it
	page[22] = 0
	page[23] = 0
	page[24] = 0
	page[25] = 0
	crc := crc32(page)
	if crc != h.Crc {
		return Page{}, ErrBadCrc{h.Crc, crc}
	}

	packets := make([][]byte, len(packetlens))
	s := 0
	for i, l := range packetlens {
		packets[i] = payload[s : s+l ]
		s += l
	}

	return Page{h.HeaderType, h.Serial, h.Granule, packets}, nil
}
