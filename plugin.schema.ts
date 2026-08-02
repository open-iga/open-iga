import { z } from 'zod';
import { writeFileSync } from 'node:fs';
import path from 'node:path';

const entitlementActionSchema = z.object({
    export: z.string().describe('Name of the exported function name'),
    description: z.string().describe('Description of the exported function name'),
    endpoint: z.array(
        z.object({
            method: z.enum(['GET', 'POST', 'PATCH', 'PUT']).describe('HTTP Method'),
            url: z.string().describe('Upstream URL that needs to be whitelisted'),
        }),
    ),
});

const entitlementDetailsSchema = z.object({
    description: z
        .string()
        .describe(
            'List of actions supported by the entitlement. This will be shown to the user when choosing the entitlement',
        ),
    actions: z.record(z.enum(['provision', 'deprovision', 'list', 'update']), entitlementActionSchema),
});

const pluginSchema = z.object({
    name: z.string().describe('Name of the plugin'),
    version: z.string().describe('Version of the plugin'),
    environmentVariables: z
        .array(z.string())
        .describe(
            'Environment variables required by the plugin. Provide an empty aray if no environment variables are required.',
        ),
    entitlements: z
        .record(z.string(), entitlementDetailsSchema)
        .describe(
            'Entitlement supported by the plugin. This is a key-value object where key is the supported entitlement with actions as value.',
        ),
});

const generatePluginSchema = () => {
    const jsonSchema = pluginSchema.toJSONSchema();
    writeFileSync(path.resolve(process.cwd(), 'plugin.schema.json'), `${JSON.stringify(jsonSchema, null, 2)}\n`);
};

generatePluginSchema();
