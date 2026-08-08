import { AbsoluteFill, Easing, interpolate, useCurrentFrame } from "remotion";

import { Paper } from "../components/Paper";
import { SignalDot } from "../components/SignalDot";
import { colors, ease, mono, sans } from "../design";

const lines = [
  { code: "# Browser / npm package", color: colors.muted },
  { code: "$ npm install iris-analytics", color: colors.white },
  { code: "", color: colors.white },
  { code: "# Server / Go binary", color: colors.muted },
  { code: "$ go build -o iris-server ./cmd/server", color: colors.white },
  { code: "", color: colors.white },
  { code: "autocapture: {", color: colors.violetSoft },
  { code: "  pageviews: true,", color: colors.white },
  { code: "  webvitals: true", color: colors.white },
  { code: "}", color: colors.violetSoft },
];

export const Setup: React.FC = () => {
  const frame = useCurrentFrame();
  return (
    <AbsoluteFill style={{ padding: "130px 145px" }}>
      <Paper />
      <div style={{ position: "relative", width: 1000 }}>
        <div style={{ alignItems: "baseline", display: "flex", gap: 22, whiteSpace: "nowrap" }}>
          <div style={{ color: colors.violet, fontFamily: mono, fontSize: 20, fontWeight: 500, letterSpacing: ".16em" }}>SETUP /</div>
          <div
            style={{
              clipPath: `inset(0 ${interpolate(frame, [10, 48], [100, 0], { extrapolateLeft: "clamp", extrapolateRight: "clamp", easing: Easing.bezier(...ease) })}% 0 0)`,
              color: colors.ink,
              fontFamily: sans,
              fontSize: 70,
              fontWeight: 500,
              letterSpacing: "-.055em",
              lineHeight: 1,
            }}
          >
            One package. One binary.
          </div>
        </div>
        <div style={{ color: colors.muted, fontFamily: mono, fontSize: 23, lineHeight: 1.7, marginLeft: 144, marginTop: 42 }}>
          npm package in the browser.<br />Go binary on your server.<br />Automatic pageviews and Web Vitals.
        </div>
      </div>
      <div
        style={{
          backgroundColor: colors.black,
          boxShadow: "0 45px 100px rgba(21,21,18,.22)",
          height: 700,
          opacity: interpolate(frame, [38, 70], [0, 1], { extrapolateLeft: "clamp", extrapolateRight: "clamp" }),
          padding: "58px 62px",
          position: "absolute",
          right: 80,
          scale: interpolate(frame, [38, 92], [0.93, 1], {
            extrapolateLeft: "clamp",
            extrapolateRight: "clamp",
            easing: Easing.bezier(...ease),
            output: "perceptual-scale",
          }),
          top: 165,
          width: 680,
        }}
      >
        <div style={{ display: "flex", gap: 10, marginBottom: 45 }}>
          {[colors.red, "#d99b25", colors.green].map((color) => <span key={color} style={{ backgroundColor: color, borderRadius: "50%", height: 12, width: 12 }} />)}
        </div>
        {lines.map((line, index) => (
          <div
            key={`${line.code}-${index}`}
            style={{
              color: line.color,
              fontFamily: mono,
              fontSize: 24,
              lineHeight: 1.7,
              opacity: interpolate(frame, [62 + index * 10, 72 + index * 10], [0, 1], {
                extrapolateLeft: "clamp",
                extrapolateRight: "clamp",
              }),
            }}
          >
            {line.code || "\u00a0"}
          </div>
        ))}
        <div style={{ bottom: 65, display: "flex", position: "absolute", right: 76 }}><SignalDot delay={170} size={18} /></div>
      </div>
    </AbsoluteFill>
  );
};
