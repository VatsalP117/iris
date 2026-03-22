package api

import (
	"net/http"
	"net/url"
	"strings"
)

type CORSOptions struct {
	AllowedOrigins []string
	AllowedMethods []string
}

func NewCORSMiddleware(opts CORSOptions) func(http.HandlerFunc) http.HandlerFunc {
	allowedOrigins, allowAny := buildAllowedOrigins(opts.AllowedOrigins)
	allowedMethods := strings.Join(withDefaultMethods(opts.AllowedMethods), ", ")

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			origin := strings.TrimSpace(r.Header.Get("Origin"))
			if origin != "" {
				normalizedOrigin, ok := normalizeOrigin(origin)
				if !ok || (!allowAny && !originAllowed(allowedOrigins, normalizedOrigin)) {
					http.Error(w, "origin not allowed", http.StatusForbidden)
					return
				}

				if allowAny {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else {
					w.Header().Set("Access-Control-Allow-Origin", normalizedOrigin)
					w.Header().Add("Vary", "Origin")
				}
				w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next(w, r)
		}
	}
}

func buildAllowedOrigins(origins []string) (map[string]struct{}, bool) {
	normalized := make(map[string]struct{}, len(origins))
	if len(origins) == 0 {
		return normalized, true
	}

	for _, origin := range origins {
		origin = strings.TrimSpace(origin)
		if origin == "" {
			continue
		}
		if origin == "*" {
			return map[string]struct{}{}, true
		}

		normalizedOrigin, ok := normalizeOrigin(origin)
		if !ok {
			continue
		}
		normalized[normalizedOrigin] = struct{}{}
	}

	return normalized, len(normalized) == 0
}

func originAllowed(allowed map[string]struct{}, origin string) bool {
	_, ok := allowed[origin]
	return ok
}

func normalizeOrigin(raw string) (string, bool) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", false
	}

	return strings.ToLower(parsed.Scheme) + "://" + strings.ToLower(parsed.Host), true
}

func withDefaultMethods(methods []string) []string {
	if len(methods) == 0 {
		return []string{http.MethodGet, http.MethodPost, http.MethodOptions}
	}
	return methods
}
