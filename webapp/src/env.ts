import { defineEnvVars } from '@sveltejs/kit/env';

const DEFAULT_API_URL = 'http://localhost:8080';

export const variables = defineEnvVars({
	API_URL: {
		description: 'Base URL of the Go API the pages are built from',
		schema: (value) => (value ?? DEFAULT_API_URL).replace(/\/+$/, '')
	}
});
