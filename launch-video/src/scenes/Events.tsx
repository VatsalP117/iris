import { Video } from "@remotion/media";
import { AbsoluteFill, Easing, interpolate, staticFile, useCurrentFrame } from "remotion";

import { Paper } from "../components/Paper";
import { colors, ease, mono } from "../design";

export const Events: React.FC = () => {
  const frame = useCurrentFrame();
  return (
    <AbsoluteFill style={{ alignItems: "center", justifyContent: "center" }}>
      <Paper />
      <div
        style={{
          border: `1px solid ${colors.rule}`,
          boxShadow: "0 38px 95px rgba(21,21,18,.2)",
          height: 940,
          overflow: "hidden",
          position: "relative",
          scale: interpolate(frame, [0, 30, 329], [0.92, 0.98, 1.07], {
            extrapolateLeft: "clamp",
            extrapolateRight: "clamp",
            easing: Easing.bezier(...ease),
            output: "perceptual-scale",
          }),
          width: 1671,
        }}
      >
        <Video
          name="Custom events interaction"
          src={staticFile("clips/events-interaction.webm")}
          trimBefore={16}
          playbackRate={1.12}
          muted
          objectFit="cover"
          style={{ height: "100%", width: "100%" }}
        />
      </div>
      <div style={{ bottom: 62, color: colors.ink, fontFamily: mono, fontSize: 17, letterSpacing: ".13em", position: "absolute", right: 86, textTransform: "uppercase" }}>
        Track what matters · Select · Search · Learn
      </div>
    </AbsoluteFill>
  );
};
