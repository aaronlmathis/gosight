// tailwind.config.js

module.exports = {
    content: [
      './templates/**/*.html',
      './js/**/*.js'
    ],
    darkMode: 'class', // Enables toggling via 'dark' class
    theme: {
      extend: {
        colors: {
          brand: {
            light: '#f1f5f9',
            DEFAULT: '#3b82f6',
            dark: '#1e3a8a',
          },
        },
      },
    },
    plugins: [],
  };
  