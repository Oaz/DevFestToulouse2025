<script lang="ts">
    import { onMount } from 'svelte';

    let visible = false;

    onMount(() => {
        visible = true;
    });
</script>

<div class="phase-container">
        <h1 class="title" class:visible>Ça va commencer !</h1>
        <h2 class="subtitle" class:visible>Écoute bien les consignes...</h2>
</div>

<style>
    .phase-container {
        width: 100%;
        height: 100vh;
        display: flex;
        flex-direction: column;
        justify-content: center;
        align-items: center;
        text-align: center;
        padding: 1rem;
    }

    .title {
        font-size: 2rem;
        font-weight: 700;
        margin-bottom: 1.5rem;
        color: #2d3748;
        opacity: 0;
        transform: translateY(20px);
        transition: opacity 0.6s ease-out, transform 0.6s ease-out;
    }

    .subtitle {
        font-size: 1.3rem;
        font-weight: 500;
        color: #4a5568;
        opacity: 0;
        transform: translateY(20px);
        transition: opacity 0.6s ease-out 0.2s, transform 0.6s ease-out 0.2s;
    }

    .title.visible, .subtitle.visible {
        opacity: 1;
        transform: translateY(0);
    }

    /* Mobile optimization */
    @media (max-width: 480px) {
        .title {
            font-size: 1.8rem;
            margin-bottom: 1rem;
        }

        .subtitle {
            font-size: 1.1rem;
        }
    }
</style>