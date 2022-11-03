// © 2016 Steve McCoy under the MIT license. See LICENSE for details.

package ogg

import (
	"bytes"
	"io"
	"testing"
)

func TestBasicEncodeBOS(t *testing.T) {
	var b bytes.Buffer
	e := NewEncoder(1, &b)

	err := e.EncodeBOS(2, [][]byte{[]byte("hello")})
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

func TestEmptyEncodeBOS(t *testing.T) {
	var b bytes.Buffer
	e := NewEncoder(1, &b)

	err := e.EncodeBOS(2, nil)
	if err != nil {
		t.Fatal("unexpected Encode error:", err)
	}

	bb := b.Bytes()
	expect := []byte{
		'O', 'g', 'g', 'S',
		0,
		BOS,
		2, 0, 0, 0, 0, 0, 0, 0,
		1, 0, 0, 0,
		0, 0, 0, 0,
		0x7a, 0xa4, 0x84, 0xc2, // crc
		1,
		0, // segment table
	}

	if !bytes.Equal(bb, expect) {
		t.Fatalf("bytes != expected:\n%x\n%x", bb, expect)
	}
}

func TestEmptyEncode(t *testing.T) {
	var b bytes.Buffer
	e := NewEncoder(1, &b)

	err := e.Encode(2, nil)
	if err != nil {
		t.Fatal("unexpected Encode error:", err)
	}

	bb := b.Bytes()
	expect := []byte{
		'O', 'g', 'g', 'S',
		0,
		0,
		2, 0, 0, 0, 0, 0, 0, 0,
		1, 0, 0, 0,
		0, 0, 0, 0,
		0xda, 0xf7, 0x1c, 0xce, // crc
		1,
		0, // segment table
	}

	if !bytes.Equal(bb, expect) {
		t.Fatalf("bytes != expected:\n%x\n%x", bb, expect)
	}
}

func TestEmptyEncodeEOS(t *testing.T) {
	var b bytes.Buffer
	e := NewEncoder(1, &b)

	err := e.EncodeEOS(2, nil)
	if err != nil {
		t.Fatal("unexpected Encode error:", err)
	}

	bb := b.Bytes()
	expect := []byte{
		'O', 'g', 'g', 'S',
		0,
		EOS,
		2, 0, 0, 0, 0, 0, 0, 0,
		1, 0, 0, 0,
		0, 0, 0, 0,
		0x9a, 0x50, 0x2c, 0xd7, // crc
		1,
		0, // segment table
	}

	if !bytes.Equal(bb, expect) {
		t.Fatalf("bytes != expected:\n%x\n%x", bb, expect)
	}
}

func TestBasicEncode(t *testing.T) {
	var b bytes.Buffer
	e := NewEncoder(1, &b)

	err := e.Encode(2, [][]byte{[]byte("hello")})
	if err != nil {
		t.Fatal("unexpected Encode error:", err)
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

func TestBasicEncodeEOS(t *testing.T) {
	var b bytes.Buffer
	e := NewEncoder(1, &b)

	err := e.EncodeEOS(7, [][]byte{nil})
	if err != nil {
		t.Fatal("unexpected EncodeEOS error:", err)
	}

	bb := b.Bytes()
	expect := []byte{
		'O', 'g', 'g', 'S',
		0,
		EOS,
		7, 0, 0, 0, 0, 0, 0, 0,
		1, 0, 0, 0,
		0, 0, 0, 0,
		0x79, 0x1e, 0xe7, 0xe7, // crc
		1,
		0, // segment table
	}

	if !bytes.Equal(bb, expect) {
		t.Fatalf("bytes != expected:\n%x\n%x", bb, expect)
	}
}

func TestLongEncode(t *testing.T) {
	var b bytes.Buffer
	e := NewEncoder(1, &b)

	var junk bytes.Buffer
	for i := 0; i < maxPageSize*2; i++ {
		junk.WriteByte('x')
	}

	err := e.Encode(2, [][]byte{junk.Bytes()})
	if err != nil {
		t.Fatal("unexpected Encode error:", err)
	}

	bb := b.Bytes()
	expect := []byte{
		'O', 'g', 'g', 'S',
		0,
		0,
		2, 0, 0, 0, 0, 0, 0, 0,
		1, 0, 0, 0,
		0, 0, 0, 0,
		0xee, 0xb2, 0x0b, 0xca, // crc
		255,
	}

	if !bytes.Equal(bb[:headsz], expect) {
		t.Fatalf("bytes != expected:\n%x\n%x", bb[:headsz], expect)
	}

	expect2 := []byte{
		'O', 'g', 'g', 'S',
		0,
		COP,
		2, 0, 0, 0, 0, 0, 0, 0,
		1, 0, 0, 0,
		1, 0, 0, 0,
		0x17, 0x0d, 0xe6, 0xe6, // crc
		255,
	}

	if !bytes.Equal(bb[maxPageSize:maxPageSize+headsz], expect2) {
		t.Fatalf("bytes != expected:\n%x\n%x", bb[maxPageSize:maxPageSize+headsz], expect2)
	}
}

type limitedWriter struct {
	N int64
}

func (w *limitedWriter) Write(p []byte) (int, error) {
	if w.N <= int64(len(p)) {
		n := w.N
		w.N = 0
		return int(n), io.ErrClosedPipe
	}

	w.N -= int64(len(p))
	return len(p), nil
}

func TestShortWrites(t *testing.T) {
	e := NewEncoder(1, &limitedWriter{N: 0})
	err := e.Encode(2, [][]byte{[]byte("hello")})
	if err != io.ErrClosedPipe {
		t.Fatal("expected ErrClosedPipe, got:", err)
	}

	e = NewEncoder(1, &limitedWriter{N: maxPageSize + 1})
	var junk bytes.Buffer
	for i := 0; i < maxPageSize*2; i++ {
		junk.WriteByte('x')
	}
	err = e.Encode(2, [][]byte{junk.Bytes()})
	if err != io.ErrClosedPipe {
		t.Fatal("expected ErrClosedPipe, got:", err)
	}
}
