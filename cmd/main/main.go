package main

import (
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)
const threads int64 = 8
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
		var total_downloaded int64 = 0
		monitor := func(num int64) {
			total_downloaded+=num
		}
		for i := range threads-1 {
			go fetchPiece(args[1],pieceRange{start: pieceLen*int64(i), length: pieceLen},file,ch,monitor)
		}
		go fetchPiece(args[1],pieceRange{start: pieceLen*int64(threads-1), length: size-int64(threads-1)*pieceLen},file,ch,monitor)

		go func() {
			successful := int64(0)
			for successful<threads {
				st := <-ch
				if st.successful {
					successful++
				} else {
					go fetchPiece(args[1],st.pieceRange,file,ch,monitor)
				}
			}
		}()

		










		// current_threads := int64(0)
		// launched_threads := int64(0)
		// one_piece := int64(64*1024*1024)
		// total_threads := size/int64(one_piece)
		// successful_threads := make(chan bool,threads)
		// go func() {
		// 	successful := int64(0)
		// 	for successful<total_threads {
		// 		st := <-ch
		// 		if st.successful {
		// 			successful++
		// 			atomic.AddInt64(&current_threads,-1)
		// 			successful_threads<-true
		// 		} else {
		// 			go fetchPiece(args[1],st.pieceRange,file,ch,monitor)
		// 		}
		// 	}
		// }()
		// go func() {
		// 	for launched_threads<total_threads {
		// 		for atomic.LoadInt64(&current_threads)<threads && launched_threads<total_threads {
		// 			pieceLen := one_piece
		// 			if launched_threads+1==total_threads {
		// 				// pieceLen = min(one_piece,int(size-(total_threads-1)*(int64(one_piece))))
		// 				pieceLen += size%one_piece
		// 			}
		// 			go fetchPiece(args[1],pieceRange{start: one_piece*launched_threads, length: pieceLen},file,ch,monitor)
		// 			atomic.AddInt64(&current_threads,1);
		// 			launched_threads++
		// 		}
		// 		<-successful_threads
		// 	}
		// }()




		startTime := time.Now()
		lastDownloadedSnapshot := atomic.LoadInt64(&total_downloaded)
		lastSnapshotTime := time.Now()
		
		for {
		    currentDownloaded := atomic.LoadInt64(&total_downloaded)
		    if currentDownloaded >= size {
		        progressBar(100, 50)
		        break
		    }
		
		    now := time.Now()
		    elapsed := now.Sub(startTime)
		    snapshotDuration := now.Sub(lastSnapshotTime)
		    
		    // 1. Calculate Average Speed & ETC
		    avgSpeed := float64(currentDownloaded) / elapsed.Seconds()
		    var etc time.Duration
		    if currentDownloaded > 0 {
		        remainingBytes := size - currentDownloaded
		        etc = time.Duration(float64(remainingBytes)/avgSpeed) * time.Second
		    }
		
		    // 2. Calculate Current Instantaneous Speed
		    // Bytes divided by seconds (float64)
		    bytesSinceLast := currentDownloaded - lastDownloadedSnapshot
		    instSpeed := float64(bytesSinceLast) / snapshotDuration.Seconds()
		
		    // 3. Format Units (Bytes -> KB -> MB -> GB)
		    units := []string{"B/s", "KB/s", "MB/s", "GB/s"}
		    unitIdx := 0
		    displaySpeed := instSpeed
		    for displaySpeed >= 1024 && unitIdx < len(units)-1 {
		        displaySpeed /= 1024
		        unitIdx++
		    }
		
		    // 4. Update Snapshots
		    lastDownloadedSnapshot = currentDownloaded
		    lastSnapshotTime = now
		
		    // 5. Print Output
		    progressBar(int(100*currentDownloaded/size), 50)
		    fmt.Printf(" | Elapsed: %s | ETC: %s | Speed: %.2f %s           ",
		        elapsed.Round(time.Second),
		        etc.Round(time.Second),
		        displaySpeed,
		        units[unitIdx],
		    )
		
		    time.Sleep(500 * time.Millisecond)
		}
		// progress bar
		// startTime := time.Now()
		// last_downloaded_snapshot := total_downloaded
		// last_snapshot_time := time.Now()
		// for {
		//     currentDownloaded := total_downloaded
		//     if currentDownloaded >= size {
		//         progressBar(100, 50)
		//         break
		//     }
		//     elapsed := time.Since(startTime)
		//     percent := int(100 * currentDownloaded / size)
		//     avgSpeed := float64(currentDownloaded) / elapsed.Seconds()
		//     var etc time.Duration
		//     if currentDownloaded > 0 {
		//         remainingBytes := size - currentDownloaded
		//         etcSeconds := float64(remainingBytes) / avgSpeed
		//         etc = time.Duration(etcSeconds) * time.Second
		//     }
		//     progressBar(percent, 50)
		// 	char := " "
		// 	var downloaded float32 = float32(total_downloaded-last_downloaded_snapshot)/float32(time.Since(last_snapshot_time))
		// 	last_downloaded_snapshot = total_downloaded
		// 	last_snapshot_time = time.Now()
		// 	if downloaded>1000. {
		// 		downloaded/=1000.
		// 		if char == "M" {
		// 			char = "G";
		// 		}
		// 		if char == "K" {
		// 			char = "M";
		// 		}
		// 		if char == " " {
		// 			char = "K"
		// 		}
		// 	}
		// 	fmt.Printf(" | Elapsed: %s | ETC: %s | Speed: %f %sbps      ", 
		//         elapsed.Round(time.Second), 
		//         etc.Round(time.Second),
		// 		downloaded,
		// 		char,
		//     )
		//     time.Sleep(500 * time.Millisecond)
		// }
		// fmt.Println("")
		// fmt.Println("file downloaded")
	}
}
