# Iris launch film

A fast, 32-second Remotion launch film for Iris v1.0.0. The composition uses real recorded dashboard interactions backed by deterministic mock analytics data, an original 128-BPM electronic score, and locally stored transition effects.

## Commands

```bash
# Refresh dashboard captures (dashboard must be running on :5174)
pnpm --filter launch-video capture

# Record real dashboard interactions
pnpm --filter launch-video record

# Regenerate the original electronic score
pnpm --filter launch-video score

# Open Remotion Studio
pnpm --filter launch-video dev

# Type-check and lint
pnpm --filter launch-video lint

# Render the final 1080p60 film
pnpm --filter launch-video render
```

The final render is written to `launch-video/out/iris-v1-launch.mp4`.

## Composition

The `IrisLaunch` composition is 1920×1080 at 60 fps. Every scene is also registered separately under `Iris-Launch-Scenes` for focused timing and visual edits in Remotion Studio.
