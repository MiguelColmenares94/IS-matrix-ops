/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js}'],
  theme: {
    extend: {
      colors: {
        brand: {
          pink: '#EA0C90',
          blue: '#0855C4',
          lavender: '#C8D1F1',
          navy: '#454A6C',
          gray: '#F2F2F2',
        },
      },
    },
  },
  plugins: [],
}
