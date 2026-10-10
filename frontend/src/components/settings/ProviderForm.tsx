// ProviderForm adds or edits one provider. The parent mounts it with a `key`
// per target, so switching targets starts from a fresh state instead of an
// effect that resets fields. Validation messages mirror the backend
// (internal/app/providers.go); change them together.
import { useEffect, useId, useRef, useState } from "react";
import type { JSX, KeyboardEvent } from "react";
import { app } from "../../../wailsjs/go/models";
import type { ModelEntry, ProviderProfile, TestResult } from "../../types/provider";
import { IconAlert, IconCheck, IconEyeOff, IconKey, IconX } from "../ui/Icon";
import { PROVIDER_PRESETS, type ProviderPreset } from "./presets";

type SaveInput = app.SaveInput;

interface FormState {
  name: string;
  baseURL: string;
  models: ModelEntry[];
  activeModel: string;
  apiKey: string;
}

const EMPTY: FormState = { name: "", baseURL: "https://", models: [], activeModel: "", apiKey: "" };

function fromProfile(found: ProviderProfile): FormState {
  // Never fill apiKey: stored secret stays hidden, field means "replace".
  // Copy both id and alias so the existing alias round-trips into the form.
  return {
    name: found.name,
    baseURL: found.baseURL,
    models: found.models.map((m) => ({ id: m.id, alias: m.alias })),
    activeModel: found.activeModel || found.model,
    apiKey: "",
  };
}

function toSave(form: FormState, id: string): SaveInput {
  return new app.SaveInput({
    id,
    name: form.name,
    baseURL: form.baseURL,
    models: form.models.map((m) => ({ id: m.id, alias: m.alias })),
    model: form.activeModel,
    apiKey: form.apiKey,
    enabled: true,
  });
}

interface Props {
  /** Provider being edited; null adds a new one. */
  profile: ProviderProfile | null;
  test: TestResult | null;
  testing: boolean;
  onSave: (in_: SaveInput) => Promise<ProviderProfile | null>;
  onTest: (in_: SaveInput) => Promise<TestResult>;
  onSaved: (saved: ProviderProfile) => void;
  /** Absent when the form cannot be dismissed (no provider exists yet). */
  onCancel?: () => void;
}

