module.exports = {
  content: [
    './templates/**/*.html',
    './js/**/*.js'
  ],
  safelist: [
    'bg-green-100', 'text-green-800',
    'bg-yellow-100', 'text-yellow-800',
    'bg-red-100', 'text-red-800',
    'stat-card',

    'max-h-[1000px]',
    'opacity-0',
    'opacity-100',
    'rotate-90',
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
      fontFamily: {

        'sans': [
          'Inter', 'ui-sans-serif', 'system-ui', '-apple-system', 'system-ui', 'Segoe UI',
          'Roboto', 'Helvetica Neue', 'Arial', 'Noto Sans', 'sans-serif',
          'Apple Color Emoji', 'Segoe UI Emoji', 'Segoe UI Symbol', 'Noto Color Emoji'
        ],
      },
    },
  },
  plugins: [],
};
