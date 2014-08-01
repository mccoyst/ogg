// © 2014 Steve McCoy under the MIT license. See LICENSE for details.

/*
Package ogg implements encoding and decoding of OGG streams as defined in
http://xiph.org/ogg/doc/rfc3533.txt
and
http://xiph.org/ogg/doc/framing.html .
*/
package ogg

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"io"
)

type Encoder struct {
	serial uint32
	w      io.Writer
}

func NewEncoder(id uint32, w io.Writer) *Encoder {
	return &Encoder{id, w}
}

func (w *Encoder) EncodeBOS(granule int64, packet []byte) error {
	_, err := w.writePacket(bos, granule, packet)
	return err
}

func (w *Encoder) Encode(granule int64, packet []byte) (int, error) {
	return w.writePacket(0, granule, packet)
}

func (w *Encoder) EncodeEOS() error {
	_, err := w.writePacket(eos, 0, nil)
	return err
}

func (w *Encoder) writePacket(kind byte, granule int64, packet []byte) (int, error) {
	h := pageHeader{
		OggS:       [4]byte{'O', 'g', 'g', 'S'},
		HeaderType: kind,
		Serial:     w.serial,
		Granule: granule,
	}

	var err error
	n, m := 0, 0
	const mps = 255 * 255 // maximum 255 segments of 255 bytes in a page

	s := 0
	e := s + mps
	if e > len(packet) {
		e = len(packet)
	}
	page := packet[s:e]
	n, err = w.writePage(page, &h)
	if err != nil {
		return n, err
	}
	s = e

	last := (len(packet) / mps) * mps
	h.HeaderType &= cop
	for s < last {
		h.Page++
		e = s + mps
		page = packet[s:e]
		m, err = w.writePage(page, &h)
		n += m
		if err != nil {
			return n, err
		}
		s = e
	}

	if s != len(packet) {
		m, err = w.writePage(packet[s:], &h)
		n += m
	}
	return n, err
}

func (w *Encoder) writePage(page []byte, h *pageHeader) (int, error) {
	h.Nsegs = byte(len(page)/255 + 1)
	segtbl := make([]byte, h.Nsegs)
	for i := 0; i < len(segtbl)-1; i++ {
		segtbl[i] = 255
	}
	segtbl[len(segtbl)-1] = byte(len(page) % 255)

	var hb bytes.Buffer
	err := binary.Write(&hb, ByteOrder, h)
	if err != nil {
		return 0, err
	}

	hb.Write(segtbl)
	hb.Write(page)

	bb := hb.Bytes()
	crc := crc32.Checksum(bb, crcTable)
	err = binary.Write(bytes.NewBuffer(bb[22:22:26]), ByteOrder, crc)
	if err != nil {
		return 0, nil
	}

	n64, err := hb.WriteTo(w.w)
	return int(n64), err
}
