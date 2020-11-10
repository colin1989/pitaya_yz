package compression

import (
	"bytes"
	"compress/zlib"
	"io"
	"io/ioutil"
	"sync"
)

var (
	zlibWriterPool sync.Pool
	zlibReaderPool sync.Pool
)

func DeflateData(data []byte) ([]byte, error) {
	var (
		err error
		bb  bytes.Buffer
		w   *zlib.Writer
	)
	zw := zlibWriterPool.Get()
	if zw == nil {
		zw = zlib.NewWriter(&bb)
		w = zw.(*zlib.Writer)
	} else {
		w = zw.(*zlib.Writer)
		w.Reset(&bb)
	}
	defer zlibWriterPool.Put(zw)
	_, err = w.Write(data)
	if err != nil {
		return nil, err
	}
	w.Close()
	return bb.Bytes(), nil
}

func InflateData(data []byte) ([]byte, error) {
	var err error
	zr := zlibReaderPool.Get()
	if zr != nil {
		err = zr.(zlib.Resetter).Reset(bytes.NewBuffer(data), nil)
	} else {
		zr, err = zlib.NewReader(bytes.NewBuffer(data))
	}
	if zr != nil {
		defer zlibReaderPool.Put(zr)
	}
	if err != nil {
		return nil, err
	}
	rc := zr.(io.ReadCloser)
	defer rc.Close()

	return ioutil.ReadAll(rc)
}

func IsCompressed(data []byte) bool {
	return len(data) > 2 &&
	(
		// zlib
		(data[0] == 0x78 &&
		(data[1] == 0x9C ||
		data[1] == 0x01 ||
		data[1] == 0xDA ||
		data[1] == 0x5E)) ||
		// gzip
		(data[0] == 0x1F && data[1] == 0x8B))
}
