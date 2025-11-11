<script lang="ts">
    import {gameState} from './state';
    import {FORM_URL, SOURCE_CODE_URL} from "./config";

    $: state = $gameState;
    $: url = `${FORM_URL}?player_id=${state.user.player_id}&secret=${state.user.secret}`
</script>

<div class="phase-container">
    <div class="content-card">
        <h2 class="title">Résultat</h2>

        <div class="scores">
            {#if state.scores.length > 0 }
                <div class="score-item">
                    <h3>Match aller</h3>
                    <div class="score-value">{state.scores[0] > 0 ? state.scores[0] : "score indisponible"}</div>
                </div>
            {/if}
            {#if state.scores.length > 1 }
                <div class="score-item">
                    <h3>Match retour</h3>
                    <div class="score-value">{state.scores[1] > 0 ? state.scores[1] : "score indisponible"}</div>
                </div>
            {/if}
        </div>

        {#if state.scores.length > 1 }
            <div class="links">
                <a href={url} target="_blank" class="button primary">Répondre au questionnaire <span class="subtitle">(juste 2 petites questions)</span></a>
                <a href="{SOURCE_CODE_URL}" target="_blank" class="button secondary">Voir le code source</a>
            </div>
        {/if}
    </div>
</div>

<style>
    .phase-container {
        width: 100%;
        min-height: 100vh;
        display: flex;
        justify-content: center;
        align-items: center;
        background-color: #f5f7fa;
        padding: 1rem;
        box-sizing: border-box;
        margin: 0;
    }

    .content-card {
        background-color: white;
        border-radius: 12px;
        box-shadow: 0 6px 24px rgba(0, 0, 0, 0.1);
        padding: 1.5rem;
        width: 100%;
        max-width: 100%;
        text-align: center;
        box-sizing: border-box;
    }

    .title {
        font-size: 1.8rem;
        margin: 0 0 1.2rem 0;
        color: #2a2f45;
        font-weight: 700;
    }

    .scores {
        display: flex;
        flex-direction: column;
        gap: 1rem;
        margin-bottom: 1.8rem;
    }

    .score-item {
        padding: 0.8rem;
        background-color: #f8f9fa;
        border-radius: 8px;
        width: 100%;
        box-sizing: border-box;
    }

    .score-item h3 {
        margin: 0 0 0.4rem 0;
        color: #5d6683;
        font-size: 1.1rem;
    }

    .score-value {
        font-size: 1.6rem;
        font-weight: 700;
        color: #2a2f45;
    }

    .links {
        display: flex;
        flex-direction: column;
        gap: 0.8rem;
        margin-top: 1rem;
        width: 100%;
    }

    .button {
        display: block;
        width: 100%;
        padding: 0.9rem 1rem;
        border-radius: 8px;
        font-weight: 600;
        text-decoration: none;
        transition: all 0.2s ease;
        box-sizing: border-box;
        text-align: center;
    }

    .button:hover {
        transform: translateY(-2px);
    }

    .primary {
        background-color: #5046e5;
        color: white;
    }

    .primary:hover {
        background-color: #3e35c7;
        box-shadow: 0 4px 12px rgba(80, 70, 229, 0.3);
    }

    .secondary {
        background-color: #e2e8f0;
        color: #3a4151;
    }

    .secondary:hover {
        background-color: #d1d9e6;
    }

    .subtitle {
        font-size: 0.8rem;
        font-weight: normal;
        opacity: 0.8;
        display: block;
        margin-top: 0.2rem;
    }

    @media (min-width: 600px) {
        .content-card {
            padding: 2rem;
            max-width: 550px;
        }

        .title {
            font-size: 2.2rem;
        }

        .subtitle {
            display: inline;
            margin-left: 0.3rem;
        }
    }
</style>