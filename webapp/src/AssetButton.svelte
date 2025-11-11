<script lang="ts">
    import type {Asset} from "./game";
    import { createEventDispatcher } from 'svelte';
    const dispatch = createEventDispatcher<{
        select: boolean;
    }>();

    export let asset: Asset;
    export let owned: boolean = false;
    export let selected: boolean = false;

    function handleClick() {
        if (!owned) {
            dispatch('select', !selected);
        }
    }

    function handleKeyDown(event: KeyboardEvent) {
        if (!owned && (event.key === 'Enter' || event.key === ' ')) {
            event.preventDefault();
            dispatch('select', !selected);
        }
    }
</script>

<div
        class="asset {owned ? 'owned' : 'potential'} {selected ? 'selected' : ''}"
        on:click={handleClick}
        on:keydown={handleKeyDown}
        role="button"
        tabindex={owned ? -1 : 0}
        aria-pressed={selected}
>
    <div class="cost">{asset.cost}</div>
    <div class="text">{asset.text}</div>
</div>

<style>
    .asset {
        display: flex;
        align-items: center;
        width: 100%;
        padding: 4px;
        margin-bottom: 8px;
        border-radius: 8px;
        box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    }

    .owned {
        background-color: #90f2f1;
        color: #004d40;
    }

    .potential {
        background-color: #e5e5e5;
        color: #212121;
        cursor: pointer;
    }

    .selected {
        background-color: #9bbedb;
        border: 2px solid #1976d2;
    }

    .cost {
        background-color: rgba(0, 0, 0, 0.1);
        border-radius: 4px;
        padding: 4px 8px;
        margin-right: 16px;
        font-weight: bold;
        min-width: 48px;
        text-align: center;
    }

    .text {
        flex: 1;
        text-align: center;
    }

    /* Responsive layout for larger screens */
    @media (min-width: 800px) {
        .asset {
            max-width: 750px;
            margin-left: auto;
            margin-right: auto;
        }
    }
</style>