package download

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

type ProgressManager struct {
	Current *int64
	TotalSize int64
}

func (pm *ProgressManager) DisplayDownloadProgress(end chan struct{}) {
	startTime := time.Now()
	lastSnapTime := time.Now()
	lastSnapSize := int64(0)
	for {
		current := atomic.LoadInt64(pm.Current)
		if current >= pm.TotalSize {
			pm.Print(pm.TotalSize, time.Since(startTime), 0)
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
	end<-struct{}{}
}

func (pm *ProgressManager) Print(current int64, elapsed time.Duration, instSpeed float64) {
	percent := int(100 * current / pm.TotalSize)
	avgSpeed := float64(current) / elapsed.Seconds()
	var etc time.Duration
	if avgSpeed > 0 {
		remaining := pm.TotalSize - current
		etc = time.Duration(float64(remaining)/avgSpeed) * time.Second
	}

	bar := pm.getBar(percent, 30)
	fmt.Printf("\r%s %d%% | Elapsed: %s | ETC: %s | Speed: %s | Avg: %s          ",
		bar, percent,
		elapsed.Round(time.Second),
		etc.Round(time.Second),
		pm.formatBytes(instSpeed),
		pm.formatBytes(avgSpeed),
	)
}

func (pm *ProgressManager) getBar(percent, length int) string {
	filled := length * percent / 100
	return "[" + strings.Repeat("#", filled) + strings.Repeat(".", length-filled) + "]"
}

func (pm *ProgressManager) formatBytes(b float64) string {
	units := []string{"B/s", "KB/s", "MB/s", "GB/s"}
	idx := 0
	for b >= 1024 && idx < len(units)-1 {
		b /= 1024
		idx++
	}
	return fmt.Sprintf("%.2f %s", b, units[idx])
}
