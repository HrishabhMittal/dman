package main

import (
	"fmt"
	"sync/atomic"
	"flag"
	"time"
	"github.com/HrishabhMittal/dman/pkg/download"
	"github.com/HrishabhMittal/gotorrent/pkg/torrent"
)

func main() {
	torrentFile := flag.String("file","","path to torrent file")
	downloadLink := flag.String("link","","link to download")
	outputFile := flag.String("output","","file output path")
	flag.Parse()

	if *torrentFile=="" {
		var dl *download.Downloader
		if *outputFile=="" {
			dl = download.NewInferDownloader(*downloadLink,8)
		} else {
			dl = download.NewDownloader(*downloadLink,*outputFile,8)
		}
		fmt.Println("Initializing...")
		if err := dl.Prepare(); err != nil {
			fmt.Printf("Initialization failed: %v\n", err)
			return
		}
		fmt.Printf("File Size: %d | Multi-threaded: %v\n", dl.TotalSize, dl.SupportsRange)
		go dl.Start()
		pm := &download.ProgressManager{TotalSize: dl.TotalSize}
		startTime := time.Now()
		lastSnapTime := time.Now()
		lastSnapSize := int64(0)
		for {
			current := atomic.LoadInt64(&dl.Downloaded)
			if current >= dl.TotalSize {
				pm.Print(dl.TotalSize, time.Since(startTime), 0)
				fmt.Println("\nDownload Complete!")
				break
			}
			now := time.Now()
			duration := now.Sub(lastSnapTime)
			instSpeed := float64(current-lastSnapSize) / duration.Seconds()
			pm.Print(current, time.Since(startTime), instSpeed)
			lastSnapTime = now
			lastSnapSize = current
			time.Sleep(500 * time.Millisecond)
		}
	} else if *downloadLink=="" {
		tf, err := torrent.NewTorrentFile(*torrentFile)
		fmt.Printf("Downloading: %s\n", tf.Name)

		dn,err := torrent.NewDownloader(tf)
		if err != nil {
			fmt.Println("couldnt start download:", err)
			return
		}

		dn.Wait()
		fmt.Println("Exiting...")
		err = torrent.Verify(tf)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("Files verified successfully.")
		}
	} else {
		fmt.Print("no download speicified. exiting...")		
	}
}
