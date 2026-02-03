package download

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"sync/atomic"
)

type Downloader struct {
	URL           string
	FilePath      string
	Concurrency   int
	TotalSize     int64
	Downloaded    int64
	Client        *http.Client
	File          *os.File
	SupportsRange bool
}

type Piece struct {
	Start  int64
	Length int64
}

type Status struct {
	Piece      Piece
	Successful bool
}

func NewDownloader(url, filePath string, concurrency int) *Downloader {
	return &Downloader{
		URL:         url,
		FilePath:    filePath,
		Concurrency: concurrency,
		Client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
			},
		},
	}
}

func NewInferDownloader(url string,concurrency int) *Downloader {
	return &Downloader{
		URL:         url,
		FilePath:    "",
		Concurrency: concurrency,
		Client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
			},
		},
	}
}

func (d *Downloader) getFilename(resp *http.Response) string {
	cd := resp.Header.Get("Content-Disposition")
	if cd != "" {
		_, params, err := mime.ParseMediaType(cd)
		if err == nil {
			if filename, ok := params["filename"]; ok {
				return filename
			}
		}
	}
	filename := path.Base(resp.Request.URL.Path)
	if filename == "." || filename == "/" {
		return "downloaded_file"
	}
	return filename
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func (d *Downloader) Prepare() error {
	resp, err := d.Client.Head(d.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %s", resp.Status)
	}

	d.TotalSize = resp.ContentLength
	d.SupportsRange = resp.Header.Get("Accept-Ranges") == "bytes"
	
	if d.FilePath == "" || isDirectory(d.FilePath) {
		extractedName := d.getFilename(resp)
		if isDirectory(d.FilePath) {
			d.FilePath = path.Join(d.FilePath, extractedName)
		} else {
			d.FilePath = extractedName
		}
	}
	
	f, err := os.Create(d.FilePath)
	if err != nil {
		return err
	}
	d.File = f
	return nil
}

func (d *Downloader) Start() error {
	if !d.SupportsRange {
		return d.fetchSingleThreaded()
	}

	ch := make(chan Status, d.Concurrency)
	pieceLen := d.TotalSize / int64(d.Concurrency)

	go func() {
		successCount := 0
		for successCount < d.Concurrency {
			st := <-ch
			if st.Successful {
				successCount++
			} else {

				go d.fetchPiece(st.Piece, ch)
			}
		}
	}()

	for i := 0; i < d.Concurrency; i++ {
		start := pieceLen * int64(i)
		length := pieceLen
		if i == d.Concurrency-1 {
			length = d.TotalSize - start
		}
		go d.fetchPiece(Piece{Start: start, Length: length}, ch)
	}

	return nil
}

func (d *Downloader) fetchSingleThreaded() error {
	resp, err := d.Client.Get(d.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	writer := &OffsetWriter{
		file:    d.File,
		offset:  0,
		onWrite: func(n int64) { atomic.AddInt64(&d.Downloaded, n) },
	}
	_, err = io.Copy(writer, resp.Body)
	return err
}

func (d *Downloader) fetchPiece(p Piece, ch chan Status) {
	req, err := http.NewRequest("GET", d.URL, nil)
	if err != nil {
		ch <- Status{p, false}
		return
	}

	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", p.Start, p.Start+p.Length-1))
	resp, err := d.Client.Do(req)
	if err != nil || resp.StatusCode != http.StatusPartialContent {
		ch <- Status{p, false}
		return
	}
	defer resp.Body.Close()

	writer := &OffsetWriter{
		file:    d.File,
		offset:  p.Start,
		onWrite: func(n int64) { atomic.AddInt64(&d.Downloaded, n) },
	}

	buffer := make([]byte, 1024*1024)
	_, err = io.CopyBuffer(writer, resp.Body, buffer)
	ch <- Status{p, err == nil}
}
