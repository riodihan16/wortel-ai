// Starting points for the new-provider form. A preset only prefills fields the
// user can still edit; nothing is sent until Tes koneksi or Simpan. Model
// names are examples in placeholders, never saved, because catalogues change.
export interface ProviderPreset {
  name: string;
  baseURL: string;
  modelHint: string;
  keyHint: string;
  /** Local servers accept any non-empty key. */
  local: boolean;
}

export const PROVIDER_PRESETS: readonly ProviderPreset[] = [
  { name: "OpenAI", baseURL: "https://api.openai.com/v1", modelHint: "mis. gpt-4.1-mini", keyHint: "sk-…", local: false },
  { name: "OpenRouter", baseURL: "https://openrouter.ai/api/v1", modelHint: "mis. openrouter/auto", keyHint: "sk-or-…", local: false },
  { name: "Groq", baseURL: "https://api.groq.com/openai/v1", modelHint: "mis. llama-3.3-70b-versatile", keyHint: "gsk_…", local: false },
  { name: "LM Studio", baseURL: "http://localhost:1234/v1", modelHint: "ID model yang sedang dimuat", keyHint: "isi bebas, mis. lm-studio", local: true },
  { name: "Ollama", baseURL: "http://localhost:11434/v1", modelHint: "mis. llama3.2", keyHint: "isi bebas, mis. ollama", local: true },
];
