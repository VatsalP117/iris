export interface AutocaptureConfig {
  pageviews?: boolean;
  webvitals?: boolean;
  clicks?: boolean;
}

export interface BatchConfig {
  /** Max events to queue before flushing. Default: 10 */
  maxSize?: number;
  /** Flush interval in ms. Default: 5000 */
  flushInterval?: number;
  /** Flush remaining events on page hide / visibility change. Default: true */
  flushOnLeave?: boolean;
}

export interface IrisConfig {
  host: string;
  siteId: string;
  autocapture?: AutocaptureConfig | false;
  batching?: BatchConfig;
  debug?: boolean;
}

export type EventPayload = {
  n: string;    // event name
  u: string;    // URL
  d: string;    // domain
  r: string | null; // referrer
  w: number;    // screen width
  s: string;    // site ID
  sid: string;  // session ID
  vid: string;  // visitor ID (anonymous)
  p?: Record<string, any>; // custom properties
};
