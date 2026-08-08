import { CanvasImage, Easing, interpolate, staticFile, useCurrentFrame } from "remotion";

import { colors, ease } from "../design";

interface Props {
  crop?: "full" | "metrics" | "chart" | "lower";
  file: "overview.png" | "events.png" | "vitals.png";
  delay?: number;
}

const cropStyles = {
  full: { height: 850, objectPosition: "50% 0%", width: 1511 },
  metrics: { height: 1160, objectPosition: "50% 13%", width: 2062 },
  chart: { height: 1300, objectPosition: "50% 48%", width: 2311 },
  lower: { height: 1210, objectPosition: "50% 88%", width: 2151 },
};

export const ProductShot: React.FC<Props> = ({ crop = "full", delay = 0, file }) => {
  const frame = useCurrentFrame();
  const sizing = cropStyles[crop];

  return (
    <div
      style={{
        backgroundColor: colors.raised,
        border: `1px solid ${colors.rule}`,
        boxShadow: "0 42px 90px rgba(24,20,16,.18)",
        height: 850,
        opacity: interpolate(frame, [delay, delay + 18], [0, 1], {
          extrapolateLeft: "clamp",
          extrapolateRight: "clamp",
        }),
        overflow: "hidden",
        position: "relative",
        scale: interpolate(frame, [delay, delay + 48], [0.94, 1], {
          extrapolateLeft: "clamp",
          extrapolateRight: "clamp",
          easing: Easing.bezier(...ease),
          output: "perceptual-scale",
        }),
        translate: `0 ${interpolate(frame, [delay, delay + 48], [90, 0], {
          extrapolateLeft: "clamp",
          extrapolateRight: "clamp",
          easing: Easing.bezier(...ease),
        })}px`,
        width: 1511,
      }}
    >
      <CanvasImage
        src={staticFile(`captures/${file}`)}
        style={{
          height: sizing.height,
          left: "50%",
          objectFit: "cover",
          objectPosition: sizing.objectPosition,
          position: "absolute",
          top: "50%",
          translate: "-50% -50%",
          width: sizing.width,
        }}
      />
    </div>
  );
};
