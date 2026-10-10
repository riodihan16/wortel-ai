// Starting points for the new-provider form. A preset only prefills fields the
// user can still edit; nothing is sent until Tes koneksi or Simpan. Model
// names are examples in placeholders, never saved, because catalogues change.
export interface ProviderPreset {
  name: string;
  baseURL: string;
  modelHint: string;
  keyHint: string;
  /** Where the key and model IDs come from, shown once the preset is picked. */
  source: string;
  /** Local servers accept any non-empty key. */
  local: boolean;
}

export const PROVIDER_PRESETS: readonly ProviderPreset[] = [
  { name: "OpenAI", baseURL: "https://api.openai.com/v1", modelHint: "mis. gpt-4.1-mini", keyHint: "sk-…", source: "Key: platform.openai.com/api-keys · ID model: platform.openai.com/docs/models", local: false },
  { name: "OpenRouter", baseURL: "https://openrouter.ai/api/v1", modelHint: "mis. openrouter/auto", keyHint: "sk-or-…", source: "Key: openrouter.ai/keys · ID model: openrouter.ai/models", local: false },
  { name: "Groq", baseURL: "https://api.groq.com/openai/v1", modelHint: "mis. llama-3.3-70b-versatile", keyHint: "gsk_…", source: "Key: console.groq.com/keys · ID model: console.groq.com/docs/models", local: false },
  { name: "LM Studio", baseURL: "http://localhost:1234/v1", modelHint: "ID model yang sedang dimuat", keyHint: "isi bebas, mis. lm-studio", source: "Jalankan server di LM Studio (tab Developer → Start Server); ID model = nama model yang dimuat.", local: true },
  { name: "Ollama", baseURL: "http://localhost:11434/v1", modelHint: "mis. llama3.2", keyHint: "isi bebas, mis. ollama", source: "Jalankan Ollama; ID model dari perintah `ollama list`.", local: true },
];
