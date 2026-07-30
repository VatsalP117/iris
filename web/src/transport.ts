import { IrisConfig, EventPayload, BatchConfig } from "./config";
import { API_ENDPOINTS, API_REQUEST_CONSTANTS } from "./constants";

const BATCH_DEFAULTS = {
  maxSize: 10,
  flushInterval: 5000,
  flushOnLeave: true,
} as const;

const MAX_DELIVERY_ATTEMPTS = 3;
const RETRY_DELAY_MS = 100;
const STALLED_REQUEST_MS = 450;
const RETRYABLE_STATUS_CODES = new Set([408, 429, 502, 503, 504]);

export class Transport {
  private queue: EventPayload[] = [];
  private timer: ReturnType<typeof setInterval> | null = null;
  private batchConfig: Required<BatchConfig> | null = null;

  constructor(private config: IrisConfig) {
    if (config.batching) {
      this.batchConfig = { ...BATCH_DEFAULTS, ...config.batching };
      this.startTimer();

      if (this.batchConfig.flushOnLeave) {
        this.handleVisibilityChange = this.handleVisibilityChange.bind(this);
        this.handlePageHide = this.handlePageHide.bind(this);
        document.addEventListener("visibilitychange", this.handleVisibilityChange);
        window.addEventListener("pagehide", this.handlePageHide);
      }
    }
  }

  send(payload: EventPayload) {
    if (this.batchConfig) {
      this.queue.push(payload);
      if (this.config.debug) console.log("Iris: Event queued", payload, `(${this.queue.length}/${this.batchConfig.maxSize})`);
      if (this.queue.length >= this.batchConfig.maxSize) {
        this.flush();
      }
    } else {
      this.sendImmediate(payload);
    }
  }

  flush(useBeacon = false) {
    if (this.queue.length === 0) return;

    const events = this.queue.splice(0);
    const url = `${this.config.host}/${API_ENDPOINTS.BATCH_EVENTS}`;
    const body = JSON.stringify(events);

    if (this.config.debug) console.log(`Iris: Flushing ${events.length} events`);

    if (useBeacon && navigator.sendBeacon) {
      const blob = new Blob([body], { type: API_REQUEST_CONSTANTS.CONTENT_TYPES.JSON });
      const queued = navigator.sendBeacon(url, blob);
      if (!queued) {
        void this.sendWithFetch(url, body, "Iris: Batch flush failed", true);
      }
    } else {
      void this.sendWithFetch(url, body, "Iris: Batch flush failed", useBeacon);
    }
  }

  destroy() {
    this.flush();

    if (this.timer) {
      clearInterval(this.timer);
      this.timer = null;
    }

    if (this.batchConfig?.flushOnLeave) {
      document.removeEventListener("visibilitychange", this.handleVisibilityChange);
      window.removeEventListener("pagehide", this.handlePageHide);
    }
  }

  // --- private ---

  private sendImmediate(payload: EventPayload) {
    const url = `${this.config.host}/${API_ENDPOINTS.SINGLE_EVENT}`;
    const body = JSON.stringify(payload);

    void this.sendWithFetch(url, body, "Iris: Failed to send", false);

    if (this.config.debug) console.log("Iris: Event Sent", payload);
  }

  private async sendWithFetch(
    url: string,
    body: string,
    errorMessage: string,
    keepalive: boolean,
  ) {
    for (let attempt = 1; attempt <= MAX_DELIVERY_ATTEMPTS; attempt++) {
      const requestStartedAt = performance.now();
      try {
        const response = await fetch(url, {
          method: API_REQUEST_CONSTANTS.METHODS.POST,
          body,
          keepalive,
          headers: { [API_REQUEST_CONSTANTS.HEADERS.CONTENT_TYPE]: API_REQUEST_CONSTANTS.CONTENT_TYPES.JSON },
        });
        if (
          !RETRYABLE_STATUS_CODES.has(response.status) ||
          attempt === MAX_DELIVERY_ATTEMPTS
        ) {
          return;
        }
      } catch (err) {
        const requestDuration = performance.now() - requestStartedAt;
        const requestWasStalled = requestDuration >= STALLED_REQUEST_MS;
        if (
          attempt === MAX_DELIVERY_ATTEMPTS ||
          (navigator.onLine && !requestWasStalled)
        ) {
          if (this.config.debug) console.error(errorMessage, err);
          return;
        }
      }
      await new Promise((resolve) => setTimeout(resolve, RETRY_DELAY_MS));
    }
  }

  private startTimer() {
    if (!this.batchConfig) return;
    this.timer = setInterval(() => this.flush(), this.batchConfig.flushInterval);
  }

  private handleVisibilityChange() {
    if (document.visibilityState === "hidden") {
      this.flush(true);
    }
  }

  private handlePageHide() {
    this.flush(true);
  }
}
