import { Easing, interpolate, useCurrentFrame } from "remotion";

import { colors, ease } from "../design";

interface Props {
  delay?: number;
  size?: number;
}

export const SignalDot: React.FC<Props> = ({ delay = 0, size = 18 }) => {
  const frame = useCurrentFrame();
  return (
    <div
      style={{
        backgroundColor: colors.violet,
        borderRadius: "50%",
        boxShadow: `0 0 0 ${interpolate(frame, [delay, delay + 24], [0, size * 1.5], {
          extrapolateLeft: "clamp",
          extrapolateRight: "clamp",
          easing: Easing.bezier(...ease),
        })}px rgba(101,88,239,.08)`,
        height: size,
        opacity: interpolate(frame, [delay, delay + 8], [0, 1], {
          extrapolateLeft: "clamp",
          extrapolateRight: "clamp",
        }),
        scale: interpolate(frame, [delay, delay + 18], [0.25, 1], {
          extrapolateLeft: "clamp",
          extrapolateRight: "clamp",
          easing: Easing.bezier(...ease),
          output: "perceptual-scale",
        }),
        width: size,
      }}
    />
  );
};
