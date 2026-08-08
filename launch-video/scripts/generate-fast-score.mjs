import { mkdir, writeFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import path from "node:path";

const scriptDirectory = path.dirname(fileURLToPath(import.meta.url));
const audioDirectory = path.resolve(scriptDirectory, "../public/audio");
const outputPath = path.join(audioDirectory, "iris-score-fast-raw.wav");

const sampleRate = 48000;
const duration = 40;
const channels = 2;
const bpm = 128;
const secondsPerBeat = 60 / bpm;
const totalSamples = sampleRate * duration;
const pcm = Buffer.alloc(totalSamples * channels * 2);
const bassNotes = [73.42, 87.31, 58.27, 65.41];
const arpRatios = [2, 2.5, 3, 4, 3, 2.5, 2, 3];
let randomState = 0x1a2b3c4d;

function noise() {
    randomState = (1664525 * randomState + 1013904223) >>> 0;
    return (randomState / 0xffffffff) * 2 - 1;
}

function envelope(localTime, decay) {
    return localTime < 0 ? 0 : Math.exp(-localTime * decay);
}

function softClip(value) {
    return Math.tanh(value * 1.15) * 0.82;
}

for (let index = 0; index < totalSamples; index += 1) {
    const time = index / sampleRate;
    const beat = time / secondsPerBeat;
    const beatIndex = Math.floor(beat);
    const beatPhaseSeconds = (beat - beatIndex) * secondsPerBeat;
    const halfBeat = beat * 2;
    const halfBeatIndex = Math.floor(halfBeat);
    const halfBeatPhaseSeconds = (halfBeat - halfBeatIndex) * (secondsPerBeat / 2);
    const bar = Math.floor(beat / 4);
    const beatInBar = beat - bar * 4;

    const kickEnvelope = envelope(beatPhaseSeconds, 22);
    const kickPhase = 2 * Math.PI * (47 * beatPhaseSeconds + 7.2 * (1 - Math.exp(-18 * beatPhaseSeconds)));
    const kick = Math.sin(kickPhase) * kickEnvelope * 0.92;

    const snareDistance = Math.min(Math.abs(beatInBar - 1), Math.abs(beatInBar - 3));
    const snareTime = snareDistance * secondsPerBeat;
    const snareEnvelope = snareTime < 0.16 ? envelope(snareTime, 24) : 0;
    const snare = (noise() * 0.62 + Math.sin(2 * Math.PI * 185 * snareTime) * 0.38) * snareEnvelope * 0.5;

    const hatEnvelope = envelope(halfBeatPhaseSeconds, halfBeatIndex % 2 === 0 ? 80 : 120);
    const hat = noise() * hatEnvelope * (halfBeatIndex % 2 === 0 ? 0.14 : 0.09);

    const root = bassNotes[bar % bassNotes.length];
    const bassEnvelope = 0.24 + envelope(beatPhaseSeconds, 5.5) * 0.76;
    const bass = (
        Math.sin(2 * Math.PI * root * time) * 0.7
        + Math.sin(2 * Math.PI * root * 2 * time) * 0.2
        + Math.sin(2 * Math.PI * root * 3 * time) * 0.1
    ) * bassEnvelope * 0.23;

    const arpFrequency = root * arpRatios[halfBeatIndex % arpRatios.length];
    const arpEnvelope = envelope(halfBeatPhaseSeconds, 11);
    const arp = (Math.sin(2 * Math.PI * arpFrequency * time) + Math.sin(2 * Math.PI * arpFrequency * 2 * time) * 0.2) * arpEnvelope * 0.09;

    const pad = (
        Math.sin(2 * Math.PI * root * 2 * time)
        + Math.sin(2 * Math.PI * root * 2.5 * time)
        + Math.sin(2 * Math.PI * root * 3 * time)
    ) * 0.025;

    const build = Math.min(1, time / 2.2) * Math.min(1, (duration - time) / 1.2);
    const left = softClip((kick + snare + bass + pad + hat * 0.8 + arp * 1.1) * build);
    const right = softClip((kick + snare + bass + pad + hat * 1.15 + arp * 0.8) * build);
    pcm.writeInt16LE(Math.round(left * 32767), index * 4);
    pcm.writeInt16LE(Math.round(right * 32767), index * 4 + 2);
}

const header = Buffer.alloc(44);
const dataSize = pcm.length;
header.write("RIFF", 0);
header.writeUInt32LE(36 + dataSize, 4);
header.write("WAVE", 8);
header.write("fmt ", 12);
header.writeUInt32LE(16, 16);
header.writeUInt16LE(1, 20);
header.writeUInt16LE(channels, 22);
header.writeUInt32LE(sampleRate, 24);
header.writeUInt32LE(sampleRate * channels * 2, 28);
header.writeUInt16LE(channels * 2, 32);
header.writeUInt16LE(16, 34);
header.write("data", 36);
header.writeUInt32LE(dataSize, 40);

await mkdir(audioDirectory, { recursive: true });
await writeFile(outputPath, Buffer.concat([header, pcm]));
console.log(`Generated ${bpm} BPM score at ${outputPath}`);
