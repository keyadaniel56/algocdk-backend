/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./*.{html,js}",        // all HTML/JS in this folder
    "./**/*.{html,js}",     // all HTML/JS in subfolders
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}
