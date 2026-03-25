/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        background: "#030712", // Very dark blue/black
        foreground: "#f9fafb",
        primary: {
          DEFAULT: "#3b82f6", // Bright blue
          foreground: "#ffffff",
        },
        secondary: {
          DEFAULT: "#1f2937",
          foreground: "#f9fafb",
        },
        accent: {
          DEFAULT: "#8b5cf6", // Purple/Violet
          foreground: "#ffffff",
        },
        muted: {
          DEFAULT: "#111827",
          foreground: "#9ca3af",
        },
        border: "rgba(255, 255, 255, 0.1)",
      },
      fontFamily: {
        sans: ["Inter", "sans-serif"],
        display: ["Cal Sans", "Inter", "sans-serif"],
        mono: ["JetBrains Mono", "monospace"],
      },
      backgroundImage: {
        "gradient-radial": "radial-gradient(var(--tw-gradient-stops))",
        "gradient-conic": "conic-gradient(from 180deg at 50% 50%, var(--tw-gradient-stops))",
      },
    },
  },
  plugins: [],
}
