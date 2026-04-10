import { z } from 'zod';

const ERROR_CODES = ['INVALID_STATE'];

export const GenericErrorSchema = z.object({
    errorCode: z.enum(ERROR_CODES),
    message: z.string(),
});

export const jsonContent = ({
    schema,
    description = 'Unavailable',
}: {
    schema: z.ZodSchema<unknown>;
    description?: string;
}) => ({
    content: {
        'application/json': {
            schema,
        },
    },
    description,
});
