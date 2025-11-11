<script lang="ts">
    import {createEventDispatcher, onDestroy, onMount} from "svelte";
    import {apiCall, connectWebSocket} from "./network";
    import {assets, phases} from './game';
    import {gameState, updateGameState} from './state';
    import PhaseBegin from "./PhaseBegin.svelte";
    import PhaseEnd from "./PhaseEnd.svelte";
    import PhaseSelect from "./PhaseSelect.svelte";
    import {handlePhaseChange, type Message} from "./message";
    import {trace} from "./helpers";

    const dispatch = createEventDispatcher<{
        information: Message;
    }>();

    $: state = $gameState;

    async function fetchStatus() {
        try {
            const response = await apiCall('/status', {
                player_id: state.user.player_id,
                secret: state.user.secret,
            });
            if (response.ok) {
                const data = await response.json();
                updateGameState({
                    scores: data.scores,
                });
            }
        } catch (error) {
            console.error("Error fetching status:", error);
        }
    }

    $: if(state.ending) {
        fetchStatus();
    }

    let ws: any = null;

    onMount(() => {
        ws = connectWebSocket((data: any) => {
            trace('Received data from WebSocket:', data);
            updateGameState({
                phaseId: data.phase_id,
                selectedAssets: new Set(),
            });
            handlePhaseChange((message) => dispatch('information', message));
        });
        trace('Game assets:', assets);
        trace('Game phases:', phases);
    });

    onDestroy(() => {
        if (ws && ws.close) {
            trace('Closing WebSocket connection');
            ws.close();
        }
    });

</script>

<div class="game-container">
    {#if state.phaseType === 'begin'}
        <PhaseBegin />
    {:else if state.phaseType === 'end'}
        <PhaseEnd />
    {:else}
        <PhaseSelect />
    {/if}
</div>
