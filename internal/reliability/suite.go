package reliability

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type SuiteConfig struct {
	ServerBinary string
	OutputDir    string
	RunID        string
	Profiles     []string
	Quick        bool
	CapturePprof bool
}

type LoadProfile struct {
	Name        string
	Rate        int
	Duration    time.Duration
	BatchSize   int
	Workers     int
	ReadRate    int
	ReadWorkers int
	Stages      []RateStage
}

type SuiteProfileResult struct {
	Name       string  `json:"name"`
	ReportPath string  `json:"report_path,omitempty"`
	Passed     bool    `json:"passed"`
	Error      string  `json:"error,omitempty"`
	Events     int     `json:"events"`
	EventRate  float64 `json:"event_rate"`
	P95MS      float64 `json:"p95_ms"`
	PeakCPU    float64 `json:"peak_cpu_percent"`
	PeakRSS    int64   `json:"peak_rss_bytes"`
}

type SuiteReport struct {
	RunID        string               `json:"run_id"`
	GeneratedAt  time.Time            `json:"generated_at"`
	ServerBinary string               `json:"server_binary"`
	Profiles     []SuiteProfileResult `json:"profiles"`
	Passed       bool                 `json:"passed"`
}

type LabServer struct {
	Binary  string
	WorkDir string
	DBPath  string
	LogPath string
	Port    int
	Env     []string

	mu      sync.Mutex
	command *exec.Cmd
	logFile *os.File
	done    chan error
}

func RunSuite(ctx context.Context, config SuiteConfig) (*SuiteReport, error) {
	normalized, profiles, err := normalizeSuiteConfig(config)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(normalized.OutputDir, 0755); err != nil {
		return nil, err
	}

	suite := &SuiteReport{
		RunID:        normalized.RunID,
		GeneratedAt:  time.Now().UTC(),
		ServerBinary: normalized.ServerBinary,
		Passed:       true,
	}

	for _, profile := range profiles {
		profileDir := filepath.Join(normalized.OutputDir, profile.Name)
		server := &LabServer{
			Binary:  normalized.ServerBinary,
			WorkDir: filepath.Join(profileDir, "server"),
			DBPath:  filepath.Join(profileDir, "server", "iris.db"),
			LogPath: filepath.Join(profileDir, "server.log"),
		}
		if err := server.Start(ctx); err != nil {
			suite.Profiles = append(suite.Profiles, SuiteProfileResult{
				Name:  profile.Name,
				Error: "start isolated server: " + err.Error(),
			})
			suite.Passed = false
			continue
		}
		result := runSuiteProfile(ctx, normalized, server, profile)
		if stopErr := server.Stop(); stopErr != nil {
			result.Passed = false
			if result.Error == "" {
				result.Error = "stop isolated server: " + stopErr.Error()
			}
		}
		suite.Profiles = append(suite.Profiles, result)
		if !result.Passed {
			suite.Passed = false
		}
	}
	if err := WriteSuiteReport(suite, normalized.OutputDir); err != nil {
		return nil, err
	}
	return suite, nil
}

func normalizeSuiteConfig(config SuiteConfig) (SuiteConfig, []LoadProfile, error) {
	if config.ServerBinary == "" {
		config.ServerBinary = filepath.Join("dist", "iris-server")
	}
	absoluteBinary, err := filepath.Abs(config.ServerBinary)
	if err != nil {
		return SuiteConfig{}, nil, err
	}
	info, err := os.Stat(absoluteBinary)
	if err != nil {
		return SuiteConfig{}, nil, fmt.Errorf("server binary: %w", err)
	}
	if info.IsDir() {
		return SuiteConfig{}, nil, fmt.Errorf("server binary %q is a directory", absoluteBinary)
	}
	config.ServerBinary = absoluteBinary
	if config.RunID == "" {
		config.RunID = time.Now().UTC().Format("20060102T150405Z")
	}
	if !validRunID(config.RunID) {
		return SuiteConfig{}, nil, fmt.Errorf("invalid suite run ID %q", config.RunID)
	}
	if config.OutputDir == "" {
		config.OutputDir = filepath.Join("artifacts", "reliability", "suite-"+config.RunID)
	}
	config.OutputDir, err = filepath.Abs(config.OutputDir)
	if err != nil {
		return SuiteConfig{}, nil, err
	}

	allProfiles := suiteProfiles(config.Quick)
	if len(config.Profiles) == 0 {
		config.Profiles = []string{"smoke"}
	}
	selected := make([]LoadProfile, 0, len(config.Profiles))
	for _, name := range config.Profiles {
		profile, ok := allProfiles[name]
		if !ok {
			return SuiteConfig{}, nil, fmt.Errorf("unknown profile %q", name)
		}
		selected = append(selected, profile)
	}
	return config, selected, nil
}

