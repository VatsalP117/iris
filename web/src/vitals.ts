import type { Metric } from "web-vitals";

type TrackFn = (name: string, props: object) => void;

export async function initVitals(trackFn: TrackFn) {
  const { onCLS, onINP, onLCP } = await import("web-vitals");

  const handleMetric = (metric: Metric) => {
    trackFn("$web_vital", {
      $id: metric.id,
      $name: metric.name,
      $val: metric.value,
      $rating: metric.rating,
    });
  };

  onCLS(handleMetric);
  onINP(handleMetric);
  onLCP(handleMetric);
}
