// SettingsPanel owns the provider list, provider add/edit form, appearance,
// autostart and global-shortcut settings. Form logic (EMPTY shape, toSave,
// pick/addModel/renameAlias/removeModel/submit, validation messages,
// reset-on-success) mirrors the backend contract; change it together with
// internal/app/providers.go. Provider delete asks inline (ADR-048 pattern) and
// is never offered on the active provider, which the backend refuses.
import { useEffect, useRef, useState } from "react";
import { app } from "../../../wailsjs/go/models";
import type { ModelEntry, ProviderProfile, TestResult } from "../../types/provider";
import {
  IconAlert,
  IconCheck,
  IconChevronDown,
  IconEyeOff,
  IconKey,
  IconX,
} from "../ui/Icon";
import { Segmented } from "../ui/Segmented";

type SaveInput = app.SaveInput;

interface Props {
  providers: ProviderProfile[];
  activeId: string;
  listError: string | null;
  test: TestResult | null;
  testing: boolean;
  onSave: (in_: SaveInput) => Promise<ProviderProfile | null>;
  onDelete: (id: string) => void;
  onActivate: (id: string) => void;
  onTest: (in_: SaveInput) => Promise<TestResult>;
  onClose: () => void;
  autostart: boolean;
  autostartErr: string | null;
  onToggleAutostart: () => void;
  hotkeyErr: string;
  shortcut: app.Shortcut;
  shortcutErr: string;
  onShortcutChange: (next: app.Shortcut) => void;
  capturing: boolean;
  onCaptureToggle: () => void;
  appearance: app.Appearance;
  appearanceErr: string;
  onAppearanceChange: (next: app.Appearance) => void;
  version: string;
}

const EMPTY = { name: "", baseURL: "https://", models: [] as ModelEntry[], activeModel: "", apiKey: "" };

