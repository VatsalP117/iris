import { Audio } from "@remotion/media";
import { TransitionSeries, linearTiming } from "@remotion/transitions";
import { fade } from "@remotion/transitions/fade";
import { wipe } from "@remotion/transitions/wipe";
import { Sequence, interpolate, staticFile, useVideoConfig } from "remotion";

import { Clarity } from "./scenes/Clarity";
import { End } from "./scenes/End";
import { Events } from "./scenes/Events";
import { Intro } from "./scenes/Intro";
import { Noise } from "./scenes/Noise";
import { Overview } from "./scenes/Overview";
import { Privacy } from "./scenes/Privacy";
import { Setup } from "./scenes/Setup";
import { Vitals } from "./scenes/Vitals";

const transitionFrames = 8;

const SoundEffect: React.FC<{ from: number; src: string; volume?: number }> = ({ from, src, volume = 0.24 }) => (
  <Sequence from={from} durationInFrames={120} premountFor={60}>
    <Audio src={staticFile(`audio/${src}`)} volume={volume} />
  </Sequence>
);

export const IrisLaunch: React.FC = () => {
  const { durationInFrames } = useVideoConfig();

  return (
    <>
      <TransitionSeries>
        <TransitionSeries.Sequence durationInFrames={120} premountFor={60} name="01 · Iris">
          <Intro />
        </TransitionSeries.Sequence>
        <TransitionSeries.Transition presentation={fade()} timing={linearTiming({ durationInFrames: transitionFrames })} />

        <TransitionSeries.Sequence durationInFrames={150} premountFor={60} name="02 · Noise">
          <Noise />
        </TransitionSeries.Sequence>
        <TransitionSeries.Transition presentation={wipe({ direction: "from-right" })} timing={linearTiming({ durationInFrames: transitionFrames })} />

        <TransitionSeries.Sequence durationInFrames={150} premountFor={60} name="03 · Clarity">
          <Clarity />
        </TransitionSeries.Sequence>
        <TransitionSeries.Transition presentation={fade()} timing={linearTiming({ durationInFrames: transitionFrames })} />

        <TransitionSeries.Sequence durationInFrames={240} premountFor={60} name="04 · Setup">
          <Setup />
        </TransitionSeries.Sequence>
        <TransitionSeries.Transition presentation={wipe({ direction: "from-bottom" })} timing={linearTiming({ durationInFrames: transitionFrames })} />

        <TransitionSeries.Sequence durationInFrames={360} premountFor={60} name="05 · Overview interaction">
          <Overview />
        </TransitionSeries.Sequence>
        <TransitionSeries.Transition presentation={fade()} timing={linearTiming({ durationInFrames: transitionFrames })} />

        <TransitionSeries.Sequence durationInFrames={330} premountFor={60} name="06 · Events interaction">
          <Events />
        </TransitionSeries.Sequence>
        <TransitionSeries.Transition presentation={wipe({ direction: "from-right" })} timing={linearTiming({ durationInFrames: transitionFrames })} />

        <TransitionSeries.Sequence durationInFrames={240} premountFor={60} name="07 · Vitals interaction">
          <Vitals />
        </TransitionSeries.Sequence>
        <TransitionSeries.Transition presentation={fade()} timing={linearTiming({ durationInFrames: transitionFrames })} />

        <TransitionSeries.Sequence durationInFrames={180} premountFor={60} name="08 · Privacy">
          <Privacy />
        </TransitionSeries.Sequence>
        <TransitionSeries.Transition presentation={wipe({ direction: "from-bottom" })} timing={linearTiming({ durationInFrames: transitionFrames })} />

        <TransitionSeries.Sequence durationInFrames={210} premountFor={60} name="09 · End card">
          <End />
        </TransitionSeries.Sequence>
      </TransitionSeries>

      <Audio
        src={staticFile("audio/iris-score-fast.mp3")}
        volume={(frame) => interpolate(frame, [0, 40, durationInFrames - 70, durationInFrames], [0, 0.72, 0.72, 0], {
          extrapolateLeft: "clamp",
          extrapolateRight: "clamp",
        })}
      />
      <SoundEffect from={112} src="whip.wav" volume={0.24} />
      <SoundEffect from={254} src="switch.wav" volume={0.2} />
      <SoundEffect from={396} src="mouse-click.wav" volume={0.26} />
      <SoundEffect from={628} src="whoosh.wav" volume={0.28} />
      <SoundEffect from={710} src="mouse-click.wav" volume={0.2} />
      <SoundEffect from={980} src="whip.wav" volume={0.25} />
      <SoundEffect from={1075} src="mouse-click.wav" volume={0.2} />
      <SoundEffect from={1302} src="whip.wav" volume={0.25} />
      <SoundEffect from={1534} src="whoosh.wav" volume={0.3} />
      <SoundEffect from={1706} src="ding.wav" volume={0.24} />
    </>
  );
};
