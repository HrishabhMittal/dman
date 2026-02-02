package dman

import (
	"os"
)

type OffsetWriter struct {
	file    *os.File
	offset  int64
	onWrite func(int64)
}

func (ow *OffsetWriter) Write(p []byte) (n int, err error) {
	n, err = ow.file.WriteAt(p, ow.offset)
	if n > 0 {
		ow.offset += int64(n)
		ow.onWrite(int64(n))
	}
	return n, err
}
