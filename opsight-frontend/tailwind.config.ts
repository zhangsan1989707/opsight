import type { Config } from 'tailwindcss';

const config: Config = {
  darkMode: 'class',
  content: ['./app/**/*.{ts,tsx}', './components/**/*.{ts,tsx}'],
  theme: {
    extend: {
      fontFamily: {
        sans: ['"Space Grotesk"', 'system-ui', 'sans-serif'],
        mono: ['"JetBrains Mono"', '"SF Mono"', 'monospace'],
      },
      colors: {
        surface: {
          DEFAULT: '#09090b',
          50: '#111114',
          100: '#18181b',
          200: '#1f1f23',
          300: '#27272a',
        },
        accent: {
          DEFAULT: '#0ea5e9',
          dim: '#0369a1',
        },
        warn: '#f59e0b',
        danger: '#ef4444',
        success: '#10b981',
      },
    },
  },
  plugins: [],
};

export default config;
