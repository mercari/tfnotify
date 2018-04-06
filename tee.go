package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"github.com/mattn/go-colorable"
)

func tee(stdin io.Reader, stdout io.Writer) string {
	var b1 bytes.Buffer
	var b2 bytes.Buffer

	tee := io.TeeReader(stdin, &b1)
	s := bufio.NewScanner(tee)
	for s.Scan() {
		fmt.Fprintln(stdout, s.Text())
	}

	uncolorize := colorable.NewNonColorable(&b2)
	uncolorize.Write(b1.Bytes())

	return b2.String()
}
