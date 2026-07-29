package reliability

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
)

type FaultSuiteConfig struct {
	ServerBinary string
	OutputDir    string
	RunID        string
	Quick        bool
}

type FaultResult struct {
	Name            string `json:"name"`
	Passed          bool   `json:"passed"`
	Recovered       bool   `json:"recovered"`
	AcceptedEvents  int    `json:"accepted_events"`
	RejectedEvents  int    `json:"rejected_events"`
	RequestErrors   int    `json:"request_errors"`
	MissingAccepted int    `json:"missing_accepted"`
	DuplicateRows   int    `json:"duplicate_rows"`
	UnexpectedRows  int    `json:"unexpected_rows"`
	FieldMismatches int    `json:"field_mismatches"`
	ReportPath      string `json:"report_path,omitempty"`
	Error           string `json:"error,omitempty"`
}

type BackupResult struct {
	Passed         bool   `json:"passed"`
	BackupPath     string `json:"backup_path"`
	IntegrityCheck string `json:"integrity_check"`
	StoredRows     int    `json:"stored_rows"`
	RestoreBooted  bool   `json:"restore_booted"`
	APIChecks      int    `json:"api_checks"`
	Error          string `json:"error,omitempty"`
}

type FaultSuiteReport struct {
	RunID       string        `json:"run_id"`
	GeneratedAt time.Time     `json:"generated_at"`
	Scenarios   []FaultResult `json:"scenarios"`
	Backup      BackupResult  `json:"backup"`
	Passed      bool          `json:"passed"`
}

type FaultProxy struct {
	URL        string
	server     *http.Server
	listener   net.Listener
	target     *httputil.ReverseProxy
	requests   atomic.Int64
	outage     atomic.Bool
	failEvery  int64
	delayEvery int64
	delay      time.Duration
}

func RunFaultSuite(ctx context.Context, config FaultSuiteConfig) (*FaultSuiteReport, error) {
	normalized, err := normalizeFaultSuiteConfig(config)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(normalized.OutputDir, 0755); err != nil {
		return nil, err
	}

	server := &LabServer{
		Binary:  normalized.ServerBinary,
		WorkDir: filepath.Join(normalized.OutputDir, "server"),
		DBPath:  filepath.Join(normalized.OutputDir, "server", "iris.db"),
		LogPath: filepath.Join(normalized.OutputDir, "server.log"),
	}
	if err := server.Start(ctx); err != nil {
		return nil, err
	}
	defer server.Stop()

	report := &FaultSuiteReport{
		RunID:       normalized.RunID,
		GeneratedAt: time.Now().UTC(),
		Passed:      true,
	}

	report.Scenarios = append(report.Scenarios,
		runRestartFault(ctx, normalized, server),
		runProxyFault(ctx, normalized, server, "intermittent-503", false),
		runProxyFault(ctx, normalized, server, "network-outage", true),
		runSQLiteLockFault(ctx, normalized, server),
		runSQLiteDiskLimitFault(ctx, normalized, server),
	)
	for _, scenario := range report.Scenarios {
		if !scenario.Passed {
			report.Passed = false
		}
	}

	report.Backup = runBackupVerification(ctx, normalized, server)
	if !report.Backup.Passed {
		report.Passed = false
	}
	if err := WriteFaultSuiteReport(report, normalized.OutputDir); err != nil {
		return nil, err
	}
	return report, nil
}

func normalizeFaultSuiteConfig(config FaultSuiteConfig) (FaultSuiteConfig, error) {
	if config.ServerBinary == "" {
		config.ServerBinary = filepath.Join("dist", "iris-server")
	}
	absoluteBinary, err := filepath.Abs(config.ServerBinary)
	if err != nil {
		return FaultSuiteConfig{}, err
	}
	if _, err := os.Stat(absoluteBinary); err != nil {
		return FaultSuiteConfig{}, fmt.Errorf("server binary: %w", err)
	}
	config.ServerBinary = absoluteBinary
	if config.RunID == "" {
		config.RunID = time.Now().UTC().Format("20060102T150405Z")
	}
	if !validRunID(config.RunID) {
		return FaultSuiteConfig{}, fmt.Errorf("invalid fault suite run ID %q", config.RunID)
	}
	if config.OutputDir == "" {
		config.OutputDir = filepath.Join("artifacts", "reliability", "faults-"+config.RunID)
	}
	config.OutputDir, err = filepath.Abs(config.OutputDir)
	return config, err
}

