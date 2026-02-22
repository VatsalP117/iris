import { IrisConfig, EventPayload } from "./config";
import { API_ENDPOINTS, API_REQUEST_CONSTANTS } from "./constants";

export class Transport {
  constructor(private config: IrisConfig) { }

  send(payload: EventPayload) {
    const url = `${this.config.host}/${API_ENDPOINTS.SINGLE_EVENT}`;
    const body = JSON.stringify(payload);

    if (navigator.sendBeacon) {
      const blob = new Blob([body], { type: API_REQUEST_CONSTANTS.CONTENT_TYPES.JSON });
      navigator.sendBeacon(url, blob);
    } else {
      fetch(url, {
        method: API_REQUEST_CONSTANTS.METHODS.POST,
        body: body,
        keepalive: true,
        headers: { [API_REQUEST_CONSTANTS.HEADERS.CONTENT_TYPE]: API_REQUEST_CONSTANTS.CONTENT_TYPES.JSON },
      }).catch((err) => {
        if (this.config.debug) console.error("Iris: Failed to send", err);
      });
    }

    if (this.config.debug) console.log("Iris: Event Sent", payload);
  }
}
