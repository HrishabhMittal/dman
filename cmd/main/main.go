package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type pieceRange struct {
	start,length int64
};

type OffsetWrite struct {
	file *os.File
	offset int64
}
func (ofw *OffsetWrite) Write(p []byte) (n int, err error) {
	n, err = ofw.file.WriteAt(p,ofw.offset)
	ofw.offset+=int64(n)
	return n, err
}
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
	fmt.Println(resp)
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
func fetchPiece(url string,pr pieceRange,file *os.File,done chan status) {
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
	writer := &OffsetWrite{file: file,offset: pr.start}
	_, err = io.Copy(writer,resp.Body)
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
const threads int64 = 64
func main() {
	args := os.Args
	customClient = &http.Client{
	    Transport: &http.Transport{
	        MaxIdleConns:        100,
	        MaxIdleConnsPerHost: 100,
	        IdleConnTimeout:     20 * time.Second,
	    },
	}
	if len(args)<3 {
		fmt.Println("args not enough")
		return
	}
	size, err := getFileSize(args[1])
	if err != nil {
		fmt.Println("Couldn't get file size")
		return
	} else {
		fmt.Println("File size:",size)
	}
	b, err := SupportsRanges(args[1])
	if err != nil {
		fmt.Println("Couldn't get range support")
		return
	} else {
		fmt.Println("Range Support: ",b)
	}
	if !b {
		fmt.Println("download doesnt support ts")
		fmt.Println("single threaded download...")
		file, err := os.Create(args[2])
		if err != nil {
			fmt.Println("couldn't create file")
			return
		}
		err = fetch(args[1],file)
		if err != nil {
			fmt.Println("fetch error")
			return
		}
		fmt.Println("file downloaded")
	} else {
		ch := make(chan status,threads)
		var pieceLen int64 = size/int64(threads)
		file, err := os.Create(args[2])
		if err != nil {
			fmt.Println("couldn't create file")
			return
		}
		for i := range threads-1 {
			go fetchPiece(args[1],pieceRange{start: pieceLen*int64(i), length: pieceLen},file,ch)
		}
		go fetchPiece(args[1],pieceRange{start: pieceLen*int64(threads-1), length: size-int64(threads-1)*pieceLen},file,ch)
		successful := 0
		for successful<int(threads) {
			st := <-ch
			if st.successful {
				successful++
				fmt.Println("thread finished.")
			} else {
				go fetchPiece(args[1],st.pieceRange,file,ch)
			}
		}
		fmt.Println("file downloaded")
	}
}
