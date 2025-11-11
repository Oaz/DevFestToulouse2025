export function trace(...args: any[]) {
    console.log(...args);
}

export async function retryUntilSuccess<T>(
    asyncFn: () => Promise<T>,
    maxRetries: number = 5,
    delayMs: number = 1000
): Promise<T> {
    let lastError: Error;

    for (let attempt = 1; attempt <= maxRetries; attempt++) {
        try {
            return await asyncFn();
        } catch (error) {
            lastError = error as Error;
            trace(`Attempt ${attempt} failed: ${error}`);

            if (attempt < maxRetries) {
                trace(`Retrying in ${delayMs}ms...`);
                await new Promise(resolve => setTimeout(resolve, delayMs));
            }
        }
    }

    throw lastError!;
}

