package main

import (
	"bufio"
	"bytes"
	"io"

	"github.com/mattn/go-colorable"
)

func tee(stdin io.Reader, stdout io.Writer) string {
	var b1 bytes.Buffer
	var b2 bytes.Buffer

	tee := io.TeeReader(stdin, &b1)
	s := bufio.NewScanner(tee)
	s.Split(bufio.ScanBytes)
	for s.Scan() {
		stdout.Write(s.Bytes())
	}

	uncolorize := colorable.NewNonColorable(&b2)
	uncolorize.Write(b1.Bytes())

	return b2.String()
}
