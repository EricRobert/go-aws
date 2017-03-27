package main

import (
	"io"
	"log"
	"net/url"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			os.Exit(1)
		}
	}()

	log.SetFlags(0)

	if len(os.Args) != 2 {
		log.Panicf("usage: %s s3://bucket/object", os.Args[0])
	}

	u, err := url.Parse(os.Args[1])
	if err != nil {
		log.Panicf("url: %s", err)
	}

	r, w := io.Pipe()

	go func() {
		for {
			_, err := io.Copy(os.Stdout, r)
			if err != nil {
				log.Panicf("stdout: %s", err)
			}
		}
	}()

	p := proxy{
		chunks: make(map[int64][]byte), w: w,
	}

	s, err := session.NewSession()
	if err != nil {
		log.Panicf("session: %s", err)
	}

	d := s3manager.NewDownloader(s)

	_, err = d.Download(&p, &s3.GetObjectInput{
		Bucket: &u.Host,
		Key:    &u.Path,
	})

	if err != nil {
		log.Panicf("s3: %s", err)
	}
}

type proxy struct {
	offset int64
	chunks map[int64][]byte
	w      io.Writer
	mu     sync.Mutex
}

func (w *proxy) WriteAt(p []byte, off int64) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.chunks[off] = append([]byte{}, p...)

	for {
		b, ok := w.chunks[w.offset]
		if !ok {
			break
		}

		delete(w.chunks, w.offset)
		w.offset += int64(len(b))

		if _, err = w.w.Write(b); err != nil {
			return
		}
	}

	n = len(p)
	return
}
