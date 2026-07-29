# Iris Browser Reliability Report

**Verdict:** FAIL

Browser: `/Applications/Google Chrome.app/Contents/MacOS/Google Chrome`

| Scenario | Result | Expected | Actual |
|---|---|---:|---:|
| initial-pageview | PASS | 1 | 1 |
| start-is-idempotent | PASS | 1 | 1 |
| push-state-navigation | PASS | 2 | 2 |
| same-url-push-state-deduplicated | FAIL | 1 | 2 |
| replace-state-navigation | FAIL | 2 | 1 |
| hash-navigation | PASS | 2 | 2 |
| multiple-instances-do-not-double-count | FAIL | 1 | 2 |
| stop-removes-navigation-tracking | PASS | 1 | 1 |
| rejected-beacon-falls-back-to-fetch | FAIL | 1 | 0 |
| transient-server-failure-retries | FAIL | 1 | 0 |
| pagehide-flushes-batch | PASS | 1 | 1 |
| storage-unavailable-falls-back-to-memory | PASS | 2 | 2 |
| multi-tab-identity | FAIL | 2 | 2 |
| offline-event-retries-when-online | FAIL | 1 | 0 |
| sdk-main-thread-overhead-1000-events | FAIL | 1000 | 270 |

## Technical details

```json
[
  {
    "name": "initial-pageview",
    "passed": true,
    "expectedEvents": 1,
    "actualEvents": 1,
    "eventNames": [
      "$pageview"
    ],
    "bytes": 222,
    "consoleErrors": []
  },
  {
    "name": "start-is-idempotent",
    "passed": true,
    "expectedEvents": 1,
    "actualEvents": 1,
    "eventNames": [
      "$pageview"
    ],
    "bytes": 225,
    "consoleErrors": []
  },
  {
    "name": "push-state-navigation",
    "passed": true,
    "expectedEvents": 2,
    "actualEvents": 2,
    "eventNames": [
      "$pageview",
      "$pageview"
    ],
    "bytes": 420,
    "consoleErrors": []
  },
  {
    "name": "same-url-push-state-deduplicated",
    "passed": false,
    "expectedEvents": 1,
    "actualEvents": 2,
    "eventNames": [
      "$pageview",
      "$pageview"
    ],
    "bytes": 476,
    "consoleErrors": []
  },
  {
    "name": "replace-state-navigation",
    "passed": false,
    "expectedEvents": 2,
    "actualEvents": 1,
    "eventNames": [
      "$pageview"
    ],
    "bytes": 230,
    "consoleErrors": []
  },
  {
    "name": "hash-navigation",
    "passed": true,
    "expectedEvents": 2,
    "actualEvents": 2,
    "eventNames": [
      "$pageview",
      "$pageview"
    ],
    "bytes": 450,
    "consoleErrors": []
  },
  {
    "name": "multiple-instances-do-not-double-count",
    "passed": false,
    "expectedEvents": 1,
    "actualEvents": 2,
    "eventNames": [
      "$pageview",
      "$pageview"
    ],
    "bytes": 488,
    "consoleErrors": []
  },
  {
    "name": "stop-removes-navigation-tracking",
    "passed": true,
    "expectedEvents": 1,
    "actualEvents": 1,
    "eventNames": [
      "$pageview"
    ],
    "bytes": 238,
    "consoleErrors": []
  },
  {
    "name": "rejected-beacon-falls-back-to-fetch",
    "passed": false,
    "expectedEvents": 1,
    "actualEvents": 0,
    "eventNames": [],
    "bytes": 0,
    "consoleErrors": []
  },
  {
    "name": "transient-server-failure-retries",
    "passed": false,
    "expectedEvents": 1,
    "actualEvents": 0,
    "eventNames": [],
    "bytes": 238,
    "consoleErrors": [
      "Failed to load resource: the server responded with a status of 503 (Service Unavailable)"
    ]
  },
  {
    "name": "pagehide-flushes-batch",
    "passed": true,
    "expectedEvents": 1,
    "actualEvents": 1,
    "eventNames": [
      "$pageview"
    ],
    "bytes": 230,
    "consoleErrors": []
  },
  {
    "name": "storage-unavailable-falls-back-to-memory",
    "passed": true,
    "expectedEvents": 2,
    "actualEvents": 2,
    "eventNames": [
      "$pageview",
      "$pageview"
    ],
    "bytes": 441,
    "consoleErrors": []
  },
  {
    "name": "multi-tab-identity",
    "passed": false,
    "expectedEvents": 2,
    "actualEvents": 2,
    "sameVisitor": false,
    "distinctSessions": true
  },
  {
    "name": "offline-event-retries-when-online",
    "passed": false,
    "expectedEvents": 1,
    "actualEvents": 0
  },
  {
    "name": "sdk-main-thread-overhead-1000-events",
    "passed": false,
    "expectedEvents": 1000,
    "actualEvents": 270,
    "enqueueDurationMS": 10.600000023841858,
    "bytes": 63907
  }
]
```
