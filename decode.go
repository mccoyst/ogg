// © 2014 Steve McCoy under the MIT license. See LICENSE for details.

package ogg

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
)

type Decoder struct {
	r io.Reader
	buf [maxPageSize]byte
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

type Page struct {
	Type byte
	Serial uint32
	Granule int64
	Packet []byte
}

var ErrBadSegs = errors.New("invalid segment table size")
var ErrBadCrc = errors.New("invalid crc in packet")

var oggs = []byte{ 'O', 'g', 'g', 'S' }

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
			if hbuf[headsz-1] == 'O' {
				i = headsz-1
			} else if  hbuf[headsz-2] == 'O' && hbuf[headsz-1] == 'g' {
				i = headsz-2
			} else if hbuf[headsz-3] == 'O' && hbuf[headsz-2] == 'g' && hbuf[headsz-1] == 'g' {
				i = headsz-3
			}
		}

		if i > 0 {
			b = copy(hbuf, hbuf[i:])
		}
	}

	var h pageHeader
	err := binary.Read(bytes.NewBuffer(hbuf), ByteOrder, &h)
	if err != nil {
		return Page{}, err
	}

	if h.Nsegs < 1 {
		return Page{}, ErrBadSegs
	}

	segtbl := d.buf[headsz:headsz+h.Nsegs]
	_, err = io.ReadFull(d.r, segtbl)
	if err != nil {
		return Page{}, err
	}

	packetlen := int(mss*(h.Nsegs-1) + segtbl[h.Nsegs-1])
	packet := d.buf[headsz+h.Nsegs:headsz+int(h.Nsegs)+packetlen]
	_, err = io.ReadFull(d.r, packet)
	if err != nil {
		return Page{}, err
	}

	page := d.buf[0:headsz+int(h.Nsegs)+packetlen]
	// Clear out existing crc before calculating it
	page[22] = 0
	page[23] = 0
	page[24] = 0
	page[25] = 0
	crc := crc32.Checksum(page, crcTable)
	if crc != h.Crc {
		return Page{}, ErrBadCrc
	}

	return Page{h.HeaderType, h.Serial, h.Granule, packet}, nil
}
