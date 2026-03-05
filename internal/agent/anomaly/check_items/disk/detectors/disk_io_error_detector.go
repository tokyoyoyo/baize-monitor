package detectors

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"baize-monitor/internal/agent/anomaly/check_items/disk"
	"baize-monitor/pkg/dto/request"
)

type diskIOErrorDetector struct {
	errorThreshold int64
}

func init() {
	disk.RegisterDetector(&diskIOErrorDetector{
		errorThreshold: 100,
	})
}

func (d *diskIOErrorDetector) DetectorType() string {
	return "disk_io_error"
}

func (d *diskIOErrorDetector) Detect() (request.AnomalyResult, error) {
	diskstats, err := os.Open("/proc/diskstats")
	if err != nil {
		return request.AnomalyResult{}, err
	}
	defer diskstats.Close()

	scanner := bufio.NewScanner(diskstats)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		
		if len(fields) < 14 {
			continue
		}

		device := fields[2]
		
		ioErrors, err := strconv.ParseInt(fields[13], 10, 64)
		if err != nil {
			continue
		}

		if ioErrors > d.errorThreshold {
			return disk.CreateAnomalyResult(
				d.DetectorType(),
				"warning",
				"Disk I/O errors detected",
				map[string]interface{}{
					"device":    device,
					"io_errors": ioErrors,
					"threshold": d.errorThreshold,
					"issue":     "accumulated_io_errors",
				},
			), nil
		}
	}

	return request.AnomalyResult{}, nil
}
