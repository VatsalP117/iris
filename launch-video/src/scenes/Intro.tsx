import { AbsoluteFill, Easing, interpolate, useCurrentFrame } from "remotion";

import { IrisFlower } from "../components/IrisFlower";
import { Paper } from "../components/Paper";
import { colors, ease, mono, sans } from "../design";

export const Intro: React.FC = () => {
  const frame = useCurrentFrame();
  return (
    <AbsoluteFill style={{ alignItems: "center", justifyContent: "center" }}>
      <Paper grid={false} />
      <div
        style={{
          alignItems: "center",
          display: "flex",
          gap: 40,
          opacity: interpolate(frame, [10, 36, 238, 286], [0, 1, 1, 0], {
            extrapolateLeft: "clamp",
            extrapolateRight: "clamp",
          }),
          position: "relative",
          scale: interpolate(frame, [10, 70], [0.82, 1], {
            extrapolateLeft: "clamp",
            extrapolateRight: "clamp",
            easing: Easing.bezier(...ease),
            output: "perceptual-scale",
          }),
        }}
      >
        <IrisFlower color={colors.ink} size={132} />
        <div>
          <div style={{ color: colors.ink, fontFamily: sans, fontSize: 112, fontWeight: 500, letterSpacing: "-.06em" }}>Iris</div>
          <div style={{ color: colors.muted, fontFamily: mono, fontSize: 18, letterSpacing: ".17em", textTransform: "uppercase" }}>
            Open-source analytics · v1.0.0
          </div>
        </div>
      </div>
    </AbsoluteFill>
  );
};