function toSave(form: typeof EMPTY, id: string): SaveInput {
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

export function SettingsPanel(p: Props) {
  const [form, setForm] = useState(EMPTY);
  const [modelDraft, setModelDraft] = useState("");
  const [aliasDraft, setAliasDraft] = useState("");
  const [editing, setEditing] = useState<string>("");
  const [formError, setFormError] = useState<string | null>(null);
  const [showKey, setShowKey] = useState(false);
  const [confirmDel, setConfirmDel] = useState<string | null>(null);
  const delBtns = useRef(new Map<string, HTMLButtonElement>());
  const listTitleRef = useRef<HTMLHeadingElement>(null);
  const formRef = useRef<HTMLDivElement>(null);
  const nameRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (!editing) setForm(EMPTY);
  }, [editing, p.providers.length]);

  // Ubah must land the user on the form, not leave it loading silently below
  // the fold.
  useEffect(() => {
    if (!editing) return;
    formRef.current?.scrollIntoView?.({ block: "start" });
    nameRef.current?.focus();
  }, [editing]);

  // A confirm row for a provider that no longer exists must not stay armed.
  const armedDel = p.providers.some((x) => x.id === confirmDel) ? confirmDel : null;

  const cancelDelete = (id: string) => {
    setConfirmDel(null);
    delBtns.current.get(id)?.focus();
  };

  const confirmDelete = (id: string) => {
    setConfirmDel(null);
    if (editing === id) setEditing("");
    p.onDelete(id);
    // The card and its focused button unmount; keep focus inside the section.
    listTitleRef.current?.focus();
  };

  const pick = (id: string) => {
    const found = p.providers.find((x) => x.id === id);
    if (!found) return;
    setEditing(id);
    // Never fill apiKey: stored secret stays hidden, field means "replace".
    // Copy both id and alias so the existing alias round-trips into the form.
    setForm({
      name: found.name,
      baseURL: found.baseURL,
      models: found.models.map((m) => ({ id: m.id, alias: m.alias })),
      activeModel: found.activeModel || found.model,
      apiKey: "",
    });
    setFormError(null);
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
    const saved = await p.onSave(toSave(form, editing));
    if (saved) {
      setEditing("");
      setForm(EMPTY);
      setModelDraft("");
      setAliasDraft("");
      setShowKey(false);
    }
  };

  return (
    <>
      <div id="settings-header" className="drag-handle">
        <span className="brand">
          <span className="brand__name">Pengaturan</span>
        </span>
        <span className="header-actions">
          <button type="button" className="btn btn--sm settings-back" onClick={p.onClose}>
            <IconChevronDown size={14} className="settings-back__icon" aria-hidden="true" />
            Kembali
          </button>
        </span>
      </div>
      <div id="settings">
        <section className="section" aria-label="Provider">
          <h2 className="section__title" ref={listTitleRef} tabIndex={-1}>Provider</h2>
          {p.listError && (
            <div className="banner banner--error" role="alert">
              <span className="banner__icon"><IconAlert size={14} aria-hidden="true" /></span>
              <span>{p.listError}</span>
            </div>
          )}
          <div id="settings-list">
            {p.providers.length === 0 && <p id="settings-empty">Belum ada provider. Tambahkan endpoint di bawah.</p>}
            {p.providers.map((pr) => (
              <div key={pr.id} className={p.activeId === pr.id ? "card card--active" : "card"} role="group" aria-label={`Provider ${pr.name}`}>
                <div className="prov-head">
                  <span className="prov-name" title={pr.name}>{pr.name}</span>
                  {p.activeId === pr.id && <span className="prov-badge">Aktif</span>}
                </div>
                <div className="prov-meta">{pr.baseURL} · {pr.activeModel || pr.model} · {pr.models.length} model · {pr.hasCredential ? "key tersimpan" : "tanpa key"}</div>
                <div className="prov-actions">
                  {p.activeId !== pr.id && <button type="button" className="btn btn--sm" onClick={() => p.onActivate(pr.id)} aria-label={`Aktifkan provider ${pr.name}`}>Aktifkan</button>}
                  <button type="button" className="btn btn--sm" onClick={() => { setConfirmDel(null); pick(pr.id); }} aria-label={`Ubah provider ${pr.name}`}>Ubah</button>
                  {p.activeId !== pr.id && (
                    <button
                      type="button"
                      className="btn btn--sm btn--danger"
                      ref={(el) => { if (el) delBtns.current.set(pr.id, el); else delBtns.current.delete(pr.id); }}
                      onClick={() => setConfirmDel(pr.id)}
                      aria-label={`Hapus provider ${pr.name}`}
                      aria-expanded={armedDel === pr.id}
                    >
                      Hapus
                    </button>
                  )}
                </div>
                {armedDel === pr.id && (
                  <div
                    className="confirm-row"
                    role="group"
                    aria-label="Konfirmasi hapus provider"
                    onKeyDown={(e) => {
                      if (e.key === "Escape") cancelDelete(pr.id);
                    }}
                  >
                    <IconAlert size={14} />
                    <span className="confirm-row__msg">
                      <span>Hapus “</span>
                      <span className="confirm-row__title">{pr.name}</span>
                      <span>” dan API key-nya?</span>
                    </span>
                    <button type="button" className="btn btn--sm btn--danger" onClick={() => confirmDelete(pr.id)}>
                      Hapus
                    </button>
                    {/* Focused first: Enter on an armed row must not delete. */}
                    <button type="button" className="btn btn--sm" onClick={() => cancelDelete(pr.id)} autoFocus>
                      Batal
                    </button>
                  </div>
                )}
              </div>
            ))}
          </div>
          <div id="settings-form" ref={formRef}>
            <h3 className="settings-form-title">{editing ? "Ubah provider" : "Provider baru"}</h3>
            <div className="field">
              <label className="field__label" htmlFor="settings-name">Nama</label>
              <input id="settings-name" ref={nameRef} className="input" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} placeholder="Provider pribadi" />
            </div>
            <div className="field">
              <label className="field__label" htmlFor="settings-baseurl">Base URL</label>
              <input id="settings-baseurl" className="input" value={form.baseURL} onChange={(e) => setForm({ ...form, baseURL: e.target.value })} placeholder="https://api.example.com/v1" inputMode="url" />
            </div>
            <div className="field">
              <span className="field__label" id="settings-model-label">Model</span>
              <span className="model-add-row">
                <input
                  className="input"
                  aria-labelledby="settings-model-label"
                  value={modelDraft}
                  onChange={(e) => setModelDraft(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter") {
                      e.preventDefault();
                      addModel();
                    }
                  }}
                  placeholder="example-model"
                />
                <input
                  className="input"
                  aria-label="Alias model baru"
                  value={aliasDraft}
                  onChange={(e) => setAliasDraft(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter") {
                      e.preventDefault();
                      addModel();
                    }
                  }}
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
            <div className="field__hint field-hint">Radio memilih model aktif untuk chat.</div>
            <div className="field">
              <label className="field__label" htmlFor="settings-apikey">API key {editing && "(kosong = pakai tersimpan)"}</label>
              <span className="settings-keyrow">
                <input id="settings-apikey" className="input" type={showKey ? "text" : "password"} value={form.apiKey} onChange={(e) => setForm({ ...form, apiKey: e.target.value })} placeholder={editing ? "•••••• (tersimpan)" : "sk-..."} autoComplete="off" />
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
            </div>
            {formError && <div id="settings-form-error" className="banner banner--error" role="alert">
              <span className="banner__icon"><IconAlert size={14} aria-hidden="true" /></span>
              <span>{formError}</span>
            </div>}
            {p.test && <div id="settings-test" className={p.test.ok ? "banner banner--ok" : "banner banner--error"} role="status">
              <span className="banner__icon">{p.test.ok ? <IconCheck size={14} aria-hidden="true" /> : <IconAlert size={14} aria-hidden="true" />}</span>
              <span>{p.test.ok ? "OK: " : ""}{p.test.message}{p.test.insecure ? " (HTTP tanpa TLS!)" : ""}</span>
            </div>}
            <div id="settings-form-actions">
              <button type="button" className="btn btn--primary" onClick={submit}>Simpan</button>
              <button type="button" className="btn" onClick={() => p.onTest(toSave(form, editing))} disabled={p.testing}>
                {p.testing && <span className="spinner" aria-hidden="true" />}
                {p.testing ? "Mengetes…" : "Test connection"}
              </button>
              {editing && <button type="button" className="btn btn--ghost" onClick={() => { setEditing(""); setForm(EMPTY); }}>Batal</button>}
            </div>
          </div>
        </section>
        <section className="section" aria-label="Tampilan">
          <h2 className="section__title">Tampilan</h2>
          <div className="settings-appearance-row">
            <span>Mode warna</span>
            <Segmented
              value={p.appearance.theme}
              options={[{ value: "dark", label: "Gelap" }, { value: "light", label: "Terang" }]}
              onChange={(v) => p.onAppearanceChange(new app.Appearance({ theme: v, fontScale: p.appearance.fontScale }))}
              label="Mode warna"
            />
          </div>
          <div className="settings-appearance-row">
            <span>Ukuran teks</span>
            <Segmented
              value={p.appearance.fontScale}
              options={[{ value: 0.9, label: "Kecil" }, { value: 1, label: "Sedang" }, { value: 1.2, label: "Besar" }]}
              onChange={(v) => p.onAppearanceChange(new app.Appearance({ theme: p.appearance.theme, fontScale: v }))}
              label="Ukuran teks"
            />
          </div>
          {p.appearanceErr && <div className="banner banner--error" role="alert">
            <span className="banner__icon"><IconAlert size={14} aria-hidden="true" /></span>
            <span>{p.appearanceErr}</span>
          </div>}
        </section>
        <section className="section" aria-label="Umum">
          <h2 className="section__title">Umum</h2>
          <label className="check">
            <input type="checkbox" checked={p.autostart} onChange={p.onToggleAutostart} />
            Jalankan otomatis saat Windows menyala
          </label>
          {p.autostartErr && <div className="banner banner--error" role="alert">
            <span className="banner__icon"><IconAlert size={14} aria-hidden="true" /></span>
            <span>{p.autostartErr}</span>
          </div>}
          <div id="settings-hotkey">
            <label className="check">
              <input type="checkbox" checked={p.shortcut.enabled} disabled={p.capturing}
                onChange={(e) => p.onShortcutChange(new app.Shortcut({ enabled: e.target.checked, combo: p.shortcut.combo }))} />
              Aktifkan shortcut buka chat
            </label>
            <div className="settings-shortcut-row">
              <span className={p.capturing ? "kbd settings-combo--capturing" : "kbd"}>
                {p.capturing ? "Tekan kombinasi…" : p.shortcut.combo}
              </span>
              <button type="button" className="btn btn--sm" onClick={p.onCaptureToggle}>{p.capturing ? "Batal" : "Ubah"}</button>
            </div>
          </div>
          {p.shortcutErr && <div className="banner banner--error" role="alert">
            <span className="banner__icon"><IconAlert size={14} aria-hidden="true" /></span>
            <span>{p.shortcutErr}</span>
          </div>}
          {p.hotkeyErr && <div className="banner banner--error" role="alert">
            <span className="banner__icon"><IconAlert size={14} aria-hidden="true" /></span>
            <span>{p.hotkeyErr}</span>
          </div>}
        </section>
        <section className="section" aria-label="Tentang">
          <h2 className="section__title">Tentang</h2>
          <div className="about-row">
            <span className="about-row__key">Wortel AI</span>
            <span className="about-row__val">{p.version || "dev"}</span>
          </div>
          <div className="about-row">
            <span className="about-row__key">Lisensi</span>
            <span className="about-row__val">MIT © 2026 Ridho Dwi Ramadhan</span>
          </div>
          <div className="about-row">
            <span className="about-row__key">Font</span>
            <span className="about-row__val">Geist &amp; Geist Mono — SIL Open Font License 1.1</span>
          </div>
          <p className="about-note">
            Data chat dan provider disimpan lokal di perangkat ini. Tidak ada
            telemetry, analytics, atau cloud sync.
          </p>
        </section>
      </div>
    </>
  );
}
