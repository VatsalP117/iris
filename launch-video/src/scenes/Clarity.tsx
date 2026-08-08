import { AbsoluteFill, Easing, interpolate, useCurrentFrame } from "remotion";

import { Paper } from "../components/Paper";
import { SignalDot } from "../components/SignalDot";
import { Eyebrow, RevealText } from "../components/Type";
import { colors, ease } from "../design";

export const Clarity: React.FC = () => {
  const frame = useCurrentFrame();
  return (
    <AbsoluteFill style={{ justifyContent: "center", padding: "0 145px" }}>
      <Paper band />
      <div style={{ alignItems: "center", display: "flex", gap: 32, position: "relative" }}>
        <SignalDot delay={10} size={24} />
        <Eyebrow delay={15}>Introducing Iris</Eyebrow>
      </div>
      <div style={{ marginTop: 34, position: "relative" }}>
        <RevealText delay={25} display>
          Iris makes it <span style={{ color: colors.violet, fontStyle: "italic" }}>clear.</span>
        </RevealText>
      </div>
      <div
        style={{
          backgroundColor: colors.ink,
          bottom: 175,
          height: 2,
          left: 145,
          position: "absolute",
          width: `${interpolate(frame, [70, 150], [0, 84], {
            extrapolateLeft: "clamp",
            extrapolateRight: "clamp",
            easing: Easing.bezier(...ease),
          })}%`,
        }}
      />
    </AbsoluteFill>
  );
};
