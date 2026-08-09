package reliability

import (
	"fmt"

	"github.com/VatsalP117/iris/pkg/core"
)

var (
	labPages       = []string{"/", "/pricing", "/docs", "/blog/reliability"}
	labWidths      = []int{390, 820, 1440}
	labReferrers   = []string{"", "https://www.google.com/search?q=iris", "https://news.ycombinator.com/item?id=1"}
	labVitalNames  = []string{"LCP", "INP", "CLS"}
	labVitalValues = []float64{1800, 180, 0.08}
)

func BuildManifest(runID, siteID string, count int) []PlannedEvent {
	manifest := make([]PlannedEvent, count)
	for sequence := 0; sequence < count; sequence++ {
		manifest[sequence] = buildEvent(runID, siteID, sequence)
	}
	return manifest
}

func buildEvent(runID, siteID string, sequence int) PlannedEvent {
	kind := sequence % 10
	eventName := "$pageview"
	flow := "pageview"

	switch {
	case kind < 6:
	case kind < 8:
		eventName = "$click"
		flow = "click"
	case kind == 8:
		eventName = "signup"
		flow = "custom"
	default:
		eventName = "$web_vital"
		flow = "web-vital"
	}

	page := labPages[sequence%len(labPages)]
	properties := map[string]any{
		"$test_run":  runID,
		"$test_seq":  sequence,
		"$test_flow": flow,
	}

	if eventName == "$click" {
		properties["$tag"] = "button"
		properties["$text"] = "Synthetic action"
	}
	if eventName == "signup" {
		properties["plan"] = []string{"free", "pro"}[sequence%2]
	}
	if eventName == "$web_vital" {
		vitalIndex := (sequence / 10) % len(labVitalNames)
		properties["$name"] = labVitalNames[vitalIndex]
		properties["$val"] = labVitalValues[vitalIndex]
		properties["$rating"] = "good"
	}

	return PlannedEvent{
		Sequence: sequence,
		Event: core.Event{
			ID:            fmt.Sprintf("%s-%06d", runID, sequence),
			EventName:     eventName,
			URL:           "https://" + defaultDomain + page,
			Domain:        defaultDomain,
			Referrer:      labReferrers[sequence%len(labReferrers)],
			ScreenWidth:   labWidths[sequence%len(labWidths)],
			SiteID:        siteID,
			SessionID:     fmt.Sprintf("session-%06d", sequence/4),
			VisitorID:     fmt.Sprintf("visitor-%06d", sequence/7),
			Properties:    properties,
			SchemaVersion: 1,
			SDKVersion:    "iris-lab",
		},
	}
}
