// @ts-ignore
import rules from 'virtual:rules';

interface Acceptable {
    set: Array<{ id: string }>;
}

export interface Asset {
    id: string;
    text: string;
    cost: number;
}

export type PhaseType = 'begin' | 'single' | 'multi' | 'end';

export interface Phase {
    id: number;
    type: PhaseType;
    available: Set<string>;
    obsolete: Set<string>;
    accepted: Set<Set<string>>;
}

function MakeAssets(): Map<string, Asset> {
    let assets = new Map<string, Asset>();
    for (const assetData of rules.assets) {
        assets.set(assetData.id, assetData);
    }
    return assets;
}

export const assets = MakeAssets();

export function MakePhases(): Map<number, Phase> {
    let phases = new Map<number, Phase>();
    for (const phaseData of rules.phases) {
        const available = phaseData.available.map((a: { id: any; }) => a.id).filter((a: undefined) => a !== undefined);
        const obsolete = phaseData.obsolete.map((a: { id: any; }) => a.id).filter((a: undefined) => a !== undefined);
        const accepted = new Set<Set<string>>(
            phaseData.accepted.map((acc: { set: any[]; }) =>
                new Set<string>(acc.set.map(a => a.id))
            )
        );
        phases.set(phaseData.id, {
            id: phaseData.id,
            type: phaseData.type,
            available,
            obsolete,
            accepted,
        });
    }
    return phases;
}

export const phases = MakePhases();

export function GetAllAssetIDsUntil(phaseId: number): Set<string> {
    let assetIDs = new Set<string>();
    const sortedPhaseIds = Array.from(phases.keys()).sort((a, b) => a - b);
    for (const id of sortedPhaseIds) {
        if (id > phaseId) {
            break;
        }
        const phase = phases.get(id)!;
        if (phase.type === 'begin') {
            assetIDs = new Set<string>();
        }
        phase.available.forEach(assetID => assetIDs.add(assetID));
        phase.obsolete.forEach(assetID => assetIDs.delete(assetID));
    }
    return assetIDs;
}
