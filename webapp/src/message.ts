
export type MessageType = 'error' | 'info';

export interface Message {
    text: string;
    type: MessageType;
}

import { get } from 'svelte/store';
import { gameState } from './state';

export function handlePhaseChange(dispatcher: (message: Message) => void): void {
    const state = get(gameState);
    if(state.shouldConfirm) return;
    dispatcher({
        text: "Le code existant suffit à satisfaire le besoin actuel !",
        type: "info"
    });
}