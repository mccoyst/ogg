// © 2014 Steve McCoy under the MIT license. See LICENSE for details.

package ogg

import (
	"bytes"
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

func TestSyncDecode(t *testing.T) {
	var b bytes.Buffer
	b.Write([]byte("junk, junk, and more junk"))

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
