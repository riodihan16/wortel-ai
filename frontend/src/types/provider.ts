export interface ModelEntry {
  id: string;
  alias: string;
}

export interface ProviderProfile {
  id: string;
  name: string;
  baseURL: string;
  /** Active model (mirror of activeModel for older readers). */
  model: string;
  /** All models configured for this provider. */
  models: ModelEntry[];
  /** Model used for chat on this provider. */
  activeModel: string;
  /** Model used for image generation on this provider. Empty = none. */
  imageModel?: string;
  credentialRef: string;
  hasCredential: boolean;
  enabled: boolean;
  insecureTransport?: boolean;
}

export interface SaveInput {
  id?: string;
  name: string;
  baseURL: string;
  models: ModelEntry[];
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
