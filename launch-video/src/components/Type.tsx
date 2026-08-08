import { Easing, interpolate, useCurrentFrame } from "remotion";

import { colors, ease, mono, sans, serif } from "../design";

interface RevealTextProps {
  children: React.ReactNode;
  delay?: number;
  display?: boolean;
  style?: React.CSSProperties;
}

export const RevealText: React.FC<RevealTextProps> = ({ children, delay = 0, display = false, style }) => {
  const frame = useCurrentFrame();
  return (
    <div
      style={{
        clipPath: `inset(0 ${interpolate(frame, [delay, delay + 40], [100, 0], {
          extrapolateLeft: "clamp",
          extrapolateRight: "clamp",
          easing: Easing.bezier(...ease),
        })}% 0 0)`,
        color: colors.ink,
        fontFamily: display ? serif : sans,
        fontSize: display ? 146 : 112,
        fontWeight: display ? 400 : 500,
        letterSpacing: display ? "-0.045em" : "-0.055em",
        lineHeight: 0.94,
        ...style,
      }}
    >
      {children}
    </div>
  );
};

export const Eyebrow: React.FC<RevealTextProps> = ({ children, delay = 0, style }) => {
  const frame = useCurrentFrame();
  return (
    <div
      style={{
        color: colors.violet,
        fontFamily: mono,
        fontSize: 20,
        fontWeight: 500,
        letterSpacing: ".16em",
        opacity: interpolate(frame, [delay, delay + 18], [0, 1], {
          extrapolateLeft: "clamp",
          extrapolateRight: "clamp",
        }),
        textTransform: "uppercase",
        ...style,
      }}
    >
      {children}
    </div>
  );
};
