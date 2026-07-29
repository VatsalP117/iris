package reliability

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var readEndpoints = []string{
	"/api/stats",
	"/api/pages",
	"/api/referrers",
	"/api/vitals",
	"/api/devices",
	"/api/timeseries",
	"/api/timeseries/visitors",
	"/api/timeseries/sessions",
}

type readResult struct {
	Endpoint string
	Status   int
	Latency  time.Duration
	Err      error
}

func executeReadLoad(ctx context.Context, config Config) ReadSummary {
	client := &http.Client{Timeout: config.RequestTimeout}
	jobs := make(chan string, config.ReadWorkers*4)
	results := make(chan readResult, config.ReadWorkers*4)

	var workers sync.WaitGroup
	for i := 0; i < config.ReadWorkers; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for endpoint := range jobs {
				result := performRead(ctx, client, config, endpoint)
				if result.Err != nil && ctx.Err() != nil {
					continue
				}
				results <- result
			}
		}()
	}

	startedAt := time.Now()
	go scheduleReads(ctx, config.ReadRate, jobs)
	go func() {
		workers.Wait()
		close(results)
	}()

	summary := ReadSummary{Endpoints: map[string]EndpointSummary{}}
	var allLatencies []time.Duration
	endpointLatencies := map[string][]time.Duration{}
	for result := range results {
		summary.AttemptedRequests++
		allLatencies = append(allLatencies, result.Latency)

		endpoint := summary.Endpoints[result.Endpoint]
		endpoint.Requests++
		if endpoint.StatusCodes == nil {
			endpoint.StatusCodes = map[int]int{}
		}

		if result.Err != nil {
			summary.FailedRequests++
			endpoint.Errors++
			appendErrorSample(&summary.ErrorSamples, result.Err.Error())
		} else {
			endpoint.StatusCodes[result.Status]++
			if result.Status == http.StatusOK {
				summary.SuccessfulRequests++
			} else {
				summary.FailedRequests++
				endpoint.Errors++
				appendErrorSample(
					&summary.ErrorSamples,
					fmt.Sprintf("%s returned HTTP %d", result.Endpoint, result.Status),
				)
			}
		}

		endpointLatencies[result.Endpoint] = append(endpointLatencies[result.Endpoint], result.Latency)
		summary.Endpoints[result.Endpoint] = endpoint
	}

	elapsed := time.Since(startedAt).Seconds()
	if elapsed > 0 {
		summary.AchievedRequestsPerSec = float64(summary.AttemptedRequests) / elapsed
	}
	summary.Latency = summarizeLatencies(allLatencies)
	for name, endpoint := range summary.Endpoints {
		endpoint.Latency = summarizeLatencies(endpointLatencies[name])
		summary.Endpoints[name] = endpoint
	}
	return summary
}

func scheduleReads(ctx context.Context, rate int, jobs chan<- string) {
	defer close(jobs)
	if rate <= 0 {
		return
	}
	interval := time.Second / time.Duration(rate)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	sequence := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			endpoint := readEndpoints[sequence%len(readEndpoints)]
			sequence++
			select {
			case jobs <- endpoint:
			case <-ctx.Done():
				return
			}
		}
	}
}

func performRead(ctx context.Context, client *http.Client, config Config, endpoint string) readResult {
	startedAt := time.Now()
	requestURL := config.TargetURL + endpoint + "?site_id=" + url.QueryEscape(config.SiteID)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return readResult{Endpoint: endpoint, Err: err}
	}
	response, err := client.Do(request)
	latency := time.Since(startedAt)
	if err != nil {
		return readResult{Endpoint: endpoint, Latency: latency, Err: err}
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 64<<10))
	return readResult{Endpoint: endpoint, Status: response.StatusCode, Latency: latency}
}
