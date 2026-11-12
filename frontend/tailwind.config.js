/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js}'],
  theme: {
    extend: {
      colors: {
        // opencode-inspired: warm black, hairline borders, burnt-orange accent
        oc: {
          bg: '#0a0a0b',
          panel: '#0f0f11',
          surface: '#151518',
          hover: '#1b1b1f',
          active: '#242429',
          border: '#232327',
          border2: '#34343a',
          text: '#ececee',
          sub: '#98989f',
          faint: '#55555d',
          accent: '#ff5d00',
          accent2: '#ff7a30',
          accentDim: '#1d1008',
          accentLine: '#3d2408',
          star: '#f5b642',
          error: '#e5484d',
          ok: '#46a758',
        },
      },
      fontFamily: {
        sans: [
          '"JetBrains Mono"',
          'ui-monospace',
          'SFMono-Regular',
          'Menlo',
          'Consolas',
          'monospace',
        ],
      },
      boxShadow: {
        dialog: '0 16px 48px rgba(0, 0, 0, 0.55)',
      },
    },
  },
  plugins: [],
}
