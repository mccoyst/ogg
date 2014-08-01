// © 2014 Steve McCoy under the MIT license. See LICENSE for details.

package ogg

import (
	"encoding/binary"
	"hash/crc32"
)

const MIMEType = "application/ogg"

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
