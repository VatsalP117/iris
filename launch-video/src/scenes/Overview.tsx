import { Video } from "@remotion/media";
import { AbsoluteFill, Easing, interpolate, staticFile, useCurrentFrame } from "remotion";

import { Paper } from "../components/Paper";
import { colors, ease, mono } from "../design";

export const Overview: React.FC = () => {
  const frame = useCurrentFrame();
  return (
    <AbsoluteFill style={{ alignItems: "center", justifyContent: "center" }}>
      <Paper band />
      <div
        style={{
          border: `1px solid ${colors.rule}`,
          boxShadow: "0 38px 95px rgba(21,21,18,.2)",
          height: 940,
          opacity: interpolate(frame, [0, 10], [0, 1], { extrapolateLeft: "clamp", extrapolateRight: "clamp" }),
          overflow: "hidden",
          position: "relative",
          scale: interpolate(frame, [0, 36, 250, 359], [0.9, 0.96, 1.02, 1.06], {
            extrapolateLeft: "clamp",
            extrapolateRight: "clamp",
            easing: Easing.bezier(...ease),
            output: "perceptual-scale",
          }),
          width: 1671,
        }}
      >
        <Video
          name="Overview interaction"
          src={staticFile("clips/overview-interaction.webm")}
          trimBefore={18}
          playbackRate={1.08}
          muted
          objectFit="cover"
          style={{ height: "100%", width: "100%" }}
        />
      </div>
      <div style={{ bottom: 62, color: colors.ink, fontFamily: mono, fontSize: 17, letterSpacing: ".13em", position: "absolute", right: 86, textTransform: "uppercase" }}>
        Your traffic · Filter · Compare · Understand
      </div>
    </AbsoluteFill>
  );
};
