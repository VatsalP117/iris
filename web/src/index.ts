import { IrisConfig, AutocaptureConfig, EventPayload } from "./config";
import { Transport } from "./transport";
import { initAutoCapture } from "./autocapture";
import { initVitals } from "./vitals";
import { getVisitorId, getSessionId } from "./storage";

export class Iris {
  private transport: Transport;
  private config: IrisConfig;

  constructor(config: IrisConfig) {
    this.config = {
      autocapture: false,
      debug: false,
      ...config,
    };
    this.transport = new Transport(this.config);
  }

  public start() {
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
    const originalPushState = history.pushState;
    history.pushState = (...args) => {
      originalPushState.apply(history, args);
      this.trackPageview();
    };

    window.addEventListener("popstate", this.handlePopState);
  }

  public stop() {
    window.removeEventListener("popstate", this.handlePopState);
    // Note: restoring the original pushState is complex because other libraries
    // (like Next.js/React Router) might also patch it. We leave it patched
    // but the popstate listener is cleanly removed.
  }
}

export type { AutocaptureConfig, IrisConfig };
