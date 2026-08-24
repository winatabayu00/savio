/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        brand: {
          DEFAULT: '#0e5b4e',
          light: '#127b6a',
          dark: '#0a453c',
        },
      },
    },
  },
  plugins: [],
};