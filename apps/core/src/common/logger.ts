import { styleText } from 'node:util';

const LEVELS = { debug: 0, info: 1, warn: 2, error: 3 } as const;

type AdditionalContext = Record<string, unknown>;
type SupportedLogLevels = keyof typeof LEVELS;

const textColorMap: Record<SupportedLogLevels, Parameters<typeof styleText>[0]> = {
    info: 'blue',
    error: 'red',
    warn: 'yellow',
    debug: 'green',
};

const isPrettyPrintEnabled = process.env['PRETTY_PRINT'] === 'true';
const log = (args: { level: SupportedLogLevels; message: string; additionalContext?: AdditionalContext }) => {
    const { level, message, additionalContext } = args;
    const timestamp = new Date().toISOString();

    let logMessage = '';

    if (isPrettyPrintEnabled) {
        const unstyledLogMessage = `${level}: ${timestamp} - ${message}${additionalContext ? `. ${JSON.stringify(additionalContext)}` : ''}`;

        logMessage = styleText(textColorMap[level], unstyledLogMessage);
    } else {
        logMessage = JSON.stringify({
            level,
            timestamp,
            message,
            ...additionalContext,
        });
    }
    process[level === 'error' ? 'stderr' : 'stdout'].write(`${logMessage}\n`);
};

export const logger = {
    debug: (message: string, additionalContext?: AdditionalContext) =>
        log({ level: 'debug', message, additionalContext }),
    info: (message: string, additionalContext?: AdditionalContext) =>
        log({ level: 'info', message, additionalContext }),
    warn: (message: string, additionalContext?: AdditionalContext) =>
        log({ level: 'warn', message, additionalContext }),
    error: (message: string, additionalContext?: AdditionalContext) =>
        log({ level: 'error', message, additionalContext }),
};
