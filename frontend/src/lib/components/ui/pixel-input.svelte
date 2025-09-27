<script lang="ts">
    interface Props {
        value?: string;
        placeholder?: string;
        maxlength?: number;
    }

    let { value = $bindable(''), placeholder = '', maxlength }: Props = $props();
</script>

<input
    bind:value
    {placeholder}
    {maxlength}
    class="
		pixel-input
		font-minecraft
		pixel-text
		w-full
		border-4
		border-gray-600
		bg-gray-800
		px-4
		py-3
		text-center
		text-xl
		font-bold
		tracking-wider
		text-white
		uppercase
		transition-colors
		duration-200
		placeholder:font-bold
		placeholder:text-gray-500
		focus:border-cyan-400
		focus:outline-none
	"
    style="
		box-shadow:
			inset -4px -4px 0px rgba(0,0,0,0.3),
			inset 4px 4px 0px rgba(255,255,255,0.1),
			4px 4px 0px rgba(0,0,0,0.4);
	"
/>

<style>
    .pixel-input {
        text-rendering: optimizeSpeed;
        -webkit-font-smoothing: none;
        -moz-osx-font-smoothing: grayscale;
    }

    .pixel-input:focus {
        animation: pixel-glow 1s ease-in-out infinite alternate;
    }

    @keyframes pixel-glow {
        from {
            box-shadow:
                inset -4px -4px 0px rgba(0, 0, 0, 0.3),
                inset 4px 4px 0px rgba(255, 255, 255, 0.1),
                4px 4px 0px rgba(0, 0, 0, 0.4),
                0 0 0 0 rgba(6, 182, 212, 0.7);
        }
        to {
            box-shadow:
                inset -4px -4px 0px rgba(0, 0, 0, 0.3),
                inset 4px 4px 0px rgba(255, 255, 255, 0.1),
                4px 4px 0px rgba(0, 0, 0, 0.4),
                0 0 0 4px rgba(6, 182, 212, 0.3);
        }
    }
</style>
