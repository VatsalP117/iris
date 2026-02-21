import { IrisConfig, EventPayload } from "./config";

export class Transport {
  constructor(private config: IrisConfig) { }

  send(payload: EventPayload) {
    const url = `${this.config.host}/api/event`;
    const body = JSON.stringify(payload);

    if (navigator.sendBeacon) {
      const blob = new Blob([body], { type: "application/json" });
      navigator.sendBeacon(url, blob);
    } else {
      fetch(url, {
        method: "POST",
        body: body,
        keepalive: true,
        headers: { "Content-Type": "application/json" },
      }).catch((err) => {
        if (this.config.debug) console.error("Iris: Failed to send", err);
      });
    }

    if (this.config.debug) console.log("Iris: Event Sent", payload);
  }
}