export function ProviderForm({ profile, test, testing, onSave, onTest, onSaved, onCancel }: Props): JSX.Element {
  const editing = profile?.id ?? "";
  const [form, setForm] = useState<FormState>(() => (profile ? fromProfile(profile) : EMPTY));
  const [modelDraft, setModelDraft] = useState("");
  const [aliasDraft, setAliasDraft] = useState("");
  const [formError, setFormError] = useState<string | null>(null);
  const [showKey, setShowKey] = useState(false);
  const [preset, setPreset] = useState<ProviderPreset | null>(null);
  const rootRef = useRef<HTMLDivElement>(null);
  const nameRef = useRef<HTMLInputElement>(null);
  const firstPresetRef = useRef<HTMLButtonElement>(null);
  const titleId = useId();
  const keyNoteId = useId();
  const urlNoteId = useId();

  // Opening the form lands the user on its first control, never silently
  // below the fold: the preset row for a new provider, Nama when editing.
  useEffect(() => {
    rootRef.current?.scrollIntoView?.({ block: "nearest" });
    (firstPresetRef.current ?? nameRef.current)?.focus();
  }, []);

  // A preset replaces the endpoint, and the name only while the user has not
  // typed their own.
  const applyPreset = (next: ProviderPreset) => {
    setForm((f) => ({
      ...f,
      name: !f.name.trim() || f.name === preset?.name ? next.name : f.name,
      baseURL: next.baseURL,
    }));
    setPreset(next);
  };

  // Adds the trimmed draft to the model list. Duplicates are ignored so the
  // active model is unambiguous; the first model added becomes active.
  const addModel = () => {
    const id = modelDraft.trim();
    if (!id) return;
    if (form.models.some((m) => m.id === id)) {
      setFormError(`Model "${id}" sudah ada di daftar.`);
      return;
    }
    setFormError(null);
    setForm({
      ...form,
      models: [...form.models, { id, alias: aliasDraft.trim() }],
      activeModel: form.activeModel || id,
    });
    setModelDraft("");
    setAliasDraft("");
  };

  // Renames the display alias of an entry. The id is the identity and never
  // changes here; to change the id, remove the entry and add it again.
  const renameAlias = (id: string, alias: string) => {
    setForm({
      ...form,
      models: form.models.map((m) => (m.id === id ? { ...m, alias } : m)),
    });
  };

  const removeModel = (id: string) => {
    const models = form.models.filter((x) => x.id !== id);
    setForm({
      ...form,
      models,
      activeModel: form.activeModel === id ? (models[0]?.id ?? "") : form.activeModel,
    });
  };

  const submit = async () => {
    setFormError(null);
    if (!form.name.trim()) {
      setFormError("Nama wajib diisi.");
      return;
    }
    if (form.models.length === 0) {
      setFormError("Tambahkan minimal satu model.");
      return;
    }
    if (!editing && !form.apiKey.trim()) {
      setFormError("API key wajib diisi untuk provider baru.");
      return;
    }
    const saved = await onSave(toSave(form, editing));
    if (saved) onSaved(saved);
  };

  const enterAdds = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      e.preventDefault();
      addModel();
    }
  };

  return (
    <div className="prov-form" ref={rootRef} role="group" aria-labelledby={titleId}>
      <h3 className="prov-form__title" id={titleId}>{editing ? "Ubah provider" : "Provider baru"}</h3>
      {!editing && (
        <div className="field">
          <span className="field__label" id={`${titleId}-presets`}>Isi otomatis untuk</span>
          <div className="preset-row" role="group" aria-labelledby={`${titleId}-presets`}>
            {PROVIDER_PRESETS.map((pr, i) => (
              <button
                key={pr.name}
                type="button"
                className="chip"
                ref={i === 0 ? firstPresetRef : undefined}
                aria-pressed={preset?.name === pr.name}
                onClick={() => applyPreset(pr)}
              >
                {pr.name}
              </button>
            ))}
          </div>
        </div>
      )}
      <div className="field">
        <label className="field__label" htmlFor="settings-name">Nama</label>
        <input id="settings-name" ref={nameRef} className="input" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} placeholder="Provider pribadi" />
      </div>
      <div className="field">
        <label className="field__label" htmlFor="settings-baseurl">Base URL</label>
        <input id="settings-baseurl" className="input" value={form.baseURL} onChange={(e) => setForm({ ...form, baseURL: e.target.value })} placeholder="https://api.example.com/v1" inputMode="url" aria-describedby={urlNoteId} />
        <div className="field__hint" id={urlNoteId}>Endpoint yang kompatibel dengan OpenAI, biasanya berakhiran /v1.</div>
      </div>
      <div className="prov-form__models">
        <div className="field">
          <span className="field__label" id="settings-model-label">Model</span>
          <span className="model-add-row">
            <input
              className="input"
              aria-labelledby="settings-model-label"
              value={modelDraft}
              onChange={(e) => setModelDraft(e.target.value)}
              onKeyDown={enterAdds}
              placeholder={preset?.modelHint ?? "ID model, mis. gpt-4.1-mini"}
            />
            <input
              className="input"
              aria-label="Alias model baru"
              value={aliasDraft}
              onChange={(e) => setAliasDraft(e.target.value)}
              onKeyDown={enterAdds}
              placeholder="Alias (opsional)"
            />
            <button type="button" className="btn" onClick={addModel} disabled={!modelDraft.trim()}>Tambah</button>
          </span>
        </div>
        {form.models.length > 0 && (
          <div className="model-list">
            {form.models.map((m) => (
              <div key={m.id} className={form.activeModel === m.id ? "model-item model-item--active" : "model-item"}>
                <label className="model-pick" title="Pakai model ini untuk chat">
                  <input
                    type="radio"
                    name={`active-${editing || "new"}`}
                    checked={form.activeModel === m.id}
                    onChange={() => setForm({ ...form, activeModel: m.id })}
                  />
                  <span className="model-pick__label" title={m.id}>{m.alias ? `${m.alias} (${m.id})` : m.id}</span>
                </label>
                <input
                  className="input model-alias"
                  value={m.alias}
                  onChange={(e) => renameAlias(m.id, e.target.value)}
                  placeholder="Alias"
                  aria-label={`Alias untuk ${m.id}`}
                />
                <button type="button" className="icon-btn icon-btn--danger" onClick={() => removeModel(m.id)} title="Hapus model" aria-label={`Hapus model ${m.id}`}>
                  <IconX size={14} aria-hidden="true" />
                </button>
              </div>
            ))}
          </div>
        )}
        <div className="field__hint">Model yang dipilih dipakai untuk chat. Alias hanya nama tampilan.</div>
      </div>
      <div className="field">
        <label className="field__label" htmlFor="settings-apikey">API key</label>
        <span className="settings-keyrow">
          <input id="settings-apikey" className="input" type={showKey ? "text" : "password"} value={form.apiKey} onChange={(e) => setForm({ ...form, apiKey: e.target.value })} placeholder={editing ? "•••••• (tersimpan)" : (preset?.keyHint ?? "sk-…")} autoComplete="off" aria-describedby={keyNoteId} />
          <button
            type="button"
            className="btn btn--sm"
            onClick={() => setShowKey(!showKey)}
            title={showKey ? "Sembunyikan kunci" : "Tampilkan kunci"}
            aria-label={showKey ? "Sembunyikan API key" : "Tampilkan API key"}
          >
            {showKey ? <IconEyeOff size={14} aria-hidden="true" /> : <IconKey size={14} aria-hidden="true" />}
            {showKey ? "Sembunyikan" : "Tampilkan"}
          </button>
        </span>
        <div className="field__hint" id={keyNoteId}>
          {editing && "Kosongkan untuk tetap memakai key yang tersimpan. "}
          {preset?.local && "Server lokal tidak memeriksa key, isi apa saja. "}
          Disimpan di Windows Credential Manager, tidak pernah ditampilkan lagi.
        </div>
      </div>
      {formError && <div id="settings-form-error" className="banner banner--error" role="alert">
        <span className="banner__icon"><IconAlert size={14} aria-hidden="true" /></span>
        <span>{formError}</span>
      </div>}
      {test && <div id="settings-test" className={test.ok ? "banner banner--ok" : "banner banner--error"} role="status">
        <span className="banner__icon">{test.ok ? <IconCheck size={14} aria-hidden="true" /> : <IconAlert size={14} aria-hidden="true" />}</span>
        <span>{test.message}{test.insecure ? " Koneksi tidak terenkripsi (HTTP)." : ""}</span>
      </div>}
      <div className="prov-form__actions">
        <button type="button" className="btn" onClick={() => onTest(toSave(form, editing))} disabled={testing}>
          {testing && <span className="spinner" aria-hidden="true" />}
          {testing ? "Mengetes…" : "Tes koneksi"}
        </button>
        <button type="button" className="btn btn--primary" onClick={submit}>Simpan</button>
        {onCancel && <button type="button" className="btn btn--ghost" onClick={onCancel}>Batal</button>}
      </div>
    </div>
  );
}
