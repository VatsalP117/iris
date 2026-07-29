# Iris Browser Reliability Report

**Verdict:** FAIL

Browser: `/Applications/Google Chrome.app/Contents/MacOS/Google Chrome`

## Category summary

| Category | Passed | Failed | Total |
|---|---:|---:|---:|
| sdk-flow | 7 | 7 | 14 |
| delivery-chaos | 6 | 5 | 11 |
| state-machine | 0 | 1 | 1 |
| framework | 1 | 1 | 2 |
| lifecycle | 5 | 2 | 7 |

## Scenarios

| Category | Scenario | Result | Expected | Actual |
|---|---|---|---:|---:|
| sdk-flow | initial-pageview | PASS | 1 | 1 |
| sdk-flow | start-is-idempotent | PASS | 1 | 1 |
| sdk-flow | push-state-navigation | PASS | 2 | 2 |
| sdk-flow | same-url-push-state-deduplicated | FAIL | 1 | 2 |
| sdk-flow | replace-state-navigation | FAIL | 2 | 1 |
| sdk-flow | hash-navigation | PASS | 2 | 2 |
| sdk-flow | multiple-instances-do-not-double-count | FAIL | 1 | 2 |
| sdk-flow | stop-removes-navigation-tracking | PASS | 1 | 1 |
| sdk-flow | rejected-beacon-falls-back-to-fetch | FAIL | 1 | 0 |
| sdk-flow | transient-server-failure-retries | FAIL | 1 | 0 |
| sdk-flow | pagehide-flushes-batch | PASS | 1 | 1 |
| sdk-flow | storage-unavailable-falls-back-to-memory | PASS | 2 | 2 |
| sdk-flow | multi-tab-identity | FAIL | 2 | 2 |
| delivery-chaos | offline-event-retries-when-online | FAIL | 1 | 0 |
| sdk-flow | sdk-main-thread-overhead-1000-events | FAIL | 1000 | 270 |
| state-machine | generated-client-state-machine | FAIL | 88 | 92 |
| delivery-chaos | rate-limit-429-retries | FAIL | 1 | 0 |
| delivery-chaos | retryable-408-eventually-delivers | PASS | 1 | 1 |
| delivery-chaos | retryable-502-eventually-delivers | FAIL | 1 | 0 |
| delivery-chaos | retryable-504-eventually-delivers | FAIL | 1 | 0 |
| delivery-chaos | permanent-400-is-not-retried | PASS | 0 | 0 |
| delivery-chaos | accepted-response-lost-does-not-duplicate | FAIL | 1 | 2 |
| delivery-chaos | slow-response-does-not-double-send | PASS | 1 | 1 |
| delivery-chaos | hung-connection-closes-then-retries | PASS | 1 | 1 |
| delivery-chaos | cross-origin-preflight-allows-delivery | PASS | 1 | 1 |
| delivery-chaos | cross-origin-preflight-rejection-blocks-delivery | PASS | 0 | 0 |
| framework | react-router-declarative-navigation | PASS | 3 | 3 |
| framework | react-router-data-navigation-and-redirect | FAIL | 4 | 3 |
| lifecycle | actual-bfcache-restore-counts-return-navigation | FAIL | 2 | 1 |
| lifecycle | synthetic-pageshow-without-navigation-does-not-count | PASS | 1 | 1 |
| lifecycle | chromium-freeze-resume-preserves-navigation-tracking | PASS | 2 | 2 |
| lifecycle | hidden-visibility-flushes-batch-once | PASS | 1 | 1 |
| lifecycle | duplicated-tab-keeps-visitor-but-rotates-session | FAIL | 2 | 2 |
| lifecycle | navigation-before-sdk-load-uses-current-url | PASS | 1 | 1 |
| lifecycle | abrupt-page-close-flushes-queued-batch | PASS | 1 | 1 |

## Technical details

