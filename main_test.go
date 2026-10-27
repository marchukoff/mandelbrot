package main

import (
	"testing"
)

func TestParsePair(t *testing.T) {
	_, _, ok := parsePair[int]("", ",")
	if ok {
		t.Error("empty test failed")
	}

	_, _, ok = parsePair[int]("10,", ",")
	if ok {
		t.Error("not enough data 1st test failed")
	}

	_, _, ok = parsePair[int](",10", ",")
	if ok {
		t.Error("not enough data 2nd test failed")
	}

	x, y, ok := parsePair[int]("10,20", ",")
	if !ok || x != 10 || y != 20 {
		t.Error("success int test failed")
	}

	_, _, ok = parsePair[int]("10,20xy", ",")
	if ok {
		t.Error("extra data test failed")
	}

	_, _, ok = parsePair[float64]("0.5x", "x")
	if ok {
		t.Error("not enough data float test failed")
	}

	a, b, ok := parsePair[float64]("0.5x1.5", "x")
	if !ok || a != 0.5 || b != 1.5 {
		t.Error("success float test failed")
	}
}

func TestParseComplex(t *testing.T) {
	z, ok := parseComplex("1.25,-0.0625")
	if !ok || z != complex(1.25, -0.0625) {
		t.Error("success test failed")
	}
	_, ok = parseComplex(",-0.0625x")
	if ok {
		t.Error("negative test failed")
	}
}

func TestPixelToPoint(t *testing.T) {
	z := pixelToPoint(Pair[int]{100, 100}, Pair[int]{25, 75}, complex(-1.0, 1.0), complex(1.0, -1.0))
	if z != complex(-0.5, -0.5) {
		t.Error("success test failed")
	}
}
