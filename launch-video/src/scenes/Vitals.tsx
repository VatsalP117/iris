import { Video } from "@remotion/media";
import { AbsoluteFill, Easing, interpolate, staticFile, useCurrentFrame } from "remotion";

import { Paper } from "../components/Paper";
import { colors, ease, mono } from "../design";

export const Vitals: React.FC = () => {
  const frame = useCurrentFrame();
  return (
    <AbsoluteFill style={{ alignItems: "center", justifyContent: "center" }}>
      <Paper band />
      <div
        style={{
          border: `1px solid ${colors.rule}`,
          boxShadow: "0 38px 95px rgba(21,21,18,.2)",
          height: 940,
          overflow: "hidden",
          position: "relative",
          scale: interpolate(frame, [0, 26, 239], [0.92, 0.99, 1.06], {
            extrapolateLeft: "clamp",
            extrapolateRight: "clamp",
            easing: Easing.bezier(...ease),
            output: "perceptual-scale",
          }),
          width: 1671,
        }}
      >
        <Video
          name="Web Vitals interaction"
          src={staticFile("clips/vitals-interaction.webm")}
          trimBefore={14}
          playbackRate={1.1}
          muted
          objectFit="cover"
          style={{ height: "100%", width: "100%" }}
        />
      </div>
      <div style={{ bottom: 62, color: colors.green, fontFamily: mono, fontSize: 17, letterSpacing: ".13em", position: "absolute", right: 86, textTransform: "uppercase" }}>
        See how it feels · Real users · Real performance
      </div>
    </AbsoluteFill>
  );
};
