import { AbsoluteFill, Easing, interpolate, useCurrentFrame } from "remotion";

import { IrisFlower } from "../components/IrisFlower";
import { Paper } from "../components/Paper";
import { colors, ease, mono, sans, serif } from "../design";

export const End: React.FC = () => {
  const frame = useCurrentFrame();
  return (
    <AbsoluteFill style={{ alignItems: "center", justifyContent: "center" }}>
      <Paper band />
      <div
        style={{
          alignItems: "center",
          display: "flex",
          flexDirection: "column",
          opacity: interpolate(frame, [0, 35, 345, 405], [0, 1, 1, 0], {
            extrapolateLeft: "clamp",
            extrapolateRight: "clamp",
          }),
          position: "relative",
          scale: interpolate(frame, [0, 55], [0.92, 1], {
            extrapolateLeft: "clamp",
            extrapolateRight: "clamp",
            easing: Easing.bezier(...ease),
            output: "perceptual-scale",
          }),
          textAlign: "center",
        }}
      >
        <IrisFlower color={colors.ink} size={98} />
        <div style={{ color: colors.ink, fontFamily: sans, fontSize: 62, fontWeight: 500, marginTop: 34 }}>Iris v1.0.0</div>
        <div style={{ color: colors.ink, fontFamily: serif, fontSize: 118, fontStyle: "italic", letterSpacing: "-.04em", lineHeight: 1, marginTop: 50 }}>
          Open-source analytics,<br /><span style={{ color: colors.violet }}>beautifully clear.</span>
        </div>
        <div style={{ borderTop: `1px solid ${colors.rule}`, color: colors.muted, fontFamily: mono, fontSize: 20, letterSpacing: ".08em", marginTop: 65, paddingTop: 24, width: 610 }}>
          github.com/VatsalP117/iris
        </div>
      </div>
    </AbsoluteFill>
  );
};
