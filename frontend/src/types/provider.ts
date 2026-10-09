export interface ProviderProfile {
  id: string;
  name: string;
  baseURL: string;
  model: string;
  credentialRef: string;
  hasCredential: boolean;
  enabled: boolean;
  insecureTransport?: boolean;
}

export interface SaveInput {
  id?: string;
  name: string;
  baseURL: string;
  model: string;
  apiKey: string;
  enabled?: boolean;
}

export interface TestResult {
  ok: boolean;
  code?: string;
  message: string;
  retryable?: boolean;
  statusCode?: number;
  retryAfterSec?: number;
  insecure?: boolean;
}
