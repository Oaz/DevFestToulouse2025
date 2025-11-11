<script lang="ts">
    import {createEventDispatcher, onMount} from 'svelte';
    import {apiCall} from "./network";
    import {retryUntilSuccess} from "./helpers";
    import {updateGameState} from './state';
    import {handlePhaseChange, type Message} from "./message";


    const dispatch = createEventDispatcher<{
        information: Message;
    }>();

    function markIdentificationAsComplete(
        username: string, player_id: number, secret: string, phase_id: number,
        assets: string[], scores: number[]
    ) {
        updateGameState({
            user: {
                username,
                player_id,
                secret
            },
            phaseId: phase_id,
            ownedAssets: new Set(assets),
            scores,
            isIdentified: true,
        });
        handlePhaseChange((message) => dispatch('information', message));
    }

    function notifyError(text: string) {
        dispatch('information', {
            text,
            type: "error",
        });
    }


    let username = '';
    let isLoading = false;
    let showRegistrationForm = false;

    onMount(async () => {
        isLoading = true;

        const storedUsername = localStorage.getItem('username');
        const storedPlayerId = localStorage.getItem('player_id');
        const storedSecret = localStorage.getItem('secret');

        if (storedUsername && storedPlayerId && storedSecret) {
            try {
                await retryUntilSuccess(
                    () => verifyExistingPlayer(storedUsername, parseInt(storedPlayerId), storedSecret),
                    100,
                    2000
                );
            } catch (error) {
                showRegistrationForm = true;
            }
        } else {
            showRegistrationForm = true;
        }
        isLoading = false;
    });

    async function verifyExistingPlayer(username: string, playerId: number, secret: string): Promise<void> {
        try {
            isLoading = true;
            const response = await apiCall('/status', {
                player_id: playerId,
                secret
            });

            if (!response.ok) {
                throw new Error('Failed to verify player');
            }
            const data = await response.json();
            dispatch('information', {
                text:`Bonjour ${username} !`,
                type: "info",
            });
            markIdentificationAsComplete(username, playerId, secret, data.current_phase, data.selected_assets, data.scores);
        } catch (error) {
            throw error;
        }
    }

    async function registerPlayer() {
        if (!username.trim()) {
            notifyError("Inventez un nom si vous n'en avez pas !");
            return;
        }
        isLoading = true;
        try {
            const response = await apiCall('/register', {
                username
            });
            if (!response.ok) {
                throw new Error('Registration failed');
            }

            const data = await response.json();
            localStorage.setItem('username', username);
            localStorage.setItem('player_id', data.player_id);
            localStorage.setItem('secret', data.secret);

            markIdentificationAsComplete(username, data.player_id, data.secret, data.phase_id, [], []);
        } catch (error) {
            console.error('Registration error:', error);
            notifyError('Failed to register. Please try again.');
        } finally {
            isLoading = false;
        }
    }

    function handleSubmit() {
        registerPlayer();
    }

    function handleUsernameInput(event: Event) {
        const input = event.target as HTMLInputElement;
        const value = input.value;
        let filteredValue = value.replace(/[^a-zA-Z0-9\s-]/g, '');
        if (filteredValue.length > 20) {
            filteredValue = filteredValue.substring(0, 20);
        }
        if (value !== filteredValue) {
            input.value = filteredValue;
            username = filteredValue;
        }
    }
</script>

{#if isLoading}
    <div class="loading">
        <p>Chargement en cours...</p>
    </div>
{:else if showRegistrationForm}
    <div class="registration-form">
        <h2>Bienvenue au DevFest Toulouse 2025</h2>
        <h3>Humains vs AI : le match</h3>
        <form on:submit|preventDefault={handleSubmit}>
            <div class="form-group">
                <label for="username">Nom</label>
                <input
                        type="text"
                        id="username"
                        bind:value={username}
                        on:input={handleUsernameInput}
                        placeholder="Entrez votre nom"
                        required
                />
            </div>
            <button type="submit" class="register-button">On y va</button>
        </form>
    </div>
{/if}

<style>
    .loading {
        display: flex;
        justify-content: center;
        align-items: center;
        height: 100vh;
        font-size: 1.2rem;
    }

    .registration-form {
        display: flex;
        flex-direction: column;
        justify-content: center;
        min-height: 100vh;
        max-width: 300px;
        margin: 0 auto;
        padding: 2rem 1rem;
    }

    h2 {
        text-align: center;
        margin-bottom: 1.5rem;
    }

    h3 {
        text-align: center;
        margin-bottom: 1.5rem;
    }

    .form-group {
        margin-bottom: 1rem;
    }

    label {
        display: block;
        margin-bottom: 0.5rem;
        font-weight: bold;
    }

    input {
        width: 100%;
        padding: 0.75rem;
        border: 1px solid #ccc;
        border-radius: 4px;
        font-size: 1rem;
    }

    .register-button {
        width: 100%;
        padding: 0.75rem;
        background-color: #4a86e8;
        color: white;
        border: none;
        border-radius: 4px;
        font-size: 1rem;
        cursor: pointer;
        margin-top: 1rem;
    }

    .register-button:hover {
        background-color: #3b76d8;
    }
</style>