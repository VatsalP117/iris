export interface IrisConfig {
  host: string;
  siteId: string;
  autocapture?: boolean;
  debug?: boolean;
}

export type EventPayload = {
  n: string;
  u: string;
  d: string;
  r: string | null;
  w: number;
  p?: Record<string, any>;
};
