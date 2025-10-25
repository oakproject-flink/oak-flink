/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "../web/templates/**/*.templ",
    "../web/templates/**/*.html",
  ],
  theme: {
    extend: {},
  },
  plugins: [
    require('@tailwindcss/typography'),
    require('daisyui'),
  ],
  daisyui: {
    themes: [
      {
        oak: {
          "primary": "#10b981",      // emerald-500 (Oak green)
          "secondary": "#6366f1",    // indigo-500
          "accent": "#f59e0b",       // amber-500
          "neutral": "#1f2937",      // gray-800
          "base-100": "#0f172a",     // slate-900 (dark background)
          "base-200": "#1e293b",     // slate-800
          "base-300": "#334155",     // slate-700
          "info": "#3b82f6",         // blue-500
          "success": "#10b981",      // emerald-500
          "warning": "#f59e0b",      // amber-500
          "error": "#ef4444",        // red-500
        },
      },
      "dark",
      "light",
      "cupcake",
    ],
  },
}
