package api

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode"

	"github.com/VatsalP117/iris/pkg/core"
)

const (
	maxIdentifierLength = 128
	maxURLLength        = 2048
	maxFutureClockSkew  = 5 * time.Minute
)

func (h *Handler) prepareIncomingEvent(
	ctx context.Context,
	event *core.Event,
	receivedAt time.Time,
	origin string,
) error {
	event.ID = strings.TrimSpace(event.ID)
	event.EventName = strings.TrimSpace(event.EventName)
	event.SiteID = strings.TrimSpace(event.SiteID)
	event.SessionID = strings.TrimSpace(event.SessionID)
	event.VisitorID = strings.TrimSpace(event.VisitorID)

	for field, value := range map[string]string{
		"id": event.ID, "event name": event.EventName, "site id": event.SiteID,
		"session id": event.SessionID, "visitor id": event.VisitorID,
	} {
		if value == "" {
			return fmt.Errorf("%s is required", field)
		}
		if len(value) > maxIdentifierLength {
			return fmt.Errorf("%s exceeds %d characters", field, maxIdentifierLength)
		}
	}
	if strings.HasPrefix(event.EventName, "$") &&
		event.EventName != "$pageview" && event.EventName != "$click" && event.EventName != "$web_vital" {
		return fmt.Errorf("unsupported reserved event name %q", event.EventName)
	}
	if strings.IndexFunc(event.EventName, unicode.IsControl) >= 0 {
		return fmt.Errorf("event name contains control characters")
	}
	if event.ScreenWidth < 0 || event.ScreenWidth > 100000 {
		return fmt.Errorf("screen width is out of range")
	}

	parsedURL, err := normalizeTrackedURL(event.URL)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}
	payloadDomain := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(event.Domain)), ".")
	if payloadDomain != "" && payloadDomain != parsedURL.Hostname() {
		return fmt.Errorf("domain does not match url hostname")
	}
	event.URL = parsedURL.String()
	event.Domain = parsedURL.Hostname()
	event.Pathname = parsedURL.EscapedPath()
	if event.Pathname == "" {
		event.Pathname = "/"
	}
	if origin = strings.TrimSpace(origin); origin != "" {
		parsedOrigin, parseErr := url.Parse(origin)
		if parseErr != nil || (parsedOrigin.Scheme != "http" && parsedOrigin.Scheme != "https") ||
			strings.ToLower(parsedOrigin.Hostname()) != event.Domain {
			return fmt.Errorf("%w: request origin does not match event domain", core.ErrDomainNotAllowed)
		}
	}

	if event.Referrer != "" {
		parsedReferrer, parseErr := normalizeTrackedURL(event.Referrer)
		if parseErr != nil {
			return fmt.Errorf("invalid referrer: %w", parseErr)
		}
		event.Referrer = parsedReferrer.String()
		event.ReferrerHost = strings.TrimPrefix(parsedReferrer.Hostname(), "www.")
	}
	if err := h.Repo.ValidateSite(ctx, event.SiteID, event.Domain); err != nil {
		return err
	}

	event.ReceivedAt = receivedAt.UTC()
	if event.Timestamp.IsZero() {
		event.Timestamp = event.ReceivedAt
	} else {
		event.Timestamp = event.Timestamp.UTC()
		if event.Timestamp.After(event.ReceivedAt.Add(maxFutureClockSkew)) {
			return fmt.Errorf("event timestamp is too far in the future")
		}
	}
	if event.SchemaVersion == 0 {
		event.SchemaVersion = 1
	}
	if event.SchemaVersion != 1 {
		return fmt.Errorf("unsupported schema version %d", event.SchemaVersion)
	}
	if len(event.SDKVersion) > 64 {
		return fmt.Errorf("sdk version exceeds 64 characters")
	}
	if event.Properties == nil {
		event.Properties = map[string]any{}
	} else {
		event.Properties = truncateStrings(event.Properties, 200).(map[string]any)
	}
	return nil
}

func normalizeTrackedURL(raw string) (*url.URL, error) {
	if len(raw) == 0 || len(raw) > maxURLLength {
		return nil, fmt.Errorf("url must contain between 1 and %d characters", maxURLLength)
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" || parsed.User != nil {
		return nil, fmt.Errorf("url must be an absolute http or https url")
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.RawQuery = ""
	parsed.ForceQuery = false
	parsed.Fragment = ""
	if parsed.Path == "" {
		parsed.Path = "/"
	}
	return parsed, nil
}
