const base64url = (buf: Uint8Array): string => Buffer.from(buf).toString('base64url');

export const generateRandomValues = (byteSize: number) => base64url(crypto.getRandomValues(new Uint8Array(byteSize)));

export async function generateCodeChallenge(verifier: string): Promise<string> {
    const encoded = new TextEncoder().encode(verifier);
    const hash = await crypto.subtle.digest('SHA-256', encoded);
    return base64url(new Uint8Array(hash));
}
