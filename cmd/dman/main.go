package main

import (
	"fmt"
	"os"
	"sync/atomic"
	"time"
	"github.com/HrishabhMittal/dman/pkg"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: <url> <output_path>")
		return
	}
	dl := dman.NewDownloader(os.Args[1], os.Args[2], 8)
	fmt.Println("Initializing...")
	if err := dl.Prepare(); err != nil {
		fmt.Printf("Initialization failed: %v\n", err)
		return
	}
	fmt.Printf("File Size: %d | Multi-threaded: %v\n", dl.TotalSize, dl.SupportsRange)
	go dl.Start()
	pm := &dman.ProgressManager{TotalSize: dl.TotalSize}
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
}
