// © 2014 Steve McCoy under the MIT license. See LICENSE for details.

package ogg

import (
	"encoding/binary"
	"errors"
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

func (d *Decoder) Decode() (Page, error) {
	//BUG(mccoyst): validate checksum
	var h pageHeader
	err := binary.Read(d.r, ByteOrder, &h)
	if err != nil {
		return Page{}, err
	}

	if h.Nsegs < 1 {
		return Page{}, ErrBadSegs
	}

	_, err = io.ReadFull(d.r, d.buf[0:h.Nsegs])
	if err != nil {
		return Page{}, err
	}

	packetlen := int(255*(h.Nsegs-1) + d.buf[h.Nsegs-1])
	packet := d.buf[0:packetlen]
	_, err = io.ReadFull(d.r, packet)
	if err != nil {
		return Page{}, err
	}

	return Page{h.HeaderType, h.Serial, h.Granule, packet}, nil
}