func faultLoadConfig(config FaultSuiteConfig, server *LabServer, name string) Config {
	duration := 10 * time.Second
	if config.Quick {
		duration = 4 * time.Second
	}
	return Config{
		TargetURL:      server.URL(),
		DBPath:         server.DBPath,
		RunID:          config.RunID + "-" + name,
		SiteID:         "iris-lab-fault-" + config.RunID + "-" + name,
		Rate:           500,
		Duration:       duration,
		BatchSize:      1,
		Workers:        96,
		RequestTimeout: 2 * time.Second,
	}
}

func runRestartFault(
	ctx context.Context,
	config FaultSuiteConfig,
	server *LabServer,
) FaultResult {
	loadConfig := faultLoadConfig(config, server, "restart")
	delay := loadConfig.Duration / 3
	restartDone := make(chan error, 1)
	go func() {
		time.Sleep(delay)
		restartDone <- server.Restart(ctx)
	}()
	report, err := Run(ctx, loadConfig)
	restartErr := <-restartDone
	recovered := restartErr == nil && serverHealthy(ctx, server.URL())
	return finalizeFaultResult(config.OutputDir, "restart", report, err, recovered, restartErr)
}

func runProxyFault(
	ctx context.Context,
	config FaultSuiteConfig,
	server *LabServer,
	name string,
	outage bool,
) FaultResult {
	proxy, err := NewFaultProxy(server.URL())
	if err != nil {
		return FaultResult{Name: name, Error: err.Error()}
	}
	defer proxy.Close()

	loadConfig := faultLoadConfig(config, server, name)
	loadConfig.TargetURL = proxy.URL
	if outage {
		go func() {
			time.Sleep(loadConfig.Duration / 3)
			proxy.outage.Store(true)
			time.Sleep(loadConfig.Duration / 4)
			proxy.outage.Store(false)
		}()
	} else {
		proxy.failEvery = 7
		proxy.delayEvery = 5
		proxy.delay = 25 * time.Millisecond
	}

	report, runErr := Run(ctx, loadConfig)
	recovered := serverHealthy(ctx, server.URL())
	return finalizeFaultResult(config.OutputDir, name, report, runErr, recovered, nil)
}

func runSQLiteLockFault(
	ctx context.Context,
	config FaultSuiteConfig,
	server *LabServer,
) FaultResult {
	loadConfig := faultLoadConfig(config, server, "sqlite-lock")
	faultDone := make(chan error, 1)
	go func() {
		time.Sleep(loadConfig.Duration / 3)
		faultDone <- holdExclusiveLock(ctx, server.DBPath, loadConfig.Duration/4)
	}()
	report, err := Run(ctx, loadConfig)
	faultErr := <-faultDone
	recovered := faultErr == nil && serverHealthy(ctx, server.URL())
	return finalizeFaultResult(config.OutputDir, "sqlite-lock", report, err, recovered, faultErr)
}

func runSQLiteDiskLimitFault(
	ctx context.Context,
	config FaultSuiteConfig,
	server *LabServer,
) FaultResult {
	loadConfig := faultLoadConfig(config, server, "sqlite-disk-limit")
	if err := server.Stop(); err != nil {
		return FaultResult{Name: "sqlite-disk-limit", Error: err.Error()}
	}
	server.Env = []string{"IRIS_LAB_DB_EXTRA_PAGES=2"}
	if err := server.Start(ctx); err != nil {
		server.Env = nil
		return FaultResult{Name: "sqlite-disk-limit", Error: err.Error()}
	}
	report, err := Run(ctx, loadConfig)
	stopErr := server.Stop()
	server.Env = nil
	restartErr := server.Start(ctx)
	recoveryErr := errors.Join(stopErr, restartErr)
	recovered := recoveryErr == nil && serverHealthy(ctx, server.URL())
	result := finalizeFaultResult(config.OutputDir, "sqlite-disk-limit", report, err, recovered, recoveryErr)
	if result.Passed && result.RejectedEvents == 0 && result.RequestErrors == 0 {
		result.Passed = false
		result.Error = "disk limit did not produce a write failure"
	}
	return result
}

