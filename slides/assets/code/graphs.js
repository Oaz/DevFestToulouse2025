
function getColorFromVariable(colorVariable) {
    return getComputedStyle(document.documentElement).getPropertyValue(colorVariable).trim();
}

function convertToPixels(value, element = document.body) {
    if (typeof value === 'number') {
        return value;
    }
    const tempElement = document.createElement('div');
    tempElement.style.position = 'absolute';
    tempElement.style.visibility = 'hidden';
    tempElement.style.height = value;
    element.appendChild(tempElement);
    const pixels = tempElement.getBoundingClientRect().height;
    element.removeChild(tempElement);
    return pixels;
}

const transparentLayout = {
    plot_bgcolor: 'rgba(0,0,0,0)',
    paper_bgcolor: 'rgba(0,0,0,0)',
};

const coreLayout = {
    ...transparentLayout,
    showlegend: false,
    margin: {
        l: 0,
        r: 0,
        t: 0,
        b: 0
    },
    xaxis: {
        visible: false,
    },
    yaxis: {
        visible: false,
    },
};

const gaugeLayout = {
    ...coreLayout,
    barmode: 'stack',
};

function gaugePercentage(current, total) {
    const currentValue = Math.max(0, Number(current) || 0);
    const totalValue = Math.max(1, Number(total) || 1); // Prevent division by zero
    const percentage = (currentValue / totalValue) * 100;
    return Math.max(0, Math.min(100, percentage));
}

function gaugeText(text, color, size=undefined) {
    return {
        text: [text],
        textposition: 'inside',
        textfont: {
            color: color,
            size: size,
        }
    }
}

function verticalBar(value, color) {
    return {
        x: [''],
        y: [value],
        type: 'bar',
        marker: {
            color: color,
        }
    }
}

function verticalGauge(
    current, total, text, height,
    fillColor = 'blue', emptyColor = 'orange',
    textColor = 'white', textSize = undefined
) {
    const percentage = gaugePercentage(current, total);
    const data = [
        {
            ...verticalBar(percentage, fillColor),
            ...gaugeText(text, textColor, textSize)
        },
        verticalBar(100 - percentage, emptyColor),
    ];
    const layout = {
        ...gaugeLayout,
        height: convertToPixels(height),
    };
    return {data, layout};
}

function horizontalBar(value, color) {
    return {
        y: [''],
        x: [value],
        type: 'bar',
        orientation: 'h',
        marker: {
            color: color,
        }
    }
}

function horizontalGauge(
    current, total, text, height= 35,
    fillColor = 'blue', emptyColor = 'orange',
    textColor = 'white', textSize = undefined
) {
    const percentage = gaugePercentage(current, total);
    const data = [
        {
            ...horizontalBar(percentage, fillColor),
            ...gaugeText(text, textColor, textSize)
        },
        horizontalBar(100 - percentage, emptyColor),
    ];
    const layout = {
        ...gaugeLayout,
        height: convertToPixels(height),
    };
    return {data, layout};
}


function histogram(data, part, title, color) {
    const scores = data.map(item => item.Score);
    return {
        data: [{
            x: scores,
            type: 'histogram',
            histnorm: 'percent',
            xbins: {
                size: part.binsize,
                start: part.range[0],
                end: part.range[1]
            },
            marker: {
                color: color
            }
        }],
        layout: {
            ...transparentLayout,
            width: 1000,
            height: 350,
            margin: {
                l: 100,
                r: 20,
                t: 0,
                b: 50
            },
            autosize: false,
            showlegend: false,
            xaxis: {
                range: part.range
            },
            yaxis: {
                title: {
                    text: `${title} (%)`
                }
            }
        }
    };
}

function stackedBarTrace(title, values) {
    return {
        y: Object.keys(values).map(v => `${v}  `),
        x: Object.values(values),
        name: title,
        type: 'bar',
        orientation: 'h',
        text: Object.values(values).map(v => `${v}%`),
        textposition: 'inside',
        insidetextanchor: 'middle',
    };
}

function stackedBars(cases, values) {
    const extractValues = n => Object.fromEntries(Object.entries(values).map(([k, v]) => [k, v[n]]));
    const data = Object.keys(cases).map(n => stackedBarTrace(cases[n], extractValues(n)));
    const layout = {
        ...transparentLayout,
        barmode: 'stack',
        height: 600,
        width: 1600,
        legend: {
            x: 0,
            y: 1 + 0.15 * Object.keys(cases).length,
            font: {
                size: 36
            }
        },
        yaxis: {
            visible: true,
            tickfont: {
                size: 32
            },
            automargin: true
        },
        xaxis: {
            visible: false,
        },
    };
    return {data, layout};
}


function groupedBarsTrace(trace, values) {
    return {
        x: Object.keys(values),
        y: Object.values(values),
        name: trace.title,
        type: 'bar',
        text: Object.values(values).map(v => `${v}%`),
        textposition: 'inside',
        insidetextanchor: 'middle',
        marker: {
            color: trace.fillColor,
        },
        textfont: {
            color: trace.textColor,
        },
    };
}

function groupedBars(cases, values) {
    const extractValues = n => Object.fromEntries(Object.entries(values).map(([k, v]) => [k, v[n]]));
    const data = Object.keys(cases).map(n => groupedBarsTrace(cases[n], extractValues(n)));

    const layout = {
        ...transparentLayout,
        barmode: 'group',
        height: 600,
        width: 1400,
        legend: {
            x: 0,
            y: 1.2,
            orientation: 'h',
            font: {
                size: 36
            }
        },
        xaxis: {
            visible: true,
            tickfont: {
                size: 32
            },
            automargin: true
        },
        yaxis: {
            visible: false,
        },
    };
    return {data, layout};
}