func suiteProfiles(quick bool) map[string]LoadProfile {
	if quick {
		return map[string]LoadProfile{
			"smoke":       {Name: "smoke", Rate: 10, Duration: 2 * time.Second, BatchSize: 1, Workers: 8},
			"baseline":    {Name: "baseline", Rate: 100, Duration: 3 * time.Second, BatchSize: 1, Workers: 16},
			"target-500":  {Name: "target-500", Rate: 500, Duration: 5 * time.Second, BatchSize: 1, Workers: 64},
			"target-1000": {Name: "target-1000", Rate: 1000, Duration: 5 * time.Second, BatchSize: 10, Workers: 64},
			"mixed":       {Name: "mixed", Rate: 500, Duration: 5 * time.Second, BatchSize: 10, Workers: 64, ReadRate: 25, ReadWorkers: 8},
			"ramp": {
				Name: "ramp", BatchSize: 10, Workers: 96,
				Stages: []RateStage{
					{Rate: 100, Duration: time.Second},
					{Rate: 500, Duration: time.Second},
					{Rate: 1000, Duration: time.Second},
					{Rate: 2000, Duration: time.Second},
				},
			},
			"spike": {
				Name: "spike", BatchSize: 10, Workers: 96,
				Stages: []RateStage{
					{Rate: 100, Duration: time.Second},
					{Rate: 2000, Duration: 2 * time.Second},
					{Rate: 100, Duration: time.Second},
				},
			},
			"soak": {Name: "soak", Rate: 250, Duration: 10 * time.Second, BatchSize: 10, Workers: 32, ReadRate: 10, ReadWorkers: 4},
		}
	}

	return map[string]LoadProfile{
		"smoke":       {Name: "smoke", Rate: 10, Duration: 30 * time.Second, BatchSize: 1, Workers: 16},
		"baseline":    {Name: "baseline", Rate: 100, Duration: 2 * time.Minute, BatchSize: 1, Workers: 32},
		"target-500":  {Name: "target-500", Rate: 500, Duration: 5 * time.Minute, BatchSize: 1, Workers: 128},
		"target-1000": {Name: "target-1000", Rate: 1000, Duration: 5 * time.Minute, BatchSize: 10, Workers: 128},
		"mixed":       {Name: "mixed", Rate: 500, Duration: 5 * time.Minute, BatchSize: 10, Workers: 128, ReadRate: 25, ReadWorkers: 16},
		"ramp": {
			Name: "ramp", BatchSize: 10, Workers: 192,
			Stages: []RateStage{
				{Rate: 100, Duration: 2 * time.Minute},
				{Rate: 500, Duration: 2 * time.Minute},
				{Rate: 1000, Duration: 2 * time.Minute},
				{Rate: 2000, Duration: 2 * time.Minute},
			},
		},
		"spike": {
			Name: "spike", BatchSize: 10, Workers: 192,
			Stages: []RateStage{
				{Rate: 100, Duration: 10 * time.Second},
				{Rate: 2000, Duration: 30 * time.Second},
				{Rate: 100, Duration: 20 * time.Second},
			},
		},
		"soak": {Name: "soak", Rate: 250, Duration: 30 * time.Minute, BatchSize: 10, Workers: 64, ReadRate: 10, ReadWorkers: 8},
	}
}