func finalizeFaultResult(
	outputDir, name string,
	report *Report,
	runErr error,
	recovered bool,
	faultErr error,
) FaultResult {
	result := FaultResult{Name: name, Recovered: recovered}
	if runErr != nil {
		result.Error = runErr.Error()
		return result
	}
	if faultErr != nil {
		result.Error = faultErr.Error()
	}
	if report == nil {
		if result.Error == "" {
			result.Error = "no report generated"
		}
		return result
	}

	reportDir := filepath.Join(outputDir, name)
	if _, err := WriteReport(report, reportDir); err != nil {
		result.Error = err.Error()
		return result
	}
	result.ReportPath = filepath.Join(reportDir, "report.md")
	result.AcceptedEvents = report.Load.AcceptedEvents
	result.RejectedEvents = report.Load.RejectedEvents
	result.RequestErrors = report.Load.RequestErrors
	result.MissingAccepted = report.Storage.MissingEvents
	result.DuplicateRows = report.Storage.DuplicateRows
	result.UnexpectedRows = report.Storage.UnexpectedRows
	result.FieldMismatches = report.Storage.FieldMismatches
	result.Passed = recovered &&
		faultErr == nil &&
		report.Storage.MissingEvents == 0 &&
		report.Storage.DuplicateRows == 0 &&
		report.Storage.UnexpectedRows == 0 &&
		report.Storage.FieldMismatches == 0
	return result
}

func NewFaultProxy(targetURL string) (*FaultProxy, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	proxy := &FaultProxy{
		target: httputil.NewSingleHostReverseProxy(target),
	}
	proxy.server = &http.Server{Handler: http.HandlerFunc(proxy.serveHTTP)}
	proxy.URL = "http://" + listener.Addr().String()
	proxy.listener = listener
	go proxy.server.Serve(listener)
	return proxy, nil
}

func (p *FaultProxy) serveHTTP(w http.ResponseWriter, r *http.Request) {
	sequence := p.requests.Add(1)
	if r.Method == http.MethodPost {
		if p.outage.Load() || p.failEvery > 0 && sequence%p.failEvery == 0 {
			http.Error(w, "injected transient failure", http.StatusServiceUnavailable)
			return
		}
		if p.delayEvery > 0 && sequence%p.delayEvery == 0 {
			time.Sleep(p.delay)
		}
	}
	p.target.ServeHTTP(w, r)
}

func (p *FaultProxy) Close() {
	_ = p.server.Close()
	_ = p.listener.Close()
}

func serverHealthy(ctx context.Context, baseURL string) bool {
	request, _ := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/sites", nil)
	response, err := (&http.Client{Timeout: 2 * time.Second}).Do(request)
	if err != nil {
		return false
	}
	defer response.Body.Close()
	return response.StatusCode == http.StatusOK
}

func holdExclusiveLock(ctx context.Context, dbPath string, duration time.Duration) error {
	database, err := sql.Open("sqlite3", dbPath+"?_busy_timeout=2000")
	if err != nil {
		return err
	}
	defer database.Close()
	connection, err := database.Conn(ctx)
	if err != nil {
		return err
	}
	defer connection.Close()
	if _, err := connection.ExecContext(ctx, "BEGIN EXCLUSIVE"); err != nil {
		return err
	}
	time.Sleep(duration)
	_, err = connection.ExecContext(ctx, "COMMIT")
	return err
}

