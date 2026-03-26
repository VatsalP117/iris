# Agent Guidelines for Iris Analytics

## Project Overview

Iris Analytics is a monorepo containing:
- **Go backend**: `cmd/server` (API server) + `pkg/` (core logic)
- **Dashboard**: React 18 + Vite app in `dashboard/`
- **SDK**: TypeScript package in `web/` (iris-analytics npm package)
- **Marketing**: React site in `marketing/`

## Build/Lint/Test Commands

### Go (Backend)

```bash
# Install dependencies
go mod download

# Build the server
go build -o dist/iris-server ./cmd/server

# Run all tests
go test ./...

# Run tests in a specific package
go test ./pkg/db
go test ./pkg/api

# Run a single test by name
go test -run TestGetStatsAndSitesUseSiteIDAcrossDomains ./pkg/db

# Run tests with verbose output
go test -v ./pkg/db
```

### JavaScript/TypeScript (pnpm monorepo)

```bash
# Install all dependencies (from root)
pnpm install

# Build all JS packages (web SDK + dashboard + marketing)
task build:js

# Build individual packages
pnpm --filter iris-analytics build    # Web SDK
pnpm --filter @iris/dashboard build    # Dashboard
pnpm --filter marketing build          # Marketing site

# Run development servers
task dev:backend    # Go API on :8080
task dev:dashboard  # Vite dev server on :5173 (proxies /api to :8080)
task dev:marketing  # Marketing site dev server

# Lint (marketing only - has ESLint configured)
pnpm --filter marketing lint

# Type-check marketing
pnpm --filter marketing build  # runs tsc -b before vite build
```

### Task Runner (Taskfile.yml)

```bash
task              # Default: runs task build
task dev          # Runs backend + dashboard concurrently
task build        # Builds both Go backend and all JS packages
task build:backend
task build:js
task publish:js   # Builds and publishes iris-analytics to npm
```

## Code Style Guidelines

### Go

**Formatting**: Run `gofmt` before committing. Use standard Go formatting conventions.

**Error Handling**:
- Return errors from functions; do not panic.
- Log errors with `log.Printf("[FunctionName] error: %v", err)` before returning.
- HTTP handlers use `http.Error(w, "message", statusCode)` for client errors.
- Database errors return `http.StatusInternalServerError`.

```go
// Good
if err := h.Repo.Insert(r.Context(), &event); err != nil {
    log.Printf("[TrackEvent] DB Insert error: %v", err)
    http.Error(w, "Failed to save event", http.StatusInternalServerError)
    return
}

// Bad - no logging
if err != nil {
    return err
}
```

**Imports**: Group stdlib first, then third-party, then internal packages:

```go
import (
    "encoding/json"
    "log"
    "net/http"
    "time"

    "github.com/VatsalP117/iris/pkg/core"
    "github.com/google/uuid"
)
```

**Struct Tags**: Use JSON field tags for public API structs.

**Context**: Pass `context.Context` as the first argument to repository/database methods.

### TypeScript/React

**Indentation**: 4 spaces (as per project convention in dashboard/web code).

**Naming**:
- Components: PascalCase (`StatsCards`, `TopPages`)
- Functions/variables: camelCase (`fetchAll`, `handlePreset`)
- Types/interfaces: PascalCase (`StatsResult`, `DateWindow`)
- Constants: PascalCase if exported, camelCase if module-scoped

**Imports**: Separate external and internal imports with a blank line:

```typescript
import { useState, useEffect, useCallback } from "react";
import { format, subDays, subHours } from "date-fns";
import { api, StatsResult, PageStat } from "./api";
import { StatsCards } from "./components/StatsCards";
```

**Exports**: Prefer named exports for utilities; default exports for page components.

**Types**: Use `interface` for object shapes, `type` for unions/intersections:

```typescript
interface Props {
    stats: StatsResult | null;
    loading: boolean;
}

type Tab = "overview" | "pages" | "referrers";
```

**Async/Error Handling**:
```typescript
// Use async/await with try/catch in React components
const fetchAll = useCallback(async (siteId: string) => {
    try {
        const data = await api.stats(siteId);
        setStats(data);
    } catch (err) {
        console.error("Iris: fetch error", err);
    }
}, []);
```

**Component Patterns**:
- Props interfaces should be defined above the component.
- Use `React.CSSProperties` for inline style objects.
- Event handlers: `handleEventName` naming convention.

### Package Structure

```
cmd/server/          # Entry point for Go server
pkg/api/             # HTTP handlers, CORS middleware
pkg/db/              # SQLite repository, query functions
pkg/core/            # Core domain types and interfaces
dashboard/src/       # React dashboard components
dashboard/src/components/
web/src/              # TypeScript SDK (iris-analytics)
marketing/src/       # Marketing site
```

### Testing

**Go Tests**:
- Test files: `*_test.go` in same package
- Use table-driven tests where appropriate
- Helper functions should take `*testing.T` and call `t.Helper()`
- Use `t.Fatalf` for fatal errors in setup, `t.Error` for assertion failures

```go
func newTestRepo(t *testing.T) *SqliteRepository {
    t.Helper()
    repo, err := NewSqliteDB(filepath.Join(t.TempDir(), "iris.db"))
    if err != nil {
        t.Fatalf("NewSqliteDB returned error: %v", err)
    }
    return repo
}
```

**Test Naming**: `TestFunctionName_Scenario` pattern.

### General

- Do not commit commented-out code.
- Keep functions small and focused.
- Document complex SQL queries with comments.
- Environment variables: use meaningful defaults, document in README.md.
