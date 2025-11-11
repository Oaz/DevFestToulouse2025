<script lang="ts">
    import Identification from './Identification.svelte';
    import Game from './Game.svelte';
    import Information from './Information.svelte';
    import {gameState} from './state';
    import type {Message} from "./message";
    import {FEEDBACK_URL} from "./config";

    let message: Message | null = null;
    $: state = $gameState;

    const urlParams = new URLSearchParams(window.location.search);
    if(urlParams.has('reset')) {
        localStorage.clear();
        window.location.assign(FEEDBACK_URL);
    }

    function handleInformation(event: CustomEvent<Message>) {
        message = null;
        setTimeout(() => {
            message = event.detail;
        }, 0);
    }
</script>

<main>
    {#if message}
        <Information message={message} />
    {/if}

    {#if state.isIdentified}
        <Game on:information={handleInformation}/>
    {:else}
        <Identification on:information={handleInformation}/>
    {/if}
</main>

<style>
    main {
        padding: 0;
        margin: 0;
        width: 100%;
        height: 100%;
        font-family: Arial, sans-serif;
        position: relative;
    }
</style>