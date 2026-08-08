import { AbsoluteFill, Easing, interpolate, useCurrentFrame } from "remotion";

import { Paper } from "../components/Paper";
import { colors, ease, mono, sans } from "../design";

const fragments = [
  ["SESSION_7A92", 180, 170],
  ["COOKIE_ID", 1310, 210],
  ["UTM_SOURCE", 250, 730],
  ["DEVICE_FINGERPRINT", 1160, 770],
  ["38.2%", 1490, 520],
  ["12,847", 420, 420],
  ["ATTRIBUTION", 870, 140],
  ["PROFILE_1842", 820, 820],
] as const;

export const Noise: React.FC = () => {
  const frame = useCurrentFrame();
  return (
    <AbsoluteFill>
      <Paper />
      {fragments.map(([label, left, top], index) => (
        <div
          key={label}
          style={{
            backgroundColor: index % 3 === 0 ? colors.ink : colors.raised,
            border: `1px solid ${index % 3 === 0 ? colors.ink : colors.rule}`,
            color: index % 3 === 0 ? colors.white : colors.muted,
            filter: `blur(${interpolate(frame, [48, 70], [0, 12], {
              extrapolateLeft: "clamp",
              extrapolateRight: "clamp",
            })}px)`,
            fontFamily: mono,
            fontSize: index === 4 || index === 5 ? 42 : 19,
            left,
            opacity: interpolate(frame, [index * 3, index * 3 + 8, 48, 68], [0, 1, 1, 0], {
              extrapolateLeft: "clamp",
              extrapolateRight: "clamp",
            }),
            padding: index === 4 || index === 5 ? "26px 34px" : "15px 22px",
            position: "absolute",
            rotate: `${(index % 2 === 0 ? -1 : 1) * (index + 1) * 0.9}deg`,
            scale: interpolate(frame, [index * 4, index * 4 + 15], [0.7, 1], {
              extrapolateLeft: "clamp",
              extrapolateRight: "clamp",
              easing: Easing.bezier(...ease),
              output: "perceptual-scale",
            }),
            top,
          }}
        >
          {label}
        </div>
      ))}
      <div
        style={{
          color: colors.ink,
          fontFamily: sans,
          fontSize: 126,
          fontWeight: 500,
          left: 130,
          letterSpacing: "-.06em",
          lineHeight: 0.94,
          opacity: interpolate(frame, [68, 82, 126, 146], [0, 1, 1, 0], {
            extrapolateLeft: "clamp",
            extrapolateRight: "clamp",
          }),
          position: "absolute",
          top: 420,
        }}
      >
        Analytics got noisy.
      </div>
    </AbsoluteFill>
  );
};