func runSuiteProfile(
	ctx context.Context,
	suiteConfig SuiteConfig,
	server *LabServer,
	profile LoadProfile,
) SuiteProfileResult {
	profileDir := filepath.Join(suiteConfig.OutputDir, profile.Name)
	_ = os.MkdirAll(profileDir, 0755)

	config := Config{
		TargetURL:      server.URL(),
		DBPath:         server.DBPath,
		RunID:          suiteConfig.RunID + "-" + profile.Name,
		SiteID:         "iris-lab-" + suiteConfig.RunID + "-" + profile.Name,
		ReportDir:      profileDir,
		Rate:           profile.Rate,
		Duration:       profile.Duration,
		BatchSize:      profile.BatchSize,
		Workers:        profile.Workers,
		ReadRate:       profile.ReadRate,
		ReadWorkers:    profile.ReadWorkers,
		RequestTimeout: 10 * time.Second,
		Stages:         profile.Stages,
	}

	sampler := StartResourceSampler(
		server.PID(),
		server.DBPath,
		filepath.Join(profileDir, "resources.csv"),
		500*time.Millisecond,
	)

	var profileDone chan error
	if suiteConfig.CapturePprof && profileRuntime(profile) >= 3*time.Second {
		profileDone = make(chan error, 1)
		seconds := min(5, max(1, int(profileRuntime(profile).Seconds())-1))
		go func() {
			profileDone <- captureURL(
				ctx,
				server.URL()+"/debug/pprof/profile?seconds="+strconv.Itoa(seconds),
				filepath.Join(profileDir, "cpu.pprof"),
				time.Duration(seconds+5)*time.Second,
			)
		}()
	}

	report, err := Run(ctx, config)
	resources := sampler.Stop()
	if profileDone != nil {
		_ = <-profileDone
		_ = captureURL(ctx, server.URL()+"/debug/pprof/heap", filepath.Join(profileDir, "heap.pprof"), 10*time.Second)
		_ = captureURL(ctx, server.URL()+"/debug/pprof/goroutine?debug=1", filepath.Join(profileDir, "goroutines.txt"), 10*time.Second)
	}

	result := SuiteProfileResult{Name: profile.Name}
	if err != nil {
		result.Error = err.Error()
		return result
	}
	report.Resources = resources
	if _, err := WriteReport(report, profileDir); err != nil {
		result.Error = err.Error()
		return result
	}

	result.ReportPath = filepath.Join(profileDir, "report.md")
	result.Passed = report.Passed
	result.Events = report.Load.AcceptedEvents
	result.EventRate = report.Load.AchievedEventsPerSec
	result.P95MS = report.Load.Latency.P95MS
	result.PeakCPU = report.Resources.PeakCPUPercent
	result.PeakRSS = report.Resources.PeakRSSBytes
	return result
}

func profileRuntime(profile LoadProfile) time.Duration {
	if len(profile.Stages) > 0 {
		return stagesDuration(profile.Stages)
	}
	return profile.Duration
}

func (s *LabServer) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.command != nil {
		return errors.New("lab server is already running")
	}
	if err := os.MkdirAll(s.WorkDir, 0755); err != nil {
		return err
	}
	if s.Port == 0 {
		port, err := availablePort()
		if err != nil {
			return err
		}
		s.Port = port
	}
	if s.logFile == nil {
		logFile, err := os.OpenFile(s.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		s.logFile = logFile
	}

	command := exec.Command(s.Binary)
	command.Dir = s.WorkDir
	command.Env = append(
		os.Environ(),
		"PORT="+strconv.Itoa(s.Port),
		"DB_PATH="+s.DBPath,
		"DASHBOARD_DIR="+filepath.Join(s.WorkDir, "dashboard"),
		"IRIS_LAB_PPROF=1",
	)
	command.Env = append(command.Env, s.Env...)
	command.Stdout = s.logFile
	command.Stderr = s.logFile
	if err := command.Start(); err != nil {
		return err
	}
	s.command = command
	s.done = make(chan error, 1)
	go func() { s.done <- command.Wait() }()

	baseURL := s.URL()
	s.mu.Unlock()
	err := waitForServer(ctx, baseURL, 10*time.Second)
	s.mu.Lock()
	if err != nil {
		_ = command.Process.Kill()
		s.command = nil
		return err
	}
	return nil
}

