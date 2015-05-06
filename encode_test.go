// © 2014 Steve McCoy under the MIT license. See LICENSE for details.

package ogg

import (
	"bytes"
	"testing"
)

func TestBasicEncodeBOS(t *testing.T) {
	var b bytes.Buffer
	e := NewEncoder(1, &b)

	err := e.EncodeBOS(2, []byte("hello"))
	if err != nil {
		t.Fatal("unexpected EncodeBOS error:", err)
	}

	bb := b.Bytes()
	expect := []byte{
		'O', 'g', 'g', 'S',
		0,
		BOS,
		2, 0, 0, 0, 0, 0, 0, 0,
		1, 0, 0, 0,
		0, 0, 0, 0,
		0x7e, 0xdf, 0x2e, 0x1e, // crc
		1,
		5, // segment table
		'h', 'e', 'l', 'l', 'o',
	}

	if !bytes.Equal(bb, expect) {
		t.Fatalf("bytes != expected:\n%x\n%x", bb, expect)
	}
}

func TestBasicEncode(t *testing.T) {
	var b bytes.Buffer
	e := NewEncoder(1, &b)

	err := e.Encode(2, []byte("hello"))
	if err != nil {
		t.Fatal("unexpected EncodeBOS error:", err)
	}

	bb := b.Bytes()
	expect := []byte{
		'O', 'g', 'g', 'S',
		0,
		0,
		2, 0, 0, 0, 0, 0, 0, 0,
		1, 0, 0, 0,
		0, 0, 0, 0,
		0xc8, 0x21, 0xcc, 0x1c, // crc
		1,
		5, // segment table
		'h', 'e', 'l', 'l', 'o',
	}

	if !bytes.Equal(bb, expect) {
		t.Fatalf("bytes != expected:\n%x\n%x", bb, expect)
	}
}
