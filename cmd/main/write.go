package main

import "os"

type pieceRange struct {
	start,length int64
};

type OffsetWrite struct {
	file *os.File
	offset int64
	written func(int64)
}
func (ofw *OffsetWrite) Write(p []byte) (n int, err error) {
	n, err = ofw.file.WriteAt(p,ofw.offset)
	ofw.offset+=int64(n)
	ofw.written(int64(n))
	return n, err
}
