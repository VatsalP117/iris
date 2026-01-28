import { IrisConfig, EventPayload } from "./config";
import { Transport } from "./transport";
import { initAutoCapture } from "./autocapture";
import { initVitals } from "./vitals";

export class Iris {
  private transport: Transport;
  private config: IrisConfig;

  constructor(config: IrisConfig) {
    this.config = {
      autocapture: true,
      debug: false,
      ...config,
    };
    this.transport = new Transport(this.config);
  }

  public start() {
    this.trackPageview();
    this.enableHistoryPatch();

    if (this.config.autocapture) {
      initAutoCapture(this.track.bind(this));
    }
    initVitals(this.track.bind(this));
  }

  public track(name: string, props?: object) {
    const payload: EventPayload = {
      n: name,
      u: window.location.href,
      d: window.location.hostname,
      r: document.referrer || null,
      w: window.innerWidth,
      p: props,
    };
    this.transport.send(payload);
  }

  private trackPageview() {
    this.track("$pageview");
  }

  private enableHistoryPatch() {
    const originalPushState = history.pushState;
    history.pushState = (...args) => {
      originalPushState.apply(history, args);
      this.trackPageview();
    };

    window.addEventListener("popstate", () => {
      this.trackPageview();
    });
  }
}
