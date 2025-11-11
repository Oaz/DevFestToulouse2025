<script lang="ts">
    import {apiCall} from "./network";
    import {gameState, type State, updateGameState} from './state';
    import {trace} from "./helpers";

    let state: State;
    let isLoading = false;
    let isSuccess = false;
    let isLocked = false;
    let isVisible = false;

    $:{
        state = $gameState;
        isLocked = !state.canConfirm;
        isVisible = state.shouldConfirm;
    }

    function handleConfirm() {
        trace('Confirming selection');
        let ownedAssets = state.ownedAssets;
        state.selectedAssets.forEach(assetId => {
            ownedAssets.add(assetId);
        })
        updateGameState({
            ownedAssets,
            selectedAssets: new Set(),
        });
    }

    async function handleClick() {
        if (isLoading || isLocked) return;

        isLoading = true;

        try {
            const response = await apiCall('/select', {
                player_id: state.user.player_id,
                secret: state.user.secret,
                phase_id: state.phaseId,
                asset_ids: Array.from(state.selectedAssets)
            });
            if (response.ok) {
                isLoading = false;
                isSuccess = true;
                handleConfirm();
                setTimeout(() => {
                    isVisible = false;
                    // Reset states when button disappears
                    isSuccess = false;
                }, 1000);
            } else {
                // If API call fails, reset to initial state
                isLoading = false;
            }
        } catch (error) {
            // Handle any errors by resetting to initial state
            isLoading = false;
        }
    }

    function handleKeyDown(event: KeyboardEvent) {
        if (!isLoading && !isLocked && (event.key === 'Enter' || event.key === ' ')) {
            event.preventDefault();
            handleClick();
        }
    }
</script>

{#if isVisible}
    <div
            class="button {isLoading ? 'loading' : ''} {isSuccess ? 'success' : ''} {isLocked ? 'locked' : ''}"
            on:click={handleClick}
            on:keydown={handleKeyDown}
            role="button"
            tabindex={isLocked ? -1 : 0}
            aria-disabled={isLocked || isLoading}
    >
        <div class="text">
            {#if isLoading}
                Envoi en cours...
            {:else if isSuccess}
                La sélection est validée !
            {:else if isLocked}
                Il manque du code !
            {:else}
                Je valide la sélection
            {/if}
        </div>
    </div>
{/if}

<style>
    .button {
        display: flex;
        align-items: center;
        /*width: 100%;*/
        padding: 10px 100px;
        margin-bottom: 8px;
        border-radius: 8px;
        box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
        background-color: #bb9edb;
        transition: all 0.3s ease;
        cursor: pointer;
    }

    .button.loading {
        background-color: orange;
        cursor: wait;
        opacity: 0.8;
    }

    .button.success {
        background-color: #90ee90;
    }

    .button.locked {
        background-color: #bb9edb;
        cursor: not-allowed;
        opacity: 0.4;
    }

    .text {
        flex: 1;
        text-align: center;
    }

</style>