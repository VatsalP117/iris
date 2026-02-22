export interface AutocaptureConfig {
  pageviews?: boolean;
  webvitals?: boolean;
  clicks?: boolean;
}

export interface IrisConfig {
  host: string;
  siteId: string;
  autocapture?: AutocaptureConfig | false;
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
