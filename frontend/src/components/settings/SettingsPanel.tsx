import { useEffect, useState } from "react";
import { app } from "../../../wailsjs/go/models";
import type { ProviderProfile, TestResult } from "../../types/provider";

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
}

const EMPTY = { name: "", baseURL: "https://", models: [] as string[], activeModel: "", apiKey: "" };

function toSave(form: typeof EMPTY, id: string): SaveInput {
  return new app.SaveInput({
    id,
    name: form.name,
    baseURL: form.baseURL,
    models: form.models,
    model: form.activeModel,
    apiKey: form.apiKey,
    enabled: true,
  });
}

export function SettingsPanel(p: Props) {
  const [form, setForm] = useState(EMPTY);
  const [modelDraft, setModelDraft] = useState("");
  const [editing, setEditing] = useState<string>("");
  const [formError, setFormError] = useState<string | null>(null);
  const [showKey, setShowKey] = useState(false);

  useEffect(() => {
    if (!editing) setForm(EMPTY);
  }, [editing, p.providers.length]);

  const pick = (id: string) => {
    const found = p.providers.find((x) => x.id === id);
    if (!found) return;
    setEditing(id);
    // Never fill apiKey: stored secret stays hidden, field means "replace".
    setForm({
      name: found.name,
      baseURL: found.baseURL,
      models: [...found.models],
      activeModel: found.activeModel || found.model,
      apiKey: "",
    });
    setFormError(null);
  };

  // Adds the trimmed draft to the model list. Duplicates are ignored so the
  // active model is unambiguous; the first model added becomes active.
  const addModel = () => {
    const m = modelDraft.trim();
    if (!m) return;
    if (form.models.includes(m)) {
      setFormError(`Model "${m}" sudah ada di daftar.`);
      return;
    }
    setFormError(null);
    setForm({
      ...form,
      models: [...form.models, m],
      activeModel: form.activeModel || m,
    });
    setModelDraft("");
  };

  const removeModel = (m: string) => {
    const models = form.models.filter((x) => x !== m);
    setForm({
      ...form,
      models,
      activeModel: form.activeModel === m ? (models[0] ?? "") : form.activeModel,
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
      setShowKey(false);
    }
  };

  return (
    <div id="settings">
      <div id="settings-header">
        <span>Settings — Provider</span>
        <button onClick={p.onClose}>Kembali</button>
      </div>
      {p.listError && <div id="settings-error" role="alert">{p.listError}</div>}
      <div id="settings-general">
        <h4>Umum</h4>
        <label id="settings-autostart" className="checkrow">
          <input type="checkbox" checked={p.autostart} onChange={p.onToggleAutostart} />
          Jalankan otomatis saat Windows menyala
        </label>
        {p.autostartErr && <div id="settings-general-error" role="alert">{p.autostartErr}</div>}
        <div id="settings-hotkey">
          <label id="settings-shortcut-enabled" className="checkrow">
            <input type="checkbox" checked={p.shortcut.enabled} disabled={p.capturing}
              onChange={(e) => p.onShortcutChange(new app.Shortcut({ enabled: e.target.checked, combo: p.shortcut.combo }))} />
            Aktifkan shortcut buka chat
          </label>
          <div className="shortcut-row">
            <span className={`hotkey-combo${p.capturing ? " hotkey-capturing" : ""}`}>
              {p.capturing ? "Tekan kombinasi…" : p.shortcut.combo}
            </span>
            <button onClick={p.onCaptureToggle}>{p.capturing ? "Batal" : "Ubah"}</button>
          </div>
        </div>
        {p.shortcutErr && <div id="settings-hotkey-error" role="alert">{p.shortcutErr}</div>}
        {p.hotkeyErr && <div id="settings-hotkey-error" role="alert">{p.hotkeyErr}</div>}
      </div>
      <div id="settings-appearance">
        <h4>Tampilan</h4>
        <div className="appearance-row">
          <span>Mode warna</span>
          <div className="seg">
            <button className={p.appearance.theme === 'dark' ? 'seg-on' : ''} onClick={() => p.onAppearanceChange(new app.Appearance({ theme: 'dark', fontScale: p.appearance.fontScale }))}>Gelap</button>
            <button className={p.appearance.theme === 'light' ? 'seg-on' : ''} onClick={() => p.onAppearanceChange(new app.Appearance({ theme: 'light', fontScale: p.appearance.fontScale }))}>Terang</button>
          </div>
        </div>
        <div className="appearance-row">
          <span>Ukuran teks</span>
          <div className="seg">
            {([[0.9, 'Kecil'], [1, 'Sedang'], [1.2, 'Besar']] as const).map(([scale, label]) => (
              <button key={label} className={p.appearance.fontScale === scale ? 'seg-on' : ''} onClick={() => p.onAppearanceChange(new app.Appearance({ theme: p.appearance.theme, fontScale: scale }))}>{label}</button>
            ))}
          </div>
        </div>
        {p.appearanceErr && <div id="settings-appearance-error" role="alert">{p.appearanceErr}</div>}
      </div>
      <div id="settings-list">
        {p.providers.length === 0 && <p id="settings-empty">Belum ada provider. Tambahkan endpoint di bawah.</p>}
        {p.providers.map((pr) => (
          <div key={pr.id} className={`prov${p.activeId === pr.id ? " prov-active" : ""}`}>
            <div className="prov-name">{pr.name} {p.activeId === pr.id ? "(aktif)" : ""}</div>
            <div className="prov-meta">{pr.baseURL} · {pr.activeModel || pr.model} · {pr.models.length} model · {pr.hasCredential ? "key tersimpan" : "tanpa key"}</div>
            <div className="prov-actions">
              {p.activeId !== pr.id && <button onClick={() => p.onActivate(pr.id)}>Aktifkan</button>}
              <button onClick={() => pick(pr.id)}>Ubah</button>
              <button onClick={() => p.onDelete(pr.id)}>Hapus</button>
            </div>
          </div>
        ))}
      </div>
      <div id="settings-form">
        <h4>{editing ? "Ubah provider" : "Provider baru"}</h4>
        <label>Nama<input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} placeholder="Provider pribadi" /></label>
        <label>Base URL<input value={form.baseURL} onChange={(e) => setForm({ ...form, baseURL: e.target.value })} placeholder="https://api.example.com/v1" inputMode="url" /></label>
        <label>Model
          <span className="model-add">
            <input
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
            <button type="button" onClick={addModel} disabled={!modelDraft.trim()}>Tambah</button>
          </span>
        </label>
        {form.models.length > 0 && (
          <div className="model-list">
            {form.models.map((m) => (
              <div key={m} className={`model-item${form.activeModel === m ? " model-active" : ""}`}>
                <label className="model-pick" title="Pakai model ini untuk chat">
                  <input
                    type="radio"
                    name={`active-${editing || "new"}`}
                    checked={form.activeModel === m}
                    onChange={() => setForm({ ...form, activeModel: m })}
                  />
                  <span>{m}</span>
                </label>
                <button type="button" onClick={() => removeModel(m)} title="Hapus model">✕</button>
              </div>
            ))}
          </div>
        )}
        <div className="field-hint">Radio memilih model aktif untuk chat.</div>
        <label>API key {editing && "(kosong = pakai tersimpan)"}
          <span className="keyrow">
            <input type={showKey ? "text" : "password"} value={form.apiKey} onChange={(e) => setForm({ ...form, apiKey: e.target.value })} placeholder={editing ? "•••••• (tersimpan)" : "sk-..."} autoComplete="off" />
            <button type="button" onClick={() => setShowKey(!showKey)}>{showKey ? "Hide" : "Show"}</button>
          </span>
        </label>
        {formError && <div id="settings-form-error" role="alert">{formError}</div>}
        {p.test && <div id="settings-test" role="status">{p.test.ok ? "OK: " : ""}{p.test.message}{p.test.insecure ? " (HTTP tanpa TLS!)" : ""}</div>}
        <div id="settings-form-actions">
          <button onClick={submit}>Simpan</button>
          <button onClick={() => p.onTest(toSave(form, editing))} disabled={p.testing}>{p.testing ? "Mengetes…" : "Test connection"}</button>
          {editing && <button onClick={() => { setEditing(""); setForm(EMPTY); }}>Batal</button>}
        </div>
      </div>
    </div>
  );
}
