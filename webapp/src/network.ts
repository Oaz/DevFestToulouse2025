import {trace} from "./helpers";

function getApiUrl() : string {
    // Access the global GAME_API_URL that was defined in vite.config.js
    // @ts-ignore - TypeScript might not recognize this global variable
    return GAME_API_URL;
}

export async function apiCall(route: string, body: any) {
    try {
        trace('API call:', route, body);
        const response = await fetch(`${getApiUrl()}${route}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(body)
        });
        return response;
    } catch (error) {
        trace('API call failed:', error);
        throw error;
    }
}

export function connectWebSocket(onMessage: (data: any) => void): any {
    let ws: WebSocket | null = null;
    let reconnectAttempts = 0;
    const maxReconnectAttempts = 30;
    const reconnectDelay = 1000;

    function connect() {
        const GAME_WS_URL = getApiUrl().replace('http', 'ws') + '/ws';
        ws = new WebSocket(GAME_WS_URL);

        ws.onopen = () => {
            trace('WebSocket connected');
            reconnectAttempts = 0;
        };

        ws.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                onMessage(data);
            } catch (error) {
                trace('Error parsing WebSocket message:', error);
            }
        };

        ws.onclose = (event) => {
            trace(`WebSocket closed: ${event.code} ${event.reason}`);
            attemptReconnect();
        };

        ws.onerror = (error) => {
            trace('WebSocket error:', error);
        };

        // Setup ping interval to keep connection alive
        const pingInterval = setInterval(() => {
            if (ws && ws.readyState === WebSocket.OPEN) {
                ws.send(JSON.stringify({ type: 'ping' }));
            } else {
                clearInterval(pingInterval);
            }
        }, 30000); // Send ping every 30 seconds

        return {
            send: (data: any) => {
                if (ws && ws.readyState === WebSocket.OPEN) {
                    ws.send(JSON.stringify(data));
                }
            },
            close: () => {
                clearInterval(pingInterval);
                if (ws) {
                    ws.close();
                    ws = null;
                }
            }
        };
    }

    function attemptReconnect() {
        if (reconnectAttempts < maxReconnectAttempts) {
            reconnectAttempts++;
            trace(`Attempting to reconnect (${reconnectAttempts}/${maxReconnectAttempts})...`);
            setTimeout(connect, reconnectDelay * reconnectAttempts);
        } else {
            trace('Max reconnect attempts reached');
        }
    }

    // Handle device coming back online
    window.addEventListener('online', () => {
        if (!ws || ws.readyState !== WebSocket.OPEN) {
            connect();
        }
    });

    return connect();
}



