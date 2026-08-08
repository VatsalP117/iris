import { Composition, Folder } from "remotion";

import { IrisLaunch } from "./IrisLaunch";
import { Clarity } from "./scenes/Clarity";
import { End } from "./scenes/End";
import { Events } from "./scenes/Events";
import { Intro } from "./scenes/Intro";
import { Noise } from "./scenes/Noise";
import { Overview } from "./scenes/Overview";
import { Privacy } from "./scenes/Privacy";
import { Setup } from "./scenes/Setup";
import { Vitals } from "./scenes/Vitals";

const fps = 60;
const width = 1920;
const height = 1080;

export const MyComposition: React.FC = () => (
  <>
    <Composition id="IrisLaunch" component={IrisLaunch} durationInFrames={1916} fps={fps} width={width} height={height} />
    <Folder name="Iris-Launch-Scenes">
      <Composition id="SceneIntro" component={Intro} durationInFrames={120} fps={fps} width={width} height={height} />
      <Composition id="SceneNoise" component={Noise} durationInFrames={150} fps={fps} width={width} height={height} />
      <Composition id="SceneClarity" component={Clarity} durationInFrames={150} fps={fps} width={width} height={height} />
      <Composition id="SceneSetup" component={Setup} durationInFrames={240} fps={fps} width={width} height={height} />
      <Composition id="SceneOverview" component={Overview} durationInFrames={360} fps={fps} width={width} height={height} />
      <Composition id="SceneEvents" component={Events} durationInFrames={330} fps={fps} width={width} height={height} />
      <Composition id="SceneVitals" component={Vitals} durationInFrames={240} fps={fps} width={width} height={height} />
      <Composition id="ScenePrivacy" component={Privacy} durationInFrames={180} fps={fps} width={width} height={height} />
      <Composition id="SceneEnd" component={End} durationInFrames={210} fps={fps} width={width} height={height} />
    </Folder>
  </>
);
