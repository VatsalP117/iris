import { loadFont as loadPlexSans } from "@remotion/google-fonts/IBMPlexSans";
import { loadFont as loadPlexMono } from "@remotion/google-fonts/IBMPlexMono";
import { loadFont as loadInstrumentSerif } from "@remotion/google-fonts/InstrumentSerif";

export const { fontFamily: sans } = loadPlexSans("normal", {
  weights: ["400", "500", "600"],
  subsets: ["latin"],
});

export const { fontFamily: mono } = loadPlexMono("normal", {
  weights: ["400", "500"],
  subsets: ["latin"],
});

export const { fontFamily: serif } = loadInstrumentSerif("normal", {
  weights: ["400"],
  subsets: ["latin"],
});

export const colors = {
  paper: "#f2efe7",
  raised: "#faf8f2",
  ink: "#151512",
  muted: "#777168",
  rule: "#d6d0c4",
  violet: "#6558ef",
  violetDark: "#3328a8",
  violetSoft: "#e5e0ff",
  green: "#2f8a59",
  red: "#a44037",
  black: "#11110f",
  white: "#faf9f5",
};

export const ease = [0.16, 1, 0.3, 1] as const;
