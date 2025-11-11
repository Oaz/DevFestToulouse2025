/**
 * Move the game forward to the specified phase
 * @param {number} phaseId - The phase ID to advance to
 * @returns {Promise<Object>} Response from the server
 */
async function forwardGamePhase(phaseId) {
    try {
        const response = await fetch(`${API_URL}/forward`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                admin_password: ADMIN_PASSWORD,
                phase_id: phaseId
            })
        });

        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(`HTTP ${response.status}: ${JSON.stringify(errorData)}`);
        }

        return await response.json();
    } catch (error) {
        console.error('Error forwarding game phase:', error);
    }
}

/**
 * Get phase completion statistics
 * @param {number} phaseId - The phase ID to check completion for
 * @returns {Promise<{completed_count: number, total_players: number}>} Completion stats
 */
async function getPhaseCompletion(phaseId) {
    try {
        const response = await fetch(`${API_URL}/completion`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                phase_id: phaseId
            })
        });

        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(`HTTP ${response.status}: ${JSON.stringify(errorData)}`);
        }

        return await response.json();
    } catch (error) {
        console.error('Error getting phase completion:', error);
    }
}

/**
 * Get all players' results for the specified phase (admin only)
 * @param {number} phaseId - The phase ID to get results for
 * @returns {Promise<{phase_id: number, results: Array<{id: number, score: number, assets: Object}>}>} Results data
 */
async function getPhaseResults(phaseId) {
    try {
        const response = await fetch(`${API_URL}/results`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                admin_password: ADMIN_PASSWORD,
                phase_id: phaseId
            })
        });

        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(`HTTP ${response.status}: ${JSON.stringify(errorData)}`);
        }

        const json = await response.json();
        return json.results;
    } catch (error) {
        console.error('Error getting phase results:', error);
    }
}