/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./components/**/*.templ"],
  theme: {
    extend: {
      colors: {
        void: '#050505',
        panel: '#0a0a0a',
        border: '#1f1f1f',
        steel: '#2a2a2a',
        text: '#e5e5e5',
        dim: '#888888',
        orange: '#ff4d00',
      },
      fontFamily: {
        mono: ['JetBrains Mono', 'monospace'],
        sans: ['Inter', 'sans-serif'],
      },
      animation: {
        'typewriter': 'type 3s steps(40, end) 1s 1 normal both',
        'cursor': 'blink 1s step-end infinite',
      },
      keyframes: {
        type: {
          '0%': { width: '0' },
          '100%': { width: '100%' },
        },
        blink: {
          '0%, 100%': { borderColor: 'transparent' },
          '50%': { borderColor: '#ff4d00' }, // Orange cursor
        }
      }
    },
  },
  plugins: [],
}
