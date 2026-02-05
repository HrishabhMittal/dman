package main

import (
	"flag"
	"fmt"
	"github.com/HrishabhMittal/dman/pkg/download"
	"github.com/HrishabhMittal/gotorrent/pkg/torrent"
)


func main() {
	torrentFile := flag.String("file","","path to torrent file")
	downloadLink := flag.String("link","","link to download")
	numConcurrentDown := flag.Int("num",16,"number of concurrent downloads")
	outputFile := flag.String("output","","file output path (for downloads)")
	flag.Parse()

	if *torrentFile=="" {
		var dl *download.Downloader
		if *outputFile=="" {
			dl = download.NewInferDownloader(*downloadLink,*numConcurrentDown)
		} else {
			dl = download.NewDownloader(*downloadLink,*outputFile,*numConcurrentDown)
		}
		fmt.Println("Initializing...")
		if err := dl.Prepare(); err != nil {
			fmt.Printf("Initialization failed: %v\n", err)
			return
		}
		fmt.Printf("File Size: %d | Multi-threaded: %v\n", dl.TotalSize, dl.SupportsRange)
		go dl.Start()
		pm := &download.ProgressManager{TotalSize: dl.TotalSize,Current: &dl.Downloaded}
		end := make(chan struct{})
		go pm.DisplayDownloadProgress(end)
		<-end
	} else if *downloadLink=="" {
		tf, err := torrent.NewTorrentFile(*torrentFile)
		fmt.Printf("Downloading: %s\n", tf.Name)
		dn,err := torrent.NewDownloader(tf)
		if err != nil {
			fmt.Println("couldnt start download:", err)
			return
		}

		// for now
		dn.PrintLogs()


		// pm := &download.ProgressManager{TotalSize: &dl.TotalSize,Current: &dl.Downloaded}
		// end := make(chan struct{})
		// go pm.DisplayDownloadProgress(end)
		dn.Wait()
		// <-end
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
