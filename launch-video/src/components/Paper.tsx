import { AbsoluteFill, Easing, interpolate, useCurrentFrame } from "remotion";

import { colors, ease } from "../design";

interface Props {
  band?: boolean;
  dark?: boolean;
  grid?: boolean;
}

export const Paper: React.FC<Props> = ({ band = false, dark = false, grid = true }) => {
  const frame = useCurrentFrame();
  const bandProgress = interpolate(frame, [0, 48], [-18, 7], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
    easing: Easing.bezier(...ease),
  });

  return (
    <AbsoluteFill
      style={{
        backgroundColor: dark ? colors.black : colors.paper,
        backgroundImage: grid && !dark
          ? "linear-gradient(rgba(21,21,18,.026) 1px, transparent 1px), linear-gradient(90deg, rgba(21,21,18,.026) 1px, transparent 1px)"
          : undefined,
        backgroundSize: "72px 72px",
        overflow: "hidden",
      }}
    >
      {band ? (
        <div
          style={{
            backgroundColor: dark ? "rgba(101,88,239,.2)" : "rgba(101,88,239,.08)",
            height: 310,
            left: -180,
            position: "absolute",
            rotate: "-10deg",
            top: `${bandProgress}%`,
            width: 2350,
          }}
        />
      ) : null}
    </AbsoluteFill>
  );
};
