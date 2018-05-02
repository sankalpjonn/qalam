package qalam

import (
	"bufio"
	"github.com/lestrrat-go/strftime"
	"os"
	"path/filepath"
	"time"
)

// How to make something thread safe? Research and implement

type (
	Qalam struct {
		fp *os.File
		// the location of the file
		location *strftime.Strftime

		// file name created by builder
		path string

		// time location
		tloc *time.Location

		// bufio size
		bufSize int
		// bufio writer
		bw *bufio.Writer
	}
)

var (
	DefaultBufferSize = 4096
)

func New(location string) *Qalam {
	p, err := strftime.New(location)
	if err != nil {
		panic(err)
	}

	return &Qalam{
		location: p,
		tloc:     time.Local,
		bufSize:  DefaultBufferSize,
	}
}

func (q *Qalam) Likho(b []byte) (int, error) {
	return q.Write(b)
}

/*
	SetBufferSize set's the size of the buffer
	which is kept in memory before pushing to disk.
	Defaults to 4096, the default page size on older
	SSDs, can be set accordingly
*/
func (q *Qalam) SetBufferSize(b int) {
	q.bufSize = b
}

func (q *Qalam) initBuffer(path string) (err error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	fp, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	bw := bufio.NewWriter(fp)
	bw = bufio.NewWriterSize(bw, q.bufSize)

	q.path = path
	q.fp = fp
	q.bw = bw
	return nil
}

func (q *Qalam) Write(b []byte) (int, error) {
	ct := time.Now()
	path := q.location.FormatString(ct.In(q.tloc))
	if q.path != path {
		if q.fp != nil {
			q.fp.Close()
		}

		err := q.initBuffer(path)
		if err != nil {
			return 0, err
		}
	}
	return q.write(b)
}

func (q *Qalam) bytesAvailable() int {
	return q.bw.Available()
}

func (q *Qalam) Writeln(b []byte) (int, error) {
	ct := time.Now()
	path := q.location.FormatString(ct.In(q.tloc))
	if q.path != path {
		if q.fp != nil {
			q.fp.Close()
		}
		err := q.initBuffer(path)
		if err != nil {
			return 0, err
		}
	}
	return q.writeln(b)
}

func (q *Qalam) write(b []byte) (int, error) {
	if q.bytesAvailable() < len(b) {
		q.bw.Flush()
	}
	return q.bw.Write(b)
}

func (q *Qalam) writeln(b []byte) (int, error) {
	if q.bytesAvailable() < len(b) {
		q.bw.Flush()
	}
	// Newline must always be appended
	q.bw.Write(b)
	return q.bw.Write([]byte("\n"))
}
