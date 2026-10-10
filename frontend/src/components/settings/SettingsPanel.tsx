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
}

const EMPTY = { name: "", baseURL: "https://", model: "", apiKey: "" };

function toSave(form: typeof EMPTY, id: string): SaveInput {
  return new app.SaveInput({ id, name: form.name, baseURL: form.baseURL, model: form.model, apiKey: form.apiKey, enabled: true });
}

export function SettingsPanel(p: Props) {
  const [form, setForm] = useState(EMPTY);
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
    setForm({ name: found.name, baseURL: found.baseURL, model: found.model, apiKey: "" });
    setFormError(null);
  };

  const submit = async () => {
    setFormError(null);
    if (!form.name.trim() || !form.model.trim()) {
      setFormError("Nama dan model wajib diisi.");
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
      </div>
      <div id="settings-list">
        {p.providers.length === 0 && <p id="settings-empty">Belum ada provider. Tambahkan endpoint di bawah.</p>}
        {p.providers.map((pr) => (
          <div key={pr.id} className={`prov${p.activeId === pr.id ? " prov-active" : ""}`}>
            <div className="prov-name">{pr.name} {p.activeId === pr.id ? "(aktif)" : ""}</div>
            <div className="prov-meta">{pr.baseURL} · {pr.model} · {pr.hasCredential ? "key tersimpan" : "tanpa key"}</div>
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
        <label>Model<input value={form.model} onChange={(e) => setForm({ ...form, model: e.target.value })} placeholder="example-model" /></label>
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
