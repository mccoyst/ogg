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

const headsz = 27
// max segment size
const mss = 255
// max packet size
const mps = mss * 255
// == 65307, per the RFC
const maxPageSize = headsz + mss + mps

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
	COP byte = 1 + iota
	// Beginning of stream
	BOS = 1 << iota
	// End of stream
	EOS = 1 << iota
)

var crcTable = crc32.MakeTable(0x04c11db7)
