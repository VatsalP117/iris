import { IrisConfig, EventPayload, BatchConfig } from "./config";
import { API_ENDPOINTS, API_REQUEST_CONSTANTS } from "./constants";

const BATCH_DEFAULTS = {
  maxSize: 10,
  flushInterval: 5000,
  flushOnLeave: true,
} as const;

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

  flush() {
    if (this.queue.length === 0) return;

    const events = this.queue.splice(0);
    const url = `${this.config.host}/${API_ENDPOINTS.BATCH_EVENTS}`;
    const body = JSON.stringify(events);

    if (this.config.debug) console.log(`Iris: Flushing ${events.length} events`);

    if (navigator.sendBeacon) {
      const blob = new Blob([body], { type: API_REQUEST_CONSTANTS.CONTENT_TYPES.JSON });
      navigator.sendBeacon(url, blob);
    } else {
      fetch(url, {
        method: API_REQUEST_CONSTANTS.METHODS.POST,
        body,
        keepalive: true,
        headers: { [API_REQUEST_CONSTANTS.HEADERS.CONTENT_TYPE]: API_REQUEST_CONSTANTS.CONTENT_TYPES.JSON },
      }).catch((err) => {
        if (this.config.debug) console.error("Iris: Batch flush failed", err);
      });
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

    if (navigator.sendBeacon) {
      const blob = new Blob([body], { type: API_REQUEST_CONSTANTS.CONTENT_TYPES.JSON });
      navigator.sendBeacon(url, blob);
    } else {
      fetch(url, {
        method: API_REQUEST_CONSTANTS.METHODS.POST,
        body,
        keepalive: true,
        headers: { [API_REQUEST_CONSTANTS.HEADERS.CONTENT_TYPE]: API_REQUEST_CONSTANTS.CONTENT_TYPES.JSON },
      }).catch((err) => {
        if (this.config.debug) console.error("Iris: Failed to send", err);
      });
    }

    if (this.config.debug) console.log("Iris: Event Sent", payload);
  }

  private startTimer() {
    if (!this.batchConfig) return;
    this.timer = setInterval(() => this.flush(), this.batchConfig.flushInterval);
  }

  private handleVisibilityChange() {
    if (document.visibilityState === "hidden") {
      this.flush();
    }
  }

  private handlePageHide() {
    this.flush();
  }
}
