/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["../views/**/*.html", "tw.css"],
  safelist: [
    "badge-neutral",
    "badge-primary",
    "badge-accent",
    "badge-success",
    "text-primary",
    "text-secondary",
  ],
  theme: {
    extend: {},
    fontFamily: {
      'mono': ['"Berkeley Mono"'],
    }
  },
  plugins: [require("daisyui")],
  daisyui: {
    themes: ["dracula"],
  },

}

