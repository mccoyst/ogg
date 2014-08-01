// © 2014 Steve McCoy under the MIT license. See LICENSE for details.

/*
Package ogg implements encoding and decoding of OGG streams as defined in
http://xiph.org/ogg/doc/rfc3533.txt
and
http://xiph.org/ogg/doc/framing.html .
*/
package ogg

import (
	"encoding/binary"
	"hash/crc32"
)

const MIMEType = "application/ogg"

const mps = 255 * 255
// header + max segment table of 255 + max 255 segments of 255 bytes in a packet
// == 65307, per the RFC
const maxPageSize = 27 + 255 + mps

var ByteOrder = binary.LittleEndian

type pageHeader struct {
	OggS          [4]byte // 0-3, always == "OggS"
	StreamVersion byte    // 4, always == 0
	HeaderType    byte    // 5
	Granule       int64   // 6-13, codec-specific
	Serial        uint32  // 14-17, associated with a logical stream
	Page          uint32  // 18-21, sequence number of page in packet
	Crc           uint32  // 22-25
	Nsegs         byte    // 26
}

const (
	// Continuation of packet
	cop byte = 1 + iota
	// Beginning of stream
	bos = 1 << iota
	// End of stream
	eos = 1 << iota
)

var crcTable = crc32.MakeTable(0x04c11db7)
