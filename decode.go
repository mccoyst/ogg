// © 2014 Steve McCoy under the MIT license. See LICENSE for details.

package ogg

import (
	"bytes"
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

var oggs = []byte{ 'O', 'g', 'g', 'S' }

func (d *Decoder) Decode() (Page, error) {
	buf := d.buf[0:27]
	b := 0
	for {
		_, err := io.ReadFull(d.r, buf[b:])
		if err != nil {
			return Page{}, err
		}

		i := bytes.Index(buf, oggs)
		if i == 0 {
			break
		}

		if i < 0 {
			if buf[26] == 'O' {
				i = 26
			} else if  buf[25] == 'O' && buf[26] == 'g' {
				i = 25
			} else if buf[24] == 'O' && buf[25] == 'g' && buf[26] == 'g' {
				i = 24
			}
		}

		if i > 0 {
			b = copy(buf, buf[i:])
		}
	}

	//BUG(mccoyst): validate checksum
	var h pageHeader
	err := binary.Read(bytes.NewBuffer(buf), ByteOrder, &h)
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
