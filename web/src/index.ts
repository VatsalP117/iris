import { IrisConfig, AutocaptureConfig, BatchConfig, EventPayload } from "./config";
import { Transport } from "./transport";
import { initAutoCapture } from "./autocapture";
import { initVitals } from "./vitals";
import { generateId, getVisitorIdentity, getSessionId } from "./storage";

// Module-level state prevents multiple Iris instances from duplicating pageviews.
let pushStatePatched = false;
const pageviewInstances = new Set<Iris>();

export class Iris {
  private transport: Transport;
  private config: IrisConfig;
  private isStarted = false;
  private vitalsRunId = 0;
  private originalPushState: typeof history.pushState | null = null;
  private originalReplaceState: typeof history.replaceState | null = null;
  private autocaptureCleanup: (() => void) | null = null;
  private visitorIdentity = getVisitorIdentity();
  private pendingVisitorEvents = new Set<Omit<EventPayload, "vid">>();

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
    const vitalsRunId = ++this.vitalsRunId;

    const ac = this.config.autocapture as AutocaptureConfig | false | undefined;

    if (ac && ac.pageviews === true) {
      pageviewInstances.add(this);
      if (!pushStatePatched) {
        this.trackPageview();
        this.enableHistoryPatch();
      }
    }
    if (ac && ac.clicks === true) {
      this.autocaptureCleanup = initAutoCapture(this.track.bind(this));
    }
    if (ac && ac.webvitals === true) {
      void initVitals((name, props) => {
        if (this.isStarted && this.vitalsRunId === vitalsRunId) {
          this.track(name, props);
        }
      }).catch((err) => {
        if (this.config.debug) {
          console.error("Iris: Failed to initialize Web Vitals", err);
        }
      });
    }
  }

  public track(name: string, props?: object) {
    const payload: Omit<EventPayload, "vid"> = {
      id: generateId(),
      n: name,
      u: window.location.href,
      d: window.location.hostname,
      r: document.referrer || null,
      w: window.innerWidth,
      s: this.config.siteId,
      sid: getSessionId(),
      p: props as Record<string, any> | undefined,
    };
    this.pendingVisitorEvents.add(payload);
    this.attachIdentityLifecycleListeners();
    void this.visitorIdentity.ready
      .then((vid) => {
        if (this.pendingVisitorEvents.delete(payload)) {
          this.transport.send({ ...payload, vid });
        }
        if (this.pendingVisitorEvents.size === 0) {
          this.removeIdentityLifecycleListeners();
        }
      })
      .catch((err) => {
        if (this.config.debug) {
          console.error("Iris: Failed to initialize visitor ID", err);
        }
      });
  }

  private trackPageview() {
    this.track("$pageview");
  }

  private handlePopState = () => {
    this.trackPageview();
  };

  private handlePageShow = (event: PageTransitionEvent) => {
    if (event.persisted && event.isTrusted) {
      this.trackPageview();
    }
  };

  private handleIdentityVisibilityChange = () => {
    if (document.visibilityState === "hidden") {
      this.flushPendingVisitorEvents();
    }
  };

  private attachIdentityLifecycleListeners() {
    if (this.pendingVisitorEvents.size !== 1) return;
    document.addEventListener("visibilitychange", this.handleIdentityVisibilityChange);
    window.addEventListener("pagehide", this.flushPendingVisitorEvents);
  }

  private removeIdentityLifecycleListeners() {
    document.removeEventListener("visibilitychange", this.handleIdentityVisibilityChange);
    window.removeEventListener("pagehide", this.flushPendingVisitorEvents);
  }

  private flushPendingVisitorEvents = () => {
    for (const payload of this.pendingVisitorEvents) {
      this.transport.send({ ...payload, vid: this.visitorIdentity.current });
    }
    this.pendingVisitorEvents.clear();
    this.removeIdentityLifecycleListeners();
    this.transport.flush(true);
  };

  private enableHistoryPatch() {
    if (pushStatePatched) return;
    pushStatePatched = true;
    this.originalPushState = history.pushState;
    this.originalReplaceState = history.replaceState;
    const self = this;
    history.pushState = function (...args) {
      const previousUrl = window.location.href;
      self.originalPushState!.apply(history, args);
      if (window.location.href !== previousUrl) {
        self.trackPageview();
      }
    };
    history.replaceState = function (...args) {
      const previousUrl = window.location.href;
      self.originalReplaceState!.apply(history, args);
      if (window.location.href !== previousUrl) {
        self.trackPageview();
      }
    };
    window.addEventListener("popstate", this.handlePopState);
    window.addEventListener("pageshow", this.handlePageShow);
  }

  public stop() {
    if (!this.isStarted) return;
    this.isStarted = false;
    this.vitalsRunId++;

    this.flushPendingVisitorEvents();
    this.transport.destroy();

    if (this.autocaptureCleanup) {
      this.autocaptureCleanup();
      this.autocaptureCleanup = null;
    }

    if (this.originalPushState) {
      history.pushState = this.originalPushState;
      this.originalPushState = null;
      history.replaceState = this.originalReplaceState!;
      this.originalReplaceState = null;
      pushStatePatched = false;
    }
    window.removeEventListener("popstate", this.handlePopState);
    window.removeEventListener("pageshow", this.handlePageShow);

    pageviewInstances.delete(this);
    if (!pushStatePatched) {
      const nextInstance = pageviewInstances.values().next().value;
      nextInstance?.enableHistoryPatch();
    }
  }
}

export type { AutocaptureConfig, BatchConfig, IrisConfig };
