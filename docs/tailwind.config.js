/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./content/**/*.{html,js,vue,ts,md}",
    "./content/.vitepress/**/*.{html,js,vue,ts,md}",
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ["Outfit", "sans-serif"],
      },
      margin: {
        screen: "calc((-100vw + 100%)/2)",
      },
    },
  },
  plugins: [],
};
