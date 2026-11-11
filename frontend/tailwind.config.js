/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js}'],
  theme: {
    extend: {
      colors: {
        gmail: {
          blue: '#1a73e8',
          blueDark: '#1967d2',
          text: '#202124',
          sub: '#5f6368',
          light: '#80868b',
          border: '#e0e3e7',
          hover: '#f1f3f4',
          list: '#f1f3f4',
          rowhover: '#e8eaed',
          active: '#d3e3fd',
          selected: '#c2e6ff',
          star: '#fbbc04',
        },
      },
      fontFamily: {
        sans: ['Roboto', '-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'Arial', 'sans-serif'],
      },
    },
  },
  plugins: [],
}
