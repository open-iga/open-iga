import type { operations } from '@/utils/openapi/schema';

type ValidationDetails =
    operations['getConnectorOnboardingRequestDetails']['responses'][200]['content']['application/json'];

// The connector spec and its operations are derived from the generated OpenAPI schema so they never drift from the API.
export type ConnectorSpec = NonNullable<ValidationDetails['connectorSpec']>;
export type Operation = NonNullable<ConnectorSpec['actions'][string]['create']>;
