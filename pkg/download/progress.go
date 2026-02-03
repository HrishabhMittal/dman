package download

import (
	"fmt"
	"strings"
	"time"
)

type ProgressManager struct {
	TotalSize int64
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
