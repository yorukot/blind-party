/** @type {import('tailwindcss').Config} */
export default {
    content: ['./src/**/*.{html,js,svelte,ts}'],
    theme: {
        extend: {
            fontFamily: {
                minecraft: ['Minecraft', 'Courier New', 'Lucida Console', 'monospace'],
                pixel: ['Minecraft', 'Courier New', 'Lucida Console', 'monospace'],
                mono: ['Minecraft', 'Courier New', 'Lucida Console', 'monospace']
            }
        }
    },
    plugins: []
};
