
<script lang="ts">
    import type {Message} from "./message";

    export let message: Message | null = null;

    let isVisible = !!message;
    $: if (message) {
        isVisible = true;
        startAutoDismissTimer();
    }

    function dismiss() {
        isVisible = false;
    }

    function startAutoDismissTimer() {
        setTimeout(() => {
            isVisible = false;
        }, 5000);
    }
</script>

{#if isVisible && message}
    <div class="information" class:error={message.type === 'error'} class:info={message.type === 'info'}>
        <p>{message.text}</p>
        <button class="close-button" on:click={dismiss}>×</button>
    </div>
{/if}

<style>
    .information {
        position: fixed;
        top: 0;
        left: 0;
        right: 0;
        padding: 1rem;
        z-index: 1000;
        display: flex;
        justify-content: space-between;
        align-items: center;
    }

    .error {
        background-color: #f8d7da;
        color: #721c24;
    }

    .info {
        background-color: #d4edda;
        color: #155724;
    }

    p {
        margin: 0;
    }

    .close-button {
        background: none;
        border: none;
        font-size: 1.5rem;
        cursor: pointer;
    }

    .error .close-button {
        color: #721c24;
    }

    .info .close-button {
        color: #155724;
    }
</style>