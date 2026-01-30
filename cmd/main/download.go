package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)


type status struct {
	pieceRange
	successful bool
}
func getFileSize(url string) (int64, error) {
    resp, err := http.Head(url)
    if err != nil {
        return 0, err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return 0, fmt.Errorf("server returned status: %s", resp.Status)
    }
	// fmt.Println(resp)
    return resp.ContentLength, nil
}
func SupportsRanges(url string) (bool, error) {
    resp, err := http.Head(url)
    if err != nil {
        return false, err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return false, fmt.Errorf("server returned status: %s", resp.Status)
    }
	if resp.Header.Get("Accept-Ranges") == "bytes" {
		return true, nil
	}
	return false, nil
}
func fetch(url string,file *os.File) (err error) {
	resp, err := http.Get(url)
	if err != nil {	
		return err
	}
	defer resp.Body.Close()
	writer := io.Writer(file)
	buffer := make([]byte, 1024*1024*4)
	_, err = io.CopyBuffer(writer,resp.Body,buffer)
	if err!=nil {
		return err
	}
	return nil
}

var customClient *http.Client
func fetchPiece(url string,pr pieceRange,file *os.File,done chan status,down func(int64)) {
	req, err := http.NewRequest("GET",url,nil)
	if err != nil {
		done <- status{
			pieceRange: pr,
			successful: false,
		}
		return
	}
	req.Header.Add("Range",fmt.Sprintf("bytes=%d-%d",pr.start,pr.length+pr.start-1))
	resp, err := customClient.Do(req)
	if err != nil {	
		done <- status{
			pieceRange: pr,
			successful: false,
		}
		return
	}
	if resp.StatusCode != http.StatusPartialContent {
		done <- status{
			pieceRange: pr,
			successful: false,
		}
		return
	}
	defer resp.Body.Close()
	writer := &OffsetWrite{file: file,offset: pr.start, written: down}
	buffer := make([]byte, 1024*1024*4)
	_, err = io.CopyBuffer(writer,resp.Body,buffer)
	if err!=nil {	
		done <- status{
			pieceRange: pr,
			successful: false,
		}
		return
	}
	done <- status{
		pieceRange: pr,
		successful: true,
	}
}