func runBackupVerification(
	ctx context.Context,
	config FaultSuiteConfig,
	server *LabServer,
) BackupResult {
	loadConfig := faultLoadConfig(config, server, "backup-source")
	loadConfig.Rate = 200
	loadConfig.Duration = 2 * time.Second
	if !config.Quick {
		loadConfig.Duration = 10 * time.Second
	}
	report, err := Run(ctx, loadConfig)
	if err != nil {
		return BackupResult{Error: err.Error()}
	}
	if !report.Passed {
		return BackupResult{Error: "backup source load did not pass"}
	}

	backupPath := filepath.Join(config.OutputDir, "backup", "iris-backup.db")
	if err := os.MkdirAll(filepath.Dir(backupPath), 0755); err != nil {
		return BackupResult{Error: err.Error()}
	}
	database, err := sql.Open("sqlite3", server.DBPath+"?_busy_timeout=5000")
	if err != nil {
		return BackupResult{Error: err.Error()}
	}
	escapedPath := strings.ReplaceAll(backupPath, "'", "''")
	_, vacuumErr := database.ExecContext(ctx, "VACUUM INTO '"+escapedPath+"'")
	_ = database.Close()
	if vacuumErr != nil {
		return BackupResult{BackupPath: backupPath, Error: vacuumErr.Error()}
	}

	backupDB, err := sql.Open("sqlite3", "file:"+filepath.ToSlash(backupPath)+"?mode=ro")
	if err != nil {
		return BackupResult{BackupPath: backupPath, Error: err.Error()}
	}
	var integrity string
	err = backupDB.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&integrity)
	_ = backupDB.Close()
	if err != nil {
		return BackupResult{BackupPath: backupPath, Error: err.Error()}
	}

	manifest := BuildManifest(loadConfig.RunID, loadConfig.SiteID, report.Load.PlannedEvents)
	accepted := make(map[int]struct{}, len(manifest))
	for _, event := range manifest {
		accepted[event.Sequence] = struct{}{}
	}
	storage, err := VerifyStorage(ctx, backupPath, manifest, accepted)
	if err != nil {
		return BackupResult{BackupPath: backupPath, IntegrityCheck: integrity, Error: err.Error()}
	}

	restoreServer := &LabServer{
		Binary:  config.ServerBinary,
		WorkDir: filepath.Join(config.OutputDir, "backup", "restored-server"),
		DBPath:  backupPath,
		LogPath: filepath.Join(config.OutputDir, "backup", "restored-server.log"),
	}
	if err := restoreServer.Start(ctx); err != nil {
		return BackupResult{
			BackupPath:     backupPath,
			IntegrityCheck: integrity,
			StoredRows:     storage.StoredRows,
			Error:          "boot restored database: " + err.Error(),
		}
	}
	defer restoreServer.Stop()
	restoreConfig := loadConfig
	restoreConfig.TargetURL = restoreServer.URL()
	checks := VerifyAggregates(ctx, restoreConfig, manifest, accepted)
	for _, check := range checks {
		if !check.Passed {
			return BackupResult{
				BackupPath:     backupPath,
				IntegrityCheck: integrity,
				StoredRows:     storage.StoredRows,
				RestoreBooted:  true,
				APIChecks:      len(checks),
				Error:          fmt.Sprintf("restored API check %s failed: %s", check.Name, check.Error),
			}
		}
	}
	return BackupResult{
		Passed: integrity == "ok" &&
			storage.MissingEvents == 0 &&
			storage.DuplicateRows == 0 &&
			storage.UnexpectedRows == 0 &&
			storage.FieldMismatches == 0,
		BackupPath:     backupPath,
		IntegrityCheck: integrity,
		StoredRows:     storage.StoredRows,
		RestoreBooted:  true,
		APIChecks:      len(checks),
	}
}

func WriteFaultSuiteReport(report *FaultSuiteReport, directory string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.WriteFile(filepath.Join(directory, "fault-summary.json"), data, 0644); err != nil {
		return err
	}

	var builder strings.Builder
	fmt.Fprintf(&builder, "# Iris Fault and Recovery Report: %s\n\n", report.RunID)
	if report.Passed {
		builder.WriteString("**Verdict:** PASS\n\n")
	} else {
		builder.WriteString("**Verdict:** FAIL\n\n")
	}
	builder.WriteString("| Scenario | Result | Recovered | Accepted | Rejected | Request errors | Missing accepted |\n")
	builder.WriteString("|---|---|---|---:|---:|---:|---:|\n")
	for _, scenario := range report.Scenarios {
		status := "PASS"
		if !scenario.Passed {
			status = "FAIL"
		}
		fmt.Fprintf(
			&builder,
			"| %s | %s | %t | %d | %d | %d | %d |\n",
			scenario.Name,
			status,
			scenario.Recovered,
			scenario.AcceptedEvents,
			scenario.RejectedEvents,
			scenario.RequestErrors,
			scenario.MissingAccepted,
		)
	}
	builder.WriteString("\n## Backup and restore\n\n")
	fmt.Fprintf(&builder, "- Result: %t\n", report.Backup.Passed)
	fmt.Fprintf(&builder, "- Integrity check: `%s`\n", report.Backup.IntegrityCheck)
	fmt.Fprintf(&builder, "- Reconciled rows: %d\n", report.Backup.StoredRows)
	fmt.Fprintf(&builder, "- Restored server booted: %t\n", report.Backup.RestoreBooted)
	fmt.Fprintf(&builder, "- Restored API checks: %d\n", report.Backup.APIChecks)
	if report.Backup.Error != "" {
		fmt.Fprintf(&builder, "- Error: `%s`\n", report.Backup.Error)
	}
	return os.WriteFile(filepath.Join(directory, "fault-report.md"), []byte(builder.String()), 0644)
}
