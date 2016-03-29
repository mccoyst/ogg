// © 2014 Steve McCoy under the MIT license. See LICENSE for details.

package ogg

import (
	"bytes"
	"io"
	"math/rand"
	"strings"
	"testing"
)

func TestBasicDecode(t *testing.T) {
	var b bytes.Buffer
	e := NewEncoder(1, &b)

	err := e.EncodeBOS(2, []byte("hello"))
	if err != nil {
		t.Fatal("unexpected EncodeBOS error:", err)
	}

	d := NewDecoder(&b)

	p, err := d.Decode()
	if err != nil {
		t.Fatal("unexpected Decode error:", err)
	}

	if p.Type != BOS {
		t.Fatal("expected BOS, got", p.Type)
	}

	if p.Serial != 1 {
		t.Fatal("expected serial 1, got", p.Serial)
	}

	if p.Granule != 2 {
		t.Fatal("expected granule 2, got", p.Granule)
	}

	expect := []byte{
		'h', 'e', 'l', 'l', 'o',
	}

	if !bytes.Equal(p.Packet, expect) {
		t.Fatalf("bytes != expected:\n%x\n%x", p.Packet, expect)
	}
}

func TestBadCrc(t *testing.T) {
	var b bytes.Buffer
	e := NewEncoder(1, &b)

	err := e.EncodeBOS(2, []byte("hello"))
	if err != nil {
		t.Fatal("unexpected EncodeBOS error:", err)
	}

	b.Bytes()[22] = 0

	d := NewDecoder(&b)

	_, err = d.Decode()
	if err == nil {
		t.Fatal("unexpected lack of Decode error")
	}
	if bs, ok := err.(ErrBadCrc); !ok {
		t.Fatal("exected ErrBadCrc, got:", err)
	} else if !strings.HasPrefix(bs.Error(), "invalid crc in packet") {
		t.Fatalf("the error message looks wrong: %q", err.Error())
	}
}

func TestShortDecode(t *testing.T) {
	var b bytes.Buffer
	d := NewDecoder(&b)
	_, err := d.Decode()
	if err != io.EOF {
		t.Fatal("expected EOF, got:", err)
	}

	e := NewEncoder(1, &b)
	err = e.Encode(2, []byte("hello"))
	if err != nil {
		t.Fatal("unexpected Encode error:", err)
	}
	d = NewDecoder(&io.LimitedReader{R: &b, N: headsz})
	_, err = d.Decode()
	if err != io.EOF {
		t.Fatal("expected EOF, got:", err)
	}

	b.Reset()
	e = NewEncoder(1, &b)
	err = e.Encode(2, []byte("hello"))
	if err != nil {
		t.Fatal("unexpected Encode error:", err)
	}
	d = NewDecoder(&io.LimitedReader{R: &b, N: int64(b.Len()) - 1})
	_, err = d.Decode()
	if err != io.ErrUnexpectedEOF {
		t.Fatal("expected ErrUnexpectedEOF, got:", err)
	}
}

func TestBadSegs(t *testing.T) {
	var b bytes.Buffer
	e := NewEncoder(1, &b)

	err := e.EncodeBOS(2, []byte("hello"))
	if err != nil {
		t.Fatal("unexpected EncodeBOS error:", err)
	}

	b.Bytes()[26] = 0

	d := NewDecoder(&b)
	_, err = d.Decode()
	if err != ErrBadSegs {
		t.Fatal("expected ErrBadSegs, got:", err)
	}
}

func TestSyncDecode(t *testing.T) {
	var b bytes.Buffer
	for i := 0; i < headsz-1; i++ {
		b.Write([]byte("x"))
	}
	b.Write([]byte("O"))

	for i := 0; i < headsz-3; i++ {
		b.Write([]byte("x"))
	}
	b.Write([]byte("Og"))

	for i := 0; i < headsz-5; i++ {
		b.Write([]byte("x"))
	}
	b.Write([]byte("Ogg"))

	e := NewEncoder(1, &b)

	err := e.EncodeBOS(2, []byte("hello"))
	if err != nil {
		t.Fatal("unexpected EncodeBOS error:", err)
	}

	d := NewDecoder(&b)

	p, err := d.Decode()
	if err != nil {
		t.Fatal("unexpected Decode error:", err)
	}

	if p.Type != BOS {
		t.Fatal("expected BOS, got", p.Type)
	}

	if p.Serial != 1 {
		t.Fatal("expected serial 1, got", p.Serial)
	}

	if p.Granule != 2 {
		t.Fatal("expected granule 2, got", p.Granule)
	}

	expect := []byte{
		'h', 'e', 'l', 'l', 'o',
	}

	if !bytes.Equal(p.Packet, expect) {
		t.Fatalf("bytes != expected:\n%x\n%x", p.Packet, expect)
	}
}

func TestLongDecode(t *testing.T) {
	var b bytes.Buffer
	e := NewEncoder(1, &b)

	var junk bytes.Buffer
	for i := 0; i < maxPageSize*2; i++ {
		c := byte(rand.Intn(26)) + 'a'
		junk.WriteByte(c)
	}

	err := e.Encode(2, junk.Bytes())
	if err != nil {
		t.Fatal("unexpected Encode error:", err)
	}

	d := NewDecoder(&b)
	p1, err := d.Decode()
	if err != nil {
		t.Fatal("unexpected Decode error:", err)
	}
	if p1.Type != 0 {
		t.Fatal("unexpected page type:", p1.Type)
	}
	if !bytes.Equal(p1.Packet, junk.Bytes()[:mps]) {
		t.Fatal("packet is wrong:\n\t%x\nvs\n\t%x\n", p1.Packet,  junk.Bytes()[:mps])
	}

	p2, err := d.Decode()
	if err != nil {
		t.Fatal("unexpected Decode error:", err)
	}
	if p2.Type != COP {
		t.Fatal("unexpected page type:", p1.Type)
	}
	if !bytes.Equal(p2.Packet,  junk.Bytes()[mps:mps+mps]) {
		t.Fatal("packet is wrong:\n\t%x\nvs\n\t%x\n", p2.Packet,  junk.Bytes()[mps:mps+mps])
	}

	p3, err := d.Decode()
	if err != nil {
		t.Fatal("unexpected Decode error:", err)
	}
	if p3.Type != COP {
		t.Fatal("unexpected page type:", p1.Type)
	}
	rem := (maxPageSize*2) - mps*2
	if !bytes.Equal(p3.Packet, junk.Bytes()[mps*2:mps*2+rem]) {
		t.Fatal("packet is wrong:\n\t%x\nvs\n\t%x\n", p3.Packet,  junk.Bytes()[mps*2:mps*2+rem])
	}
}
