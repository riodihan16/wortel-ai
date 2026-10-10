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
type Field = "name" | "model" | "key";

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
  /** A field a test depends on changed: the last result no longer applies. */
  onDirty?: () => void;
}

export function ProviderForm({ profile, test, testing, onSave, onTest, onSaved, onCancel, onDirty }: Props): JSX.Element {
  const editing = profile?.id ?? "";
  const [form, setForm] = useState<FormState>(() => (profile ? fromProfile(profile) : EMPTY));
  const [modelDraft, setModelDraft] = useState("");
  const [formError, setFormError] = useState<string | null>(null);
  const [showKey, setShowKey] = useState(false);
  const [preset, setPreset] = useState<ProviderPreset | null>(null);
  const [saving, setSaving] = useState(false);
  const [invalid, setInvalid] = useState<Field | null>(null);
  // Bumped on every failed save so the same field is refocused each time.
  const [failTick, setFailTick] = useState(0);
  const rootRef = useRef<HTMLDivElement>(null);
  const nameRef = useRef<HTMLInputElement>(null);
  const modelRef = useRef<HTMLInputElement>(null);
  const keyRef = useRef<HTMLInputElement>(null);
  const errorId = useId();
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
      fail("model", `Model "${id}" sudah ada di daftar.`);
      return;
    }
    setFormError(null);
    setForm({
      ...form,
      models: [...form.models, { id, alias: "" }],
      activeModel: form.activeModel || id,
    });
    setModelDraft("");
    onDirty?.();
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
    onDirty?.();
  };

  // A model ID typed but never added with Tambah still counts: the user's
  // intent is plain, and rejecting it with "add at least one model" while the
  // ID sits in the field is the trap the critique found. Commits it to the
  // list so the form shows what will be sent.
  const commitDraft = (): FormState => {
    const id = modelDraft.trim();
    if (!id || form.models.some((m) => m.id === id)) return form;
    const next = {
      ...form,
      models: [...form.models, { id, alias: "" }],
      activeModel: form.activeModel || id,
    };
    setForm(next);
    setModelDraft("");
    return next;
  };

  const fail = (field: Field, message: string) => {
    setInvalid(field);
    setFormError(message);
    setFailTick((t) => t + 1);
  };

  // Focus after the render that commits a model draft: the new model row
  // moves the fields below it, and focusing earlier left the field and its
  // message off-screen. Centred so the message under it is visible too.
  useEffect(() => {
    if (!invalid) return;
    const el = { name: nameRef, model: modelRef, key: keyRef }[invalid].current;
    el?.scrollIntoView?.({ block: "center" });
    el?.focus();
  }, [invalid, failTick]);

  // The reason sits under the field it is about, not at the end of the form.
  const fieldError = (field: Field) =>
    invalid === field && formError ? (
      <div id={errorId} className="field__error" role="alert">
        <IconAlert size={13} aria-hidden="true" />
        <span>{formError}</span>
      </div>
    ) : null;

  // Field edits clear their own error mark and any test result they outdate.
  const change = (patch: Partial<FormState>, field?: Field) => {
    setForm((f) => ({ ...f, ...patch }));
    if (field && invalid === field) {
      setInvalid(null);
      setFormError(null);
    }
    if (patch.baseURL !== undefined || patch.apiKey !== undefined || patch.models !== undefined || patch.activeModel !== undefined) {
      onDirty?.();
    }
  };

  const describedBy = (field: Field, extra?: string) =>
    [extra, invalid === field ? errorId : undefined].filter(Boolean).join(" ") || undefined;

  // A test without a model cannot say anything useful; say why instead.
  const runTest = () => {
    setFormError(null);
    setInvalid(null);
    const next = commitDraft();
    if (next.models.length === 0) {
      fail("model", "Tulis ID model dulu untuk dites.");
      return;
    }
    onTest(toSave(next, editing));
  };

  const submit = async () => {
    // A second press while the first save is in flight would create a second
    // profile (a new provider has no id yet).
    if (saving) return;
    setFormError(null);
    setInvalid(null);
    const next = commitDraft();
    if (!next.name.trim()) {
      fail("name", "Nama wajib diisi.");
      return;
    }
    if (next.models.length === 0) {
      fail("model", "Tulis minimal satu ID model.");
      return;
    }
    if (!editing && !next.apiKey.trim()) {
      fail("key", "API key wajib diisi untuk provider baru.");
      return;
    }
    setSaving(true);
    try {
      const saved = await onSave(toSave(next, editing));
      if (saved) onSaved(saved);
    } finally {
      setSaving(false);
    }
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
          <span className="field__label" id={`${titleId}-presets`}>Pilih layanan untuk isi otomatis (opsional)</span>
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
          {preset && <div className="field__hint">{preset.source}</div>}
        </div>
      )}
      <div className="field">
        <label className="field__label" htmlFor="settings-name">Nama</label>
        <input id="settings-name" ref={nameRef} className="input" value={form.name} onChange={(e) => change({ name: e.target.value }, "name")} placeholder="Provider pribadi" aria-invalid={invalid === "name"} aria-describedby={describedBy("name")} />
        {fieldError("name")}
      </div>
      <div className="field">
        <label className="field__label" htmlFor="settings-baseurl">Alamat API (Base URL)</label>
        <input id="settings-baseurl" className="input" value={form.baseURL} onChange={(e) => change({ baseURL: e.target.value })} placeholder="https://api.example.com/v1" inputMode="url" aria-describedby={urlNoteId} />
        <div className="field__hint" id={urlNoteId}>Endpoint yang kompatibel dengan OpenAI, biasanya berakhiran /v1.</div>
      </div>
      <div className="prov-form__models">
        <div className="field">
          <span className="field__label" id="settings-model-label">Model</span>
          <span className="model-add-row">
            <input
              ref={modelRef}
              className="input"
              aria-labelledby="settings-model-label"
              aria-invalid={invalid === "model"}
              aria-describedby={describedBy("model")}
              value={modelDraft}
              onChange={(e) => {
                setModelDraft(e.target.value);
                if (invalid === "model") {
                  setInvalid(null);
                  setFormError(null);
                }
                onDirty?.();
              }}
              onKeyDown={enterAdds}
              placeholder={preset?.modelHint ?? "ID model, mis. gpt-4.1-mini"}
            />
            <button type="button" className="btn" onClick={addModel} disabled={!modelDraft.trim()}>Tambah</button>
          </span>
          {fieldError("model")}
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
                    onChange={() => change({ activeModel: m.id })}
                  />
                  <span className="model-pick__label" title={m.id}>{m.id}</span>
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
        <div className="field__hint">
          Model yang dipilih dipakai untuk chat.
          {form.models.length > 0 && " Alias (opsional) diisi di barisnya, hanya nama tampilan."}
        </div>
      </div>
      <div className="field">
        <label className="field__label" htmlFor="settings-apikey">API key</label>
        <span className="settings-keyrow">
          <input id="settings-apikey" ref={keyRef} className="input" type={showKey ? "text" : "password"} value={form.apiKey} onChange={(e) => change({ apiKey: e.target.value }, "key")} placeholder={editing ? "•••••• (tersimpan)" : (preset?.keyHint ?? "sk-…")} autoComplete="off" aria-invalid={invalid === "key"} aria-describedby={describedBy("key", keyNoteId)} />
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
        {fieldError("key")}
      </div>
      {test && <div id="settings-test" className={test.ok ? "banner banner--ok" : "banner banner--error"} role="status">
        <span className="banner__icon">{test.ok ? <IconCheck size={14} aria-hidden="true" /> : <IconAlert size={14} aria-hidden="true" />}</span>
        <span>{test.message}{test.insecure ? " Koneksi tidak terenkripsi (HTTP)." : ""}</span>
      </div>}
      <div className="prov-form__actions">
        <button type="button" className="btn" onClick={runTest} disabled={testing}>
          {testing && <span className="spinner" aria-hidden="true" />}
          {testing ? "Mengetes…" : "Tes koneksi"}
        </button>
        <button type="button" className="btn btn--primary" onClick={submit} disabled={saving}>
          {saving && <span className="spinner" aria-hidden="true" />}
          {saving ? "Menyimpan…" : "Simpan"}
        </button>
        {onCancel && <button type="button" className="btn btn--ghost" onClick={onCancel}>Batal</button>}
      </div>
    </div>
  );
}