```json
[
  {
    "name": "initial-pageview",
    "category": "sdk-flow",
    "passed": true,
    "expectedEvents": 1,
    "actualEvents": 1,
    "requestAttempts": 1,
    "eventNames": [
      "$pageview"
    ],
    "bytes": 222,
    "consoleErrors": []
  },
  {
    "name": "start-is-idempotent",
    "category": "sdk-flow",
    "passed": true,
    "expectedEvents": 1,
    "actualEvents": 1,
    "requestAttempts": 1,
    "eventNames": [
      "$pageview"
    ],
    "bytes": 225,
    "consoleErrors": []
  },
  {
    "name": "push-state-navigation",
    "category": "sdk-flow",
    "passed": true,
    "expectedEvents": 2,
    "actualEvents": 2,
    "requestAttempts": 2,
    "eventNames": [
      "$pageview",
      "$pageview"
    ],
    "bytes": 420,
    "consoleErrors": []
  },
  {
    "name": "same-url-push-state-deduplicated",
    "category": "sdk-flow",
    "passed": false,
    "expectedEvents": 1,
    "actualEvents": 2,
    "requestAttempts": 2,
    "eventNames": [
      "$pageview",
      "$pageview"
    ],
    "bytes": 476,
    "consoleErrors": []
  },
  {
    "name": "replace-state-navigation",
    "category": "sdk-flow",
    "passed": false,
    "expectedEvents": 2,
    "actualEvents": 1,
    "requestAttempts": 1,
    "eventNames": [
      "$pageview"
    ],
    "bytes": 230,
    "consoleErrors": []
  },
  {
    "name": "hash-navigation",
    "category": "sdk-flow",
    "passed": true,
    "expectedEvents": 2,
    "actualEvents": 2,
    "requestAttempts": 2,
    "eventNames": [
      "$pageview",
      "$pageview"
    ],
    "bytes": 450,
    "consoleErrors": []
  },
  {
    "name": "multiple-instances-do-not-double-count",
    "category": "sdk-flow",
    "passed": false,
    "expectedEvents": 1,
    "actualEvents": 2,
    "requestAttempts": 2,
    "eventNames": [
      "$pageview",
      "$pageview"
    ],
    "bytes": 488,
    "consoleErrors": []
  },
  {
    "name": "stop-removes-navigation-tracking",
    "category": "sdk-flow",
    "passed": true,
    "expectedEvents": 1,
    "actualEvents": 1,
    "requestAttempts": 1,
    "eventNames": [
      "$pageview"
    ],
    "bytes": 238,
    "consoleErrors": []
  },
  {
    "name": "rejected-beacon-falls-back-to-fetch",
    "category": "sdk-flow",
    "passed": false,
    "expectedEvents": 1,
    "actualEvents": 0,
    "requestAttempts": 0,
    "eventNames": [],
    "bytes": 0,
    "consoleErrors": []
  },
  {
    "name": "transient-server-failure-retries",
    "category": "sdk-flow",
    "passed": false,
    "expectedEvents": 1,
    "actualEvents": 0,
    "requestAttempts": 1,
    "eventNames": [],
    "bytes": 238,
    "consoleErrors": [
      "Failed to load resource: the server responded with a status of 503 (Service Unavailable)"
    ]
  },
  {
    "name": "pagehide-flushes-batch",
    "category": "sdk-flow",
    "passed": true,
    "expectedEvents": 1,
    "actualEvents": 1,
    "requestAttempts": 1,
    "eventNames": [
      "$pageview"
    ],
    "bytes": 230,
    "consoleErrors": []
  },
  {
    "name": "storage-unavailable-falls-back-to-memory",
    "category": "sdk-flow",
    "passed": true,
    "expectedEvents": 2,
    "actualEvents": 2,
    "requestAttempts": 2,
    "eventNames": [
      "$pageview",
      "$pageview"
    ],
    "bytes": 441,
    "consoleErrors": []
  },
  {
    "name": "multi-tab-identity",
    "category": "sdk-flow",
    "passed": false,
    "expectedEvents": 2,
    "actualEvents": 2,
    "sameVisitor": false,
    "distinctSessions": true
  },
  {
    "name": "offline-event-retries-when-online",
    "category": "delivery-chaos",
    "passed": false,
    "expectedEvents": 1,
    "actualEvents": 0
  },
  {
    "name": "sdk-main-thread-overhead-1000-events",
    "category": "sdk-flow",
    "passed": false,
    "expectedEvents": 1000,
    "actualEvents": 270,
    "enqueueDurationMS": 10.899999976158142,
    "bytes": 63907
  },
  {
    "name": "generated-client-state-machine",
    "category": "state-machine",
    "passed": false,
    "expectedEvents": 88,
    "actualEvents": 92,
    "traces": 16,
    "failedTraceCount": 12,
    "failedTraces": [
      {
        "trace": 0,
        "actions": [
          "start",
          "start",
          "push-same",
          "replace",
          "stop",
          "stop"
        ],
        "expected": 2,
        "actual": 2,
        "passed": false,
        "expectedRecords": [
          {
            "name": "$pageview",
            "location": "/fixture?scenario=state-machine-0"
          },
          {
            "name": "$pageview",
            "location": "/state/0/replace-3"
          }
        ],
        "actualRecords": [
          {
            "name": "$pageview",
            "location": "/fixture?scenario=state-machine-0"
          },
          {
            "name": "$pageview",
            "location": "/fixture?scenario=state-machine-0"
          }
        ]
      },
      {
        "trace": 4,
        "actions": [
          "start",
          "push-same",
          "push",
          "stop",
          "replace",
          "push-same",
          "start",
          "stop",
          "track",
          "start",
          "start",
          "push-same",
          "start",
          "stop",
          "hash",
          "replace",
          "hash",
          "push"
        ],
        "expected": 5,
        "actual": 7,
        "passed": false,
        "expectedRecords": [
          {
            "name": "$pageview",
            "location": "/fixture?scenario=state-machine-4"
          },
          {
            "name": "$pageview",
            "location": "/state/4/push-2"
          },
          {
            "name": "$pageview",
            "location": "/state/4/replace-4"
          },
          {
            "name": "state-machine",
            "location": "/state/4/replace-4"
          },
          {
            "name": "$pageview",
            "location": "/state/4/replace-4"
          }
        ],
        "actualRecords": [
          {
            "name": "$pageview",
            "location": "/fixture?scenario=state-machine-4"
          },
          {
            "name": "$pageview",
            "location": "/fixture?scenario=state-machine-4"
          },
          {
            "name": "$pageview",
            "location": "/state/4/push-2"
          },
          {
            "name": "$pageview",
            "location": "/state/4/replace-4"
          },
          {
            "name": "state-machine",
            "location": "/state/4/replace-4"
          },
          {
            "name": "$pageview",
            "location": "/state/4/replace-4"
          },
          {
            "name": "$pageview",
            "location": "/state/4/replace-4"
          }
        ]
      },
      {
        "trace": 5,
        "actions": [
          "replace",
          "replace",
          "start",
          "push",
          "track",
          "start",
          "start",
          "push-same",
          "stop",
          "hash",
          "replace",
          "push-same",
          "stop",
          "start",
          "hash",
          "replace",
          "push",
          "stop"
        ],
        "expected": 7,
        "actual": 7,
        "passed": false,
        "expectedRecords": [
          {
            "name": "$pageview",
            "location": "/state/5/replace-1"
          },
          {
            "name": "$pageview",
            "location": "/state/5/push-3"
          },
          {
            "name": "state-machine",
            "location": "/state/5/push-3"
          },
          {
            "name": "$pageview",
            "location": "/state/5/replace-10"
          },
          {
            "name": "$pageview",
            "location": "/state/5/replace-10#state-5-14"
          },
          {
            "name": "$pageview",
            "location": "/state/5/replace-15"
          },
          {
            "name": "$pageview",
            "location": "/state/5/push-16"
          }
        ],
        "actualRecords": [
          {
            "name": "$pageview",
            "location": "/state/5/replace-1"
          },
          {
            "name": "$pageview",
            "location": "/state/5/push-3"
          },
          {
            "name": "state-machine",
            "location": "/state/5/push-3"
          },
          {
            "name": "$pageview",
            "location": "/state/5/push-3"
          },
          {
            "name": "$pageview",
            "location": "/state/5/replace-10"
          },
          {
            "name": "$pageview",
            "location": "/state/5/replace-10#state-5-14"
          },
          {
            "name": "$pageview",
            "location": "/state/5/push-16"
          }
        ]
      },
      {
        "trace": 6,
        "actions": [
          "stop",
          "push-same",
          "hash",
          "start",
          "push",
          "start",
          "start",
          "push-same",
          "stop",
          "start",
          "replace",
          "start",
          "hash",
          "push-same",
          "stop",
          "stop",
          "replace",
          "track"
        ],
        "expected": 6,
        "actual": 7,
        "passed": false,
        "expectedRecords": [
          {
            "name": "$pageview",
            "location": "/fixture?scenario=state-machine-6#state-6-2"
          },
          {
            "name": "$pageview",
            "location": "/state/6/push-4"
          },
          {
            "name": "$pageview",
            "location": "/state/6/push-4"
          },
          {
            "name": "$pageview",
            "location": "/state/6/replace-10"
          },
          {
            "name": "$pageview",
            "location": "/state/6/replace-10#state-6-12"
          },
          {
            "name": "state-machine",
            "location": "/state/6/replace-16"
          }
        ],
        "actualRecords": [
          {
            "name": "$pageview",
            "location": "/fixture?scenario=state-machine-6#state-6-2"
          },
          {
            "name": "$pageview",
            "location": "/state/6/push-4"
          },
          {
            "name": "$pageview",
            "location": "/state/6/push-4"
          },
          {
            "name": "$pageview",
            "location": "/state/6/push-4"
          },
          {
            "name": "$pageview",
            "location": "/state/6/replace-10#state-6-12"
          },
          {
            "name": "$pageview",
            "location": "/state/6/replace-10#state-6-12"
          },
          {
            "name": "state-machine",
            "location": "/state/6/replace-16"
          }
        ]
      },
      {
        "trace": 7,
        "actions": [
          "stop",
          "push-same",
          "start",
          "track",
          "push",
          "push-same",
          "replace",
          "push",
          "stop",
          "track",
          "stop",
          "start",
          "stop",
          "replace",
          "push",
          "start",
          "hash",
          "start"
        ],
        "expected": 9,
        "actual": 9,
        "passed": false,
        "expectedRecords": [
          {
            "name": "$pageview",
            "location": "/fixture?scenario=state-machine-7"
          },
          {
            "name": "state-machine",
            "location": "/fixture?scenario=state-machine-7"
          },
          {
            "name": "$pageview",
            "location": "/state/7/push-4"
          },
          {
            "name": "$pageview",
            "location": "/state/7/replace-6"
          },
          {
            "name": "$pageview",
            "location": "/state/7/push-7"
          },
          {
            "name": "state-machine",
            "location": "/state/7/push-7"
          },
          {
            "name": "$pageview",
            "location": "/state/7/push-7"
          },
          {
            "name": "$pageview",
            "location": "/state/7/push-14"
          },
          {
            "name": "$pageview",
            "location": "/state/7/push-14#state-7-16"
          }
        ],
        "actualRecords": [
          {
            "name": "$pageview",
            "location": "/fixture?scenario=state-machine-7"
          },
          {
            "name": "state-machine",
            "location": "/fixture?scenario=state-machine-7"
          },
          {
            "name": "$pageview",
            "location": "/state/7/push-4"
          },
          {
            "name": "$pageview",
            "location": "/state/7/push-4"
          },
          {
            "name": "$pageview",
            "location": "/state/7/push-7"
          },
          {
            "name": "state-machine",
            "location": "/state/7/push-7"
          },
          {
            "name": "$pageview",
            "location": "/state/7/push-7"
          },
          {
            "name": "$pageview",
            "location": "/state/7/push-14"
          },
          {
            "name": "$pageview",
            "location": "/state/7/push-14#state-7-16"
          }
        ]
      },
      {
        "trace": 9,
        "actions": [
          "push",
          "stop",
          "start",
          "push",
          "track",
          "replace",
          "replace",
          "start",
          "replace",
          "stop",
          "push-same",
          "stop",
          "replace",
          "push",
          "track",
          "push-same",
          "start",
          "start"
        ],
        "expected": 8,
        "actual": 5,
        "passed": false,
        "expectedRecords": [
          {
            "name": "$pageview",
            "location": "/state/9/push-0"
          },
          {
            "name": "$pageview",
            "location": "/state/9/push-3"
          },
          {
            "name": "state-machine",
            "location": "/state/9/push-3"
          },
          {
            "name": "$pageview",
            "location": "/state/9/replace-5"
          },
          {
            "name": "$pageview",
            "location": "/state/9/replace-6"
          },
          {
            "name": "$pageview",
            "location": "/state/9/replace-8"
          },
          {
            "name": "state-machine",
            "location": "/state/9/push-13"
          },
          {
            "name": "$pageview",
            "location": "/state/9/push-13"
          }
        ],
        "actualRecords": [
          {
            "name": "$pageview",
            "location": "/state/9/push-0"
          },
          {
            "name": "$pageview",
            "location": "/state/9/push-3"
          },
          {
            "name": "state-machine",
            "location": "/state/9/push-3"
          },
          {
            "name": "state-machine",
            "location": "/state/9/push-13"
          },
          {
            "name": "$pageview",
            "location": "/state/9/push-13"
          }
        ]
      },
      {
        "trace": 10,
        "actions": [
          "push-same",
          "hash",
          "start",
          "stop",
          "start",
          "replace",
          "push-same",
          "start",
          "push-same",
          "hash",
          "start",
          "stop",
          "stop",
          "push",
          "start",
          "start",
          "push",
          "stop"
        ],
        "expected": 6,
        "actual": 7,
        "passed": false,
        "expectedRecords": [
          {
            "name": "$pageview",
            "location": "/fixture?scenario=state-machine-10#state-10-1"
          },
          {
            "name": "$pageview",
            "location": "/fixture?scenario=state-machine-10#state-10-1"
          },
          {
            "name": "$pageview",
            "location": "/state/10/replace-5"
          },
          {
            "name": "$pageview",
            "location": "/state/10/replace-5#state-10-9"
          },
          {
            "name": "$pageview",
            "location": "/state/10/push-13"
          },
          {
            "name": "$pageview",
            "location": "/state/10/push-16"
          }
        ],
        "actualRecords": [
          {
            "name": "$pageview",
            "location": "/fixture?scenario=state-machine-10#state-10-1"
          },
          {
            "name": "$pageview",
            "location": "/fixture?scenario=state-machine-10#state-10-1"
          },
          {
            "name": "$pageview",
            "location": "/state/10/replace-5"
          },
          {
            "name": "$pageview",
            "location": "/state/10/replace-5"
          },
          {
            "name": "$pageview",
            "location": "/state/10/replace-5#state-10-9"
          },
          {
            "name": "$pageview",
            "location": "/state/10/push-13"
          },
          {
            "name": "$pageview",
            "location": "/state/10/push-16"
          }
        ]
      },
      {
        "trace": 11,
        "actions": [
          "track",
          "start",
          "hash",
          "start",
          "push",
          "push-same",
          "push-same",
          "start",
          "stop",
          "push-same",
          "push",
          "push",
          "push",
          "replace",
          "stop",
          "stop",
          "push",
          "hash"
        ],
        "expected": 4,
        "actual": 6,
        "passed": false,
        "expectedRecords": [
          {
            "name": "state-machine",
            "location": "/fixture?scenario=state-machine-11"
          },
          {
            "name": "$pageview",
            "location": "/fixture?scenario=state-machine-11"
          },
          {
            "name": "$pageview",
            "location": "/fixture?scenario=state-machine-11#state-11-2"
          },
          {
            "name": "$pageview",
            "location": "/state/11/push-4"
          }
        ],
        "actualRecords": [
          {
            "name": "state-machine",
            "location": "/fixture?scenario=state-machine-11"
          },
          {
            "name": "$pageview",
            "location": "/fixture?scenario=state-machine-11"
          },
          {
            "name": "$pageview",
            "location": "/fixture?scenario=state-machine-11#state-11-2"
          },
          {
            "name": "$pageview",
            "location": "/state/11/push-4"
          },
          {
            "name": "$pageview",
            "location": "/state/11/push-4"
          },
          {
            "name": "$pageview",
            "location": "/state/11/push-4"
          }
        ]
      },
      {
        "trace": 12,
        "actions": [
          "stop",
          "replace",
          "stop",
          "hash",
          "replace",
          "track",
          "push",
          "stop",
          "stop",
          "start",
          "track",
          "push-same",
          "start",
          "stop",
          "push-same",
          "track",
          "start",
          "stop"
        ],
        "expected": 5,
        "actual": 6,
        "passed": false,
        "expectedRecords": [
          {
            "name": "state-machine",
            "location": "/state/12/replace-4"
          },
          {
            "name": "$pageview",
            "location": "/state/12/push-6"
          },
          {
            "name": "state-machine",
            "location": "/state/12/push-6"
          },
          {
            "name": "state-machine",
            "location": "/state/12/push-6"
          },
          {
            "name": "$pageview",
            "location": "/state/12/push-6"
          }
        ],
        "actualRecords": [
          {
            "name": "state-machine",
            "location": "/state/12/replace-4"
          },
          {
            "name": "$pageview",
            "location": "/state/12/push-6"
          },
          {
            "name": "state-machine",
            "location": "/state/12/push-6"
          },
          {
            "name": "$pageview",
            "location": "/state/12/push-6"
          },
          {
            "name": "state-machine",
            "location": "/state/12/push-6"
          },
          {
            "name": "$pageview",
            "location": "/state/12/push-6"
          }
        ]
      },
      {
        "trace": 13,
        "actions": [
          "track",
          "stop",
          "replace",
          "hash",
          "stop",
          "push",
          "push-same",
          "push-same",
          "start",
          "track",
          "push",
          "track",
          "start",
          "replace",
          "start",
          "start",
          "hash",
          "hash"
        ],
        "expected": 8,
        "actual": 7,
        "passed": false,
        "expectedRecords": [
          {
            "name": "state-machine",
            "location": "/fixture?scenario=state-machine-13"
          },
          {
            "name": "$pageview",
            "location": "/state/13/push-5"
          },
          {
            "name": "state-machine",
            "location": "/state/13/push-5"
          },
          {
            "name": "$pageview",
            "location": "/state/13/push-10"
          },
          {
            "name": "state-machine",
            "location": "/state/13/push-10"
          },
          {
            "name": "$pageview",
            "location": "/state/13/replace-13"
          },
          {
            "name": "$pageview",
            "location": "/state/13/replace-13#state-13-16"
          },
          {
            "name": "$pageview",
            "location": "/state/13/replace-13#state-13-17"
          }
        ],
        "actualRecords": [
          {
            "name": "state-machine",
            "location": "/fixture?scenario=state-machine-13"
          },
          {
            "name": "$pageview",
            "location": "/state/13/push-5"
          },
          {
            "name": "state-machine",
            "location": "/state/13/push-5"
          },
          {
            "name": "$pageview",
            "location": "/state/13/push-10"
          },
          {
            "name": "state-machine",
            "location": "/state/13/push-10"
          },
          {
            "name": "$pageview",
            "location": "/state/13/replace-13#state-13-16"
          },
          {
            "name": "$pageview",
            "location": "/state/13/replace-13#state-13-17"
          }
        ]
      }
    ],
    "seed": 437635838
  },
  {
    "name": "rate-limit-429-retries",
    "category": "delivery-chaos",
    "passed": false,
    "expectedEvents": 1,
    "actualEvents": 0,
    "requestAttempts": 1,
    "eventNames": [],
    "bytes": 228,
    "consoleErrors": [
      "Failed to load resource: the server responded with a status of 429 (Too Many Requests)"
    ]
  },
  {
    "name": "retryable-408-eventually-delivers",
    "category": "delivery-chaos",
    "passed": true,
    "expectedEvents": 1,
    "actualEvents": 1,
    "requestAttempts": 2,
    "eventNames": [
      "$pageview"
    ],
    "bytes": 478,
    "consoleErrors": []
  },
  {
    "name": "retryable-502-eventually-delivers",
    "category": "delivery-chaos",
    "passed": false,
    "expectedEvents": 1,
    "actualEvents": 0,
    "requestAttempts": 1,
    "eventNames": [],
    "bytes": 239,
    "consoleErrors": [
      "Failed to load resource: the server responded with a status of 502 (Bad Gateway)"
    ]
  },
  {
    "name": "retryable-504-eventually-delivers",
    "category": "delivery-chaos",
    "passed": false,
    "expectedEvents": 1,
    "actualEvents": 0,
    "requestAttempts": 1,
    "eventNames": [],
    "bytes": 239,
    "consoleErrors": [
      "Failed to load resource: the server responded with a status of 504 (Gateway Timeout)"
    ]
  },
  {
    "name": "permanent-400-is-not-retried",
    "category": "delivery-chaos",
    "passed": true,
    "expectedEvents": 0,
    "actualEvents": 0,
    "requestAttempts": 1,
    "eventNames": [],
    "bytes": 234,
    "consoleErrors": [
      "Failed to load resource: the server responded with a status of 400 (Bad Request)"
    ]
  },
  {
    "name": "accepted-response-lost-does-not-duplicate",
    "category": "delivery-chaos",
    "passed": false,
    "expectedEvents": 1,
    "actualEvents": 2,
    "requestAttempts": 2,
    "eventNames": [
      "$pageview",
      "$pageview"
    ],
    "bytes": 494,
    "consoleErrors": []
  },
  {
    "name": "slow-response-does-not-double-send",
    "category": "delivery-chaos",
    "passed": true,
    "expectedEvents": 1,
    "actualEvents": 1,
    "requestAttempts": 1,
    "eventNames": [
      "$pageview"
    ],
    "bytes": 240,
    "consoleErrors": []
  },
  {
    "name": "hung-connection-closes-then-retries",
    "category": "delivery-chaos",
    "passed": true,
    "expectedEvents": 1,
    "actualEvents": 1,
    "requestAttempts": 2,
    "eventNames": [
      "$pageview"
    ],
    "bytes": 482,
    "consoleErrors": []
  },
  {
    "name": "cross-origin-preflight-allows-delivery",
    "category": "delivery-chaos",
    "passed": true,
    "expectedEvents": 1,
    "actualEvents": 1,
    "requestAttempts": 1,
    "eventNames": [
      "$pageview"
    ],
    "bytes": 218,
    "consoleErrors": []
  },
  {
    "name": "cross-origin-preflight-rejection-blocks-delivery",
    "category": "delivery-chaos",
    "passed": true,
    "expectedEvents": 0,
    "actualEvents": 0,
    "requestAttempts": 0,
    "eventNames": [],
    "bytes": 0,
    "consoleErrors": [
      "Access to fetch at 'http://127.0.0.1:52565/api/event' from origin 'http://localhost:52565' has been blocked by CORS policy: Response to preflight request doesn't pass access control check: No 'Access-Control-Allow-Origin' header is present on the requested resource.",
      "Failed to load resource: net::ERR_FAILED"
    ]
  },
  {
    "name": "react-router-declarative-navigation",
    "category": "framework",
    "passed": true,
    "expectedEvents": 3,
    "actualEvents": 3,
    "requestAttempts": 3,
    "eventNames": [
      "$pageview",
      "$pageview",
      "$pageview"
    ],
    "bytes": 736,
    "consoleErrors": []
  },
  {
    "name": "react-router-data-navigation-and-redirect",
    "category": "framework",
    "passed": false,
    "expectedEvents": 4,
    "actualEvents": 3,
    "requestAttempts": 3,
    "eventNames": [
      "$pageview",
      "$pageview",
      "$pageview"
    ],
    "bytes": 697,
    "consoleErrors": []
  },
  {
    "name": "actual-bfcache-restore-counts-return-navigation",
    "category": "lifecycle",
    "passed": false,
    "expectedEvents": 2,
    "actualEvents": 1,
    "restoredFromBFCache": true,
    "signals": [
      {
        "type": "pagehide",
        "persisted": true
      },
      {
        "type": "pageshow",
        "persisted": true
      }
    ],
    "navigationDiagnostics": {
      "type": "navigate",
      "notRestoredReasons": null
    }
  },
  {
    "name": "synthetic-pageshow-without-navigation-does-not-count",
    "category": "lifecycle",
    "passed": true,
    "expectedEvents": 1,
    "actualEvents": 1,
    "requestAttempts": 1,
    "eventNames": [
      "$pageview"
    ],
    "bytes": 258,
    "consoleErrors": []
  },
  {
    "name": "chromium-freeze-resume-preserves-navigation-tracking",
    "category": "lifecycle",
    "passed": true,
    "expectedEvents": 2,
    "actualEvents": 2,
    "requestAttempts": 2,
    "eventNames": [
      "$pageview",
      "$pageview"
    ],
    "bytes": 459,
    "consoleErrors": []
  },
  {
    "name": "hidden-visibility-flushes-batch-once",
    "category": "lifecycle",
    "passed": true,
    "expectedEvents": 1,
    "actualEvents": 1,
    "requestAttempts": 1,
    "eventNames": [
      "queued-before-hidden"
    ],
    "bytes": 255,
    "consoleErrors": []
  },
  {
    "name": "duplicated-tab-keeps-visitor-but-rotates-session",
    "category": "lifecycle",
    "passed": false,
    "expectedEvents": 2,
    "actualEvents": 2,
    "sameVisitor": true,
    "distinctSessions": false
  },
  {
    "name": "navigation-before-sdk-load-uses-current-url",
    "category": "lifecycle",
    "passed": true,
    "expectedEvents": 1,
    "actualEvents": 1,
    "actualURL": "http://127.0.0.1:52565/early-before-sdk"
  },
  {
    "name": "abrupt-page-close-flushes-queued-batch",
    "category": "lifecycle",
    "passed": true,
    "expectedEvents": 1,
    "actualEvents": 1
  }
]
```
