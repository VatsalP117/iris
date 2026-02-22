import { IrisConfig, AutocaptureConfig, EventPayload } from "./config";
import { Transport } from "./transport";
import { initAutoCapture } from "./autocapture";
import { initVitals } from "./vitals";
import { getVisitorId, getSessionId } from "./storage";

export class Iris {
  private transport: Transport;
  private config: IrisConfig;
  private isStarted = false;
  private originalPushState: typeof history.pushState | null = null;

  constructor(config: IrisConfig) {
    this.config = {
      autocapture: false,
      debug: false,
      ...config,
    };
    this.transport = new Transport(this.config);
  }

  public start() {
    if (this.isStarted) return;
    this.isStarted = true;

    const ac = this.config.autocapture as AutocaptureConfig | false | undefined;

    if (ac && ac.pageviews !== false) {
      this.trackPageview();
      this.enableHistoryPatch();
    }
    if (ac && ac.clicks !== false) {
      initAutoCapture(this.track.bind(this));
    }
    if (ac && ac.webvitals !== false) {
      initVitals(this.track.bind(this));
    }
  }

  public track(name: string, props?: object) {
    const payload: EventPayload = {
      n: name,
      u: window.location.href,
      d: window.location.hostname,
      r: document.referrer || null,
      w: window.innerWidth,
      s: this.config.siteId,
      sid: getSessionId(),
      vid: getVisitorId(),
      p: props as Record<string, any> | undefined,
    };
    this.transport.send(payload);
  }

  private trackPageview() {
    this.track("$pageview");
  }

  private handlePopState = () => {
    this.trackPageview();
  };

  private enableHistoryPatch() {
    this.originalPushState = history.pushState;
    const self = this;
    history.pushState = function (...args) {
      self.originalPushState!.apply(history, args);
      self.trackPageview();
    };

    window.addEventListener("popstate", this.handlePopState);
  }

  public stop() {
    if (!this.isStarted) return;
    this.isStarted = false;

    if (this.originalPushState) {
      history.pushState = this.originalPushState;
      this.originalPushState = null;
    }
    window.removeEventListener("popstate", this.handlePopState);
  }
}

export type { AutocaptureConfig, IrisConfig };
