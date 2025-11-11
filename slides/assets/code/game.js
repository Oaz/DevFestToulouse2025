
function playPhase(phaseId, getVisual) {
    const height = convertToPixels('90vh');
    const fillColor = getColorFromVariable('--util-dark-color');
    const emptyColor = getColorFromVariable('--util-light-color');
    return async function (state) {
        if (!state.forwarded) {
            const forward = await forwardGamePhase(phaseId);
            state.forwarded = !!forward;
        }
        let completion = await getPhaseCompletion(phaseId);
        if (!completion) {
            completion = {
                completed_count: 0,
                total_players: 1,
            }
        }
        const visual = getVisual(completion);
        const gauge = verticalGauge(
            visual.count, visual.total,
            visual.display, height,
            fillColor, emptyColor, "white", visual.textSize
        );
        if (visual.complete) {
            return gauge;
        }
        return {
            data: gauge.data,
            layout: gauge.layout,
            repeat: 1000,
            state: state
        };
    }
}

function startPart(phaseId) {
    return playPhase(phaseId, (completion) => {
        return {
            count: Math.log(completion.total_players || 1),
            total: Math.log(1000),
            display: completion.total_players.toString(),
            textSize: 100,
            complete: false,
        }
    });
}

function startPhase(phaseId) {
    return playPhase(phaseId, (completion) => {
        return {
            count: completion.completed_count,
            total: completion.total_players,
            display: completion.completed_count.toString(),
            textSize: 20,
            complete: completion.completed_count === completion.total_players,
        }
    });
}

const part1 = {
    phaseId: 8,
    range: [10, 30],
    binsize: 1,
    aiPath: 'assets/data/game1.json',
}
const part2 = {
    phaseId: 18,
    range: [30, 120],
    binsize: 2,
    aiPath: 'assets/data/game2.json',
}


function updateBestScore(data) {
    const bestScore = data.reduce((a, b) => a.Score < b.Score ? a : b).Score;
    const bestScoreElement = document.getElementById('best-score');
    if (bestScoreElement) {
        bestScoreElement.innerHTML = bestScore;
    }
    const winners = data.filter(a => a.Score === bestScore).map(a => a.Name.slice(0,20));
    const winnersElement = document.getElementById('winners');
    if (winnersElement) {
        winnersElement.innerHTML = winners.join(', ');
    }
}

async function getHumanResults(part) {
    if (!window.humanResults) {
        window.humanResults = {};
    }
    const phaseId = part.phaseId;
    window.humanResults[phaseId] = await getPhaseResults(phaseId);
    if(part.phaseId === 18) {
        updateBestScore(window.humanResults[phaseId]);
    }
    return window.humanResults[phaseId];
}

function showHumanScores(part) {
    const color = getColorFromVariable('--humans-foreground-color');
    return async function () {
        const results = await getHumanResults(part);
        return histogram(results, part, "Humains", color);
    }
}

function forwardAndShowHumanScores(part) {
    const phaseId = part.phaseId;
    return async function () {
        const forward = await forwardGamePhase(phaseId);
        return await showHumanScores(part)();
    }
}

async function getAiResults(part) {
    if (!window.aiResults) {
        window.aiResults = {};
    }
    const phaseId = part.phaseId;
    if (!window.aiResults.hasOwnProperty(phaseId)) {
        const response = await fetch(part.aiPath);
        window.aiResults[phaseId] = await response.json();
    }
    return window.aiResults[phaseId];
}

function showAiScores(part) {
    const color = getColorFromVariable('--ai-foreground-color');
    return async function () {
        const results = await getAiResults(part);
        return histogram(results, part, "AI", color);
    }
}

function countAssetSelection(results, asset, phaseId) {
    return results.filter(player => player['Assets'][asset] === phaseId).length;
}

function countChoices(choices, data) {
    const counts = choices.map(ways => ways
        .map(way => countAssetSelection(data, way[0], way[1]))
        .reduce((a, b) => a + b, 0)
    );
    const total = counts.reduce((a, b) => a + b, 0);
    return counts.map(count => Math.round(100 * count / total));
}

function compareChoices(part, choices) {
    return async function () {
        const humans = await getHumanResults(part);
        const ai = await getAiResults(part);
        const choicesDefinitions = choices.map(choice => choice[1]);
        return stackedBars(
            choices.map(choice => choice[0]),
            {
                "Humains": countChoices(choicesDefinitions, humans),
                "AI": countChoices(choicesDefinitions, ai),
            }
        );
    }
}


function createGroupedBars(choices, humanChoices, aiChoices) {
    return groupedBars(
        [
            {
                title: "Humains",
                textColor: 'white',
                fillColor: getColorFromVariable('--humans-foreground-color'),
            },
            {
                title: "AI",
                textColor: 'white',
                fillColor: getColorFromVariable('--ai-foreground-color'),
            },
        ],
        Object.fromEntries(
            Object.keys(choices).map(k => [
                choices[k][0],
                [humanChoices[k], aiChoices[k]]
            ])
        )
    );
}

function countTrendChoices(choices, data) {
    return choices.map(choice => Math.round(100 * countAssetSelection(data, choice[1], choice[2]) / data.length));
}

function compareTrends(part, choices) {
    return async function () {
        const humans = await getHumanResults(part);
        const humanChoices = countTrendChoices(choices, humans);
        const ai = await getAiResults(part);
        const aiChoices = countTrendChoices(choices, ai);
        return createGroupedBars(choices, humanChoices, aiChoices);
    }
}

function countAnyAssetSelection(results, phaseId) {
    return results.filter(player => Object.values(player['Assets']).includes(phaseId)).length;
}

function countRelativeTrendChoices(choices, data) {
    return choices.map(choice => {
        const phaseId = choice[2];
        const assetSelectionsInPhase = countAssetSelection(data, choice[1], phaseId);
        const anySelectionInPhase = countAnyAssetSelection(data, phaseId);
        return Math.round(100 * assetSelectionsInPhase / anySelectionInPhase );
    });
}

function compareRelativeTrends(part, choices) {
    return async function () {
        const humans = await getHumanResults(part);
        const humanChoices = countRelativeTrendChoices(choices, humans);
        const ai = await getAiResults(part);
        const aiChoices = countRelativeTrendChoices(choices, ai);
        return createGroupedBars(choices, humanChoices, aiChoices);
    }
}

