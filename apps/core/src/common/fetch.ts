import { type, type Type } from 'arktype';

type SafeParseResult<T> = { response: T } | { errorType: 'validationError' | 'fetchError'; error: string };

export const safeFetch = async <T>({
    endpoint,
    init,
    typeChecker,
}: {
    endpoint: string | URL;
    init?: BunFetchRequestInit;
    typeChecker: Type<T>;
}): Promise<SafeParseResult<typeof typeChecker.inferOut>> => {
    try {
        const response = await fetch(endpoint, init);

        const data = await response.json();

        const parsedResponse = typeChecker(data);
        if (parsedResponse instanceof type.errors) {
            return { errorType: 'validationError', error: parsedResponse.summary };
        }

        return { response: parsedResponse };
    } catch (error) {
        const errorMessage = error instanceof Error ? error.message : String(error);

        return {
            errorType: 'fetchError',
            error: errorMessage,
        };
    }
};
