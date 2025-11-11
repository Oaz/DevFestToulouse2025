<script lang="ts">
    import AssetButton from './AssetButton.svelte';
    import {assets, type Asset} from "./game";
    import ConfirmationButton from "./ConfirmationButton.svelte";
    import {gameState, updateGameState} from './state';

    $: state = $gameState;

    function handleSelectionChange(id: string, selected: boolean) {
        let selectedAssets = state.selectedAssets;
        if (selected) {
            selectedAssets.add(id);
        } else {
            selectedAssets.delete(id);
        }
        updateGameState({selectedAssets});
    }

    function find(assetId: string): Asset {
        return assets.get(assetId)!;
    }
</script>

<div class="phase-container">
    {#each state.ownedAssets as assetId}
        <AssetButton  asset={find(assetId)} owned={true}  />
    {/each}
    {#each state.potentialAssets as assetId}
        <AssetButton
                asset={find(assetId)}
                owned={false}
                selected={state.selectedAssets.has(assetId)}
                on:select={(event) => handleSelectionChange(assetId, event.detail)}
        />
    {/each}
    <ConfirmationButton />
</div>

<style>
    .phase-container {
        width: 100%;
        padding: 70px 0;
        display: flex;
        flex-direction: column;
        justify-content: flex-start;
        align-items: center;
        text-align: center;
    }

</style>