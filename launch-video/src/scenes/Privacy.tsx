import { AbsoluteFill, interpolate, useCurrentFrame } from "remotion";

import { IrisFlower } from "../components/IrisFlower";
import { Paper } from "../components/Paper";
import { colors, mono, sans } from "../design";

const promises = ["No cookies.", "No profiles.", "Your data stays yours."];

export const Privacy: React.FC = () => {
  const frame = useCurrentFrame();
  return (
    <AbsoluteFill style={{ alignItems: "center", justifyContent: "center" }}>
      <Paper dark grid={false} />
      <IrisFlower color={colors.violet} size={76} style={{ left: 120, position: "absolute", top: 95 }} />
      <div style={{ display: "grid", gap: 20, position: "relative", width: 1500 }}>
        {promises.map((promise, index) => (
          <div
            key={promise}
            style={{
              borderBottom: `1px solid rgba(250,249,245,.17)`,
              color: index === 2 ? colors.violetSoft : colors.white,
              fontFamily: sans,
              fontSize: index === 2 ? 118 : 96,
              fontWeight: 500,
              letterSpacing: "-.055em",
              lineHeight: 1,
              opacity: interpolate(frame, [index * 54, index * 54 + 24], [0, 1], {
                extrapolateLeft: "clamp",
                extrapolateRight: "clamp",
              }),
              padding: "18px 0 28px",
              translate: `0 ${interpolate(frame, [index * 54, index * 54 + 32], [55, 0], {
                extrapolateLeft: "clamp",
                extrapolateRight: "clamp",
              })}px`,
            }}
          >
            {promise}
          </div>
        ))}
      </div>
      <div style={{ bottom: 76, color: colors.muted, fontFamily: mono, fontSize: 18, letterSpacing: ".14em", position: "absolute", textTransform: "uppercase" }}>
        Self-hosted · Open source · Privacy-first
      </div>
    </AbsoluteFill>
  );
};
