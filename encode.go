// © 2016 Steve McCoy under the MIT license. See LICENSE for details.

package ogg

import (
	"bytes"
	"encoding/binary"
	"io"
)

// An Encoder encodes raw bytes into an ogg stream.
type Encoder struct {
	serial uint32
	page   uint32
	dummy  [1][]byte // convenience field to handle nil packets args without allocating
	w      io.Writer
	buf    [maxPageSize]byte
}

// NewEncoder creates an ogg encoder with the given serial ID.
// Multiple Encoders can be used to encode multiplexed logical streams
// by giving them distinct IDs. Users must be sure to encode the streams
// as specified by the ogg RFC:
// When Grouping, all BOS pages must come before the data
// and EOS pages, with the order of BOS pages defined by the encapsulated encoding.
// When Chaining, the EOS page of the first stream must be immediately followed by
// the BOS of the second stream, and so on.
//
// For more details, see
// http://xiph.org/ogg/doc/rfc3533.txt and
// http://xiph.org/ogg/doc/framing.html
func NewEncoder(id uint32, w io.Writer) *Encoder {
	return &Encoder{serial: id, w: w}
}

// EncodeBOS writes a beginning-of-stream packet to the ogg stream,
// using the provided granule position.
// If the packets are larger than can fit in a page, the payload is split into multiple
// pages with the continuation-of-packet flag set.
// Packets can be empty or nil, in which one segment of size 0 is encoded.
func (w *Encoder) EncodeBOS(granule int64, packets [][]byte) error {
	if len(packets) == 0 {
		packets = w.dummy[:]
	}
	return w.writePackets(BOS, granule, packets)
}

// Encode writes a data packet to the ogg stream,
// using the provided granule position.
// If the packet is larger than can fit in a page, it is split into multiple
// pages with the continuation-of-packet flag set.
// Packets can be empty or nil, in which one segment of size 0 is encoded.
func (w *Encoder) Encode(granule int64, packets [][]byte) error {
	if len(packets) == 0 {
		packets = w.dummy[:]
	}
	return w.writePackets(0, granule, packets)
}

// EncodeEOS writes an end-of-stream packet to the ogg stream.
// Packets can be empty or nil, in which one segment of size 0 is encoded.
func (w *Encoder) EncodeEOS(granule int64, packets [][]byte) error {
	if len(packets) == 0 {
		packets = w.dummy[:]
	}
	return w.writePackets(EOS, granule, packets)
}

func (w *Encoder) writePackets(kind byte, granule int64, packets [][]byte) error {
	h := pageHeader{
		OggS:       [4]byte{'O', 'g', 'g', 'S'},
		HeaderType: kind,
		Serial:     w.serial,
		Granule:    granule,
	}

	// Write the lacing values before filling in their quantity
	segtbl, car, cdr := w.segmentize(payload{packets[0], packets[1:], nil})
	err := w.writePage(&h, segtbl, car)
	if err != nil {
		return err
	}

	h.HeaderType |= COP
	for len(cdr.leftover) > 0 {
		segtbl, car, cdr = w.segmentize(cdr)
		err = w.writePage(&h, segtbl, car)
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *Encoder) writePage(h *pageHeader, segtbl []byte, pay payload) error {
	h.Page = w.page
	w.page++
	h.Nsegs = byte(len(segtbl))
	hb := bytes.NewBuffer(w.buf[0:0:cap(w.buf)])
	_ = binary.Write(hb, byteOrder, h)

	// segtbl is already written in the buffer,
	// but the writer needs to move along anyhow
	hb.Write(segtbl)

	hb.Write(pay.leftover)
	for _, p := range pay.packets {
		hb.Write(p)
	}
	hb.Write(pay.rightover)

	bb := hb.Bytes()
	crc := crc32(bb)
	_ = binary.Write(bytes.NewBuffer(bb[22:22:26]), byteOrder, crc)

	_, err := hb.WriteTo(w.w)
	return err
}

// payload represents a potentially-split group of packets.
// For the "left" portion of a split,
// rightover is the beginning portion of the *last* packet,
// and packets contains the preceding packets.
// For the "right" portion of a split,
// leftover is the *first* packet and the other packets follow.
//
// ASCII example (each run of letters represents one packet):
//
// Page 1 (left)  Page 2 (right)
// [aaaabbbbccccd][dddeeeffff]
//
// For Page 1, packets would be a slice holding the a's, b's, and c's.
// and rightover would contain the first d.
// For Page 2, leftover would contain the d's,
// and packets would contain the e's and f's
type payload struct {
	leftover  []byte
	packets   [][]byte
	rightover []byte
}

// segmentize fills the segment table with lacing values based on the packets
// provided in payload, starting with leftover (if any).
// It returns the segment table (sized appropriately),
// the payload to write with the segment table in the current page,
// and any leftover payload that remains due to not fitting in a page.
func (w *Encoder) segmentize(pay payload) ([]byte, payload, payload) {
	segtbl := w.buf[headsz : headsz+mss]
	i := 0

	s255s := len(pay.leftover) / mss
	rem := len(pay.leftover) % mss
	for i < len(segtbl) && s255s > 0 {
		segtbl[i] = mss
		i++
		s255s--
	}
	if i < mss {
		segtbl[i] = byte(rem)
		i++
	} else {
		leftStart := len(pay.leftover) - (s255s * mss) - rem
		good := payload{pay.leftover[0:leftStart], nil, nil}
		bad := payload{pay.leftover[leftStart:], pay.packets, nil}
		return segtbl, good, bad
	}

	// Now loop through the rest and track if we need to split
	for p := 0; p < len(pay.packets); p++ {
		s255s := len(pay.packets[p]) / mss
		rem := len(pay.packets[p]) % mss
		for i < len(segtbl) && s255s > 0 {
			segtbl[i] = mss
			i++
			s255s--
		}
		if i < mss {
			segtbl[i] = byte(rem)
			i++
		} else {
			right := len(pay.packets[p]) - (s255s * mss) - rem
			good := payload{pay.leftover, pay.packets[0:p], pay.packets[p][0:right]}
			bad := payload{pay.packets[p][right:], pay.packets[p+1:], nil}
			return segtbl, good, bad
		}
	}

	good := pay
	bad := payload{}
	return segtbl[0:i], good, bad
}
