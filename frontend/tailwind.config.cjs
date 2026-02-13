/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './index.html',
    './src/**/*.{js,ts,jsx,tsx}',
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ['"Plus Jakarta Sans"', 'sans-serif'],
      },
      colors: {
        accent: {
          primary: '#8b5cf6',
          secondary: '#3b82f6',
          vibrant: '#f43f5e',
        },
        bg: {
          base: '#060910',
          surface: '#0b111a',
          'surface-soft': '#0d1420',
        }
      },
      backgroundImage: {
        'vibrant-gradient': 'linear-gradient(135deg, #8b5cf6 0%, #3b82f6 100%)',
        'surface-gradient': 'linear-gradient(180deg, rgba(255, 255, 255, 0.05) 0%, rgba(255, 255, 255, 0) 100%)',
      },
      boxShadow: {
        'vibrant': '0 4px 20px -2px rgba(139, 92, 246, 0.3)',
        'inner-soft': 'inset 0 1px 1px 0 rgba(255, 255, 255, 0.05)',
      }
    },
  },
  plugins: [],
};
