/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        kawa: {
          50: '#f0fdf0',
          100: '#dcfce7',
          200: '#bbf7d0',
          300: '#86efac',
          400: '#4ade80',
          500: '#76FF03', // Kawasaki Lime
          600: '#65e600',
          700: '#4bc600',
          800: '#3ba300',
          900: '#2d8000',
          950: '#1a4d00',
        },
      },
    },
  },
  plugins: [],
}
