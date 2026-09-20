import type { Config } from 'tailwindcss';

export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        ink: '#17212b',
        market: '#0f766e',
        signal: '#f59e0b',
      },
    },
  },
  plugins: [],
} satisfies Config;

