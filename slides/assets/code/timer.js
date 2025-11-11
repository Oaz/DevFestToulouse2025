
function showTimer(seconds) {
    const fillColor = getColorFromVariable('--timer-dark-color');
    const emptyColor = getColorFromVariable('--timer-light-color');
    return function(state) {
        const max = seconds * 1000;
        const now = new Date();
        if (!state.start) {
            state.start = now;
        }
        let elapsed = now - state.start;
        if (elapsed > max) {
            elapsed = max;
        }
        const gauge = horizontalGauge(
            elapsed, max, Math.round((max-elapsed) / 1000).toString(),
            35, fillColor, emptyColor, "white", 20
        );
        return {
            data: gauge.data,
            layout: gauge.layout,
            repeat: 200,
            state: state
        };
    }
}



