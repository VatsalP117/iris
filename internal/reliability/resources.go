package reliability

import (
	"context"
	"encoding/csv"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ResourceSampler struct {
	pid      int
	dbPath   string
	interval time.Duration
	csvPath  string

	cancel context.CancelFunc
	done   chan struct{}
	mu     sync.Mutex
	result ResourceSummary
	cpuSum float64
	rssSum int64
}

func StartResourceSampler(pid int, dbPath, csvPath string, interval time.Duration) *ResourceSampler {
	if interval <= 0 {
		interval = time.Second
	}
	sampler := &ResourceSampler{
		pid:      pid,
		dbPath:   dbPath,
		interval: interval,
		csvPath:  csvPath,
		done:     make(chan struct{}),
	}
	sampler.result.DatabaseStartBytes = fileSize(dbPath)
	ctx, cancel := context.WithCancel(context.Background())
	sampler.cancel = cancel
	go sampler.run(ctx)
	return sampler
}

func (s *ResourceSampler) Stop() ResourceSummary {
	s.cancel()
	<-s.done

	s.mu.Lock()
	defer s.mu.Unlock()
	s.result.DatabaseEndBytes = fileSize(s.dbPath)
	s.result.DatabaseGrowthBytes = s.result.DatabaseEndBytes - s.result.DatabaseStartBytes
	if s.result.Samples > 0 {
		s.result.AverageCPUPercent = s.cpuSum / float64(s.result.Samples)
		s.result.AverageRSSBytes = s.rssSum / int64(s.result.Samples)
	}
	readBytes, writeBytes := processIO(s.pid)
	s.result.ProcessReadBytes = readBytes
	s.result.ProcessWriteBytes = writeBytes
	return s.result
}

func (s *ResourceSampler) run(ctx context.Context) {
	defer close(s.done)

	var output *os.File
	var writer *csv.Writer
	if s.csvPath != "" {
		if err := os.MkdirAll(filepath.Dir(s.csvPath), 0755); err == nil {
			if file, err := os.Create(s.csvPath); err == nil {
				output = file
				writer = csv.NewWriter(file)
				_ = writer.Write([]string{
					"timestamp",
					"cpu_percent",
					"rss_bytes",
					"database_bytes",
					"wal_bytes",
				})
			}
		}
	}
	if output != nil {
		defer output.Close()
		defer writer.Flush()
	}

	s.sample(writer)
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.sample(writer)
		}
	}
}

func (s *ResourceSampler) sample(writer *csv.Writer) {
	cpu, rss, ok := processStats(s.pid)
	if !ok {
		return
	}
	databaseBytes := fileSize(s.dbPath)
	walBytes := fileSize(s.dbPath + "-wal")

	s.mu.Lock()
	s.result.Samples++
	s.cpuSum += cpu
	s.rssSum += rss
	if cpu > s.result.PeakCPUPercent {
		s.result.PeakCPUPercent = cpu
	}
	if rss > s.result.PeakRSSBytes {
		s.result.PeakRSSBytes = rss
	}
	if walBytes > s.result.PeakWALBytes {
		s.result.PeakWALBytes = walBytes
	}
	s.mu.Unlock()

	if writer != nil {
		_ = writer.Write([]string{
			time.Now().UTC().Format(time.RFC3339Nano),
			strconv.FormatFloat(cpu, 'f', 2, 64),
			strconv.FormatInt(rss, 10),
			strconv.FormatInt(databaseBytes, 10),
			strconv.FormatInt(walBytes, 10),
		})
		writer.Flush()
	}
}

func processStats(pid int) (float64, int64, bool) {
	output, err := exec.Command(
		"ps",
		"-p",
		strconv.Itoa(pid),
		"-o",
		"%cpu=",
		"-o",
		"rss=",
	).Output()
	if err != nil {
		return 0, 0, false
	}
	fields := strings.Fields(string(output))
	if len(fields) < 2 {
		return 0, 0, false
	}
	cpu, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, 0, false
	}
	rssKB, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return 0, 0, false
	}
	return cpu, rssKB * 1024, true
}

func processIO(pid int) (int64, int64) {
	data, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "io"))
	if err != nil {
		return 0, 0
	}
	var readBytes, writeBytes int64
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		value, _ := strconv.ParseInt(fields[1], 10, 64)
		switch fields[0] {
		case "read_bytes:":
			readBytes = value
		case "write_bytes:":
			writeBytes = value
		}
	}
	return readBytes, writeBytes
}

func fileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}
