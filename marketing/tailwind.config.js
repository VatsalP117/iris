/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        background: "#0a0a0a",
        "bg-secondary": "#111111",
        foreground: "#e5e5e5",
        muted: {
          DEFAULT: "#737373",
          foreground: "#a3a3a3",
        },
        border: "rgba(255, 255, 255, 0.06)",
      },
      fontFamily: {
        sans: ["Inter", "system-ui", "sans-serif"],
        mono: ["JetBrains Mono", "monospace"],
      },
    },
  },
  plugins: [],
}
