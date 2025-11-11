import { writable } from 'svelte/store';
import {phases, GetAllAssetIDsUntil} from './game';

import {type PhaseType} from './game';
import {createEventDispatcher} from "svelte";
import type {Message} from "./message";

export interface User {
    username: string;
    player_id: number;
    secret: string;
}

export interface State {
    user: User
    isIdentified: boolean;
    scores: number[];
    phaseId: number;
    phaseType: PhaseType;
    ending: boolean;
    ownedAssets: Set<string>;
    potentialAssets : Set<string>;
    selectedAssets : Set<string>;
    shouldConfirm : boolean;
    canConfirm : boolean;
}

const defaultState: State = {
    user: {
        username: "",
        player_id: -1,
        secret: ""
    },
    isIdentified: false,
    scores: [],
    phaseId: 0,
    phaseType: "begin",
    ending: false,
    ownedAssets: new Set(),
    potentialAssets : new Set(),
    selectedAssets : new Set(),
    shouldConfirm : true,
    canConfirm : false,
};

export const gameState = writable<State>(defaultState);

export function updateGameState(data: Partial<State>) {
    gameState.update(state => {
        let newState = {
            ...state,
            ...data,
        };
        let ownedAssets = newState.ownedAssets;
        const phase = phases.get(newState.phaseId)!;
        if (phase.type === 'begin') {
            ownedAssets = new Set();
        }
        let potentialAssets = GetAllAssetIDsUntil(newState.phaseId);
        ownedAssets.forEach(a => potentialAssets.delete(a));
        let shouldConfirm = true;
        let canConfirm = false;
        for (const set of phase.accepted) {
            if ([...set].every(item => ownedAssets.has(item)))
                shouldConfirm = false;
            if ([...set].every(item => ownedAssets.has(item) || newState.selectedAssets.has(item)))
                canConfirm = true;
        }
        return {
            ...newState,
            phaseType: phase.type,
            ending: state.phaseId != newState.phaseId && phase.type === 'end',
            ownedAssets,
            potentialAssets,
            shouldConfirm,
            canConfirm,
        };
    });
}