func (s *LabServer) Stop() error {
	s.mu.Lock()
	if s.command == nil {
		s.mu.Unlock()
		return nil
	}
	command := s.command
	done := s.done
	s.command = nil
	s.mu.Unlock()

	_ = command.Process.Signal(os.Interrupt)
	select {
	case err := <-done:
		if s.logFile != nil {
			_ = s.logFile.Sync()
		}
		return normalizeProcessExit(err)
	case <-time.After(3 * time.Second):
		_ = command.Process.Kill()
		<-done
		return nil
	}
}

func (s *LabServer) Restart(ctx context.Context) error {
	if err := s.Stop(); err != nil {
		return err
	}
	return s.Start(ctx)
}

func (s *LabServer) PID() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.command == nil || s.command.Process == nil {
		return 0
	}
	return s.command.Process.Pid
}

func (s *LabServer) URL() string {
	return "http://127.0.0.1:" + strconv.Itoa(s.Port)
}

func availablePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func waitForServer(ctx context.Context, baseURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: time.Second}
	for time.Now().Before(deadline) {
		request, _ := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/sites", nil)
		response, err := client.Do(request)
		if err == nil {
			_, _ = io.Copy(io.Discard, response.Body)
			_ = response.Body.Close()
			if response.StatusCode == http.StatusOK {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(50 * time.Millisecond):
		}
	}
	return fmt.Errorf("server did not become ready at %s", baseURL)
}

func normalizeProcessExit(err error) error {
	if err == nil {
		return nil
	}
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		if status, ok := exitError.Sys().(syscall.WaitStatus); ok && status.Signaled() {
			return nil
		}
	}
	return err
}

func captureURL(ctx context.Context, sourceURL, destination string, timeout time.Duration) error {
	requestContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	request, err := http.NewRequestWithContext(requestContext, http.MethodGet, sourceURL, nil)
	if err != nil {
		return err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP %d", sourceURL, response.StatusCode)
	}
	output, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer output.Close()
	_, err = io.Copy(output, response.Body)
	return err
}

func WriteSuiteReport(report *SuiteReport, directory string) error {
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	jsonData = append(jsonData, '\n')
	if err := os.WriteFile(filepath.Join(directory, "suite-summary.json"), jsonData, 0644); err != nil {
		return err
	}

	var builder strings.Builder
	fmt.Fprintf(&builder, "# Iris Reliability Suite: %s\n\n", report.RunID)
	if report.Passed {
		builder.WriteString("**Verdict:** PASS\n\n")
	} else {
		builder.WriteString("**Verdict:** FAIL\n\n")
	}
	builder.WriteString("| Profile | Result | Events | Events/s | p95 | Peak CPU | Peak RSS |\n")
	builder.WriteString("|---|---|---:|---:|---:|---:|---:|\n")
	for _, profile := range report.Profiles {
		status := "PASS"
		if !profile.Passed {
			status = "FAIL"
		}
		fmt.Fprintf(
			&builder,
			"| %s | %s | %d | %.2f | %.2f ms | %.2f%% | %d bytes |\n",
			profile.Name,
			status,
			profile.Events,
			profile.EventRate,
			profile.P95MS,
			profile.PeakCPU,
			profile.PeakRSS,
		)
		if profile.Error != "" {
			fmt.Fprintf(&builder, "\n%s error: `%s`\n\n", profile.Name, profile.Error)
		}
	}
	return os.WriteFile(filepath.Join(directory, "suite-report.md"), []byte(builder.String()), 0644)
}
