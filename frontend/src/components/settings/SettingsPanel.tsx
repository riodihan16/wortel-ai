// SettingsPanel owns the provider list, appearance, autostart and
// global-shortcut settings. The add/edit form lives in ProviderForm: collapsed
// behind "Tambah provider" once a provider exists, opened in place under the
// card being edited. Provider delete asks inline (ADR-048 pattern) and is never
// offered on the active provider, which the backend refuses.
import { useEffect, useRef, useState } from "react";
import { app } from "../../../wailsjs/go/models";
import type { ProviderProfile, TestResult } from "../../types/provider";
import { IconAlert, IconCheck, IconChevronDown, IconPlus } from "../ui/Icon";
import { Segmented } from "../ui/Segmented";
import { ProviderForm } from "./ProviderForm";

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

function ErrorBanner({ text }: { text: string }) {
  return (
    <div className="banner banner--error" role="alert">
      <span className="banner__icon"><IconAlert size={14} aria-hidden="true" /></span>
      <span>{text}</span>
    </div>
  );
}

export function SettingsPanel(p: Props) {
  const [editing, setEditing] = useState<string>("");
  const [adding, setAdding] = useState(false);
  const [confirmDel, setConfirmDel] = useState<string | null>(null);
  // Success moment after Simpan; the first provider also gets the way back
  // to chat, since that is where the user was heading.
  const [savedNote, setSavedNote] = useState<{ name: string; first: boolean } | null>(null);
  const startChatRef = useRef<HTMLButtonElement>(null);
  const editBtns = useRef(new Map<string, HTMLButtonElement>());
  const delBtns = useRef(new Map<string, HTMLButtonElement>());
  const addBtnRef = useRef<HTMLButtonElement>(null);
  const listTitleRef = useRef<HTMLHeadingElement>(null);
  // Focus targets that only exist after the next render (a collapsed form
  // brings its trigger back); run once, then cleared.
  const pendingFocus = useRef<(() => void) | null>(null);

  useEffect(() => {
    pendingFocus.current?.();
    pendingFocus.current = null;
  });

  const firstRun = p.providers.length === 0;
  const editingProfile = p.providers.find((x) => x.id === editing) ?? null;
  const showNewForm = adding || firstRun;
  // Rows for a provider that no longer exists must not stay armed.
  const armedDel = p.providers.some((x) => x.id === confirmDel) ? confirmDel : null;

  const register = (map: Map<string, HTMLButtonElement>, id: string) => (el: HTMLButtonElement | null) => {
    if (el) map.set(id, el);
    else map.delete(id);
  };

  const startEdit = (id: string) => {
    setSavedNote(null);
    setConfirmDel(null);
    setAdding(false);
    setEditing(id);
  };

  const startAdd = () => {
    setSavedNote(null);
    setConfirmDel(null);
    setEditing("");
    setAdding(true);
  };

  const closeForm = () => {
    const id = editing;
    setEditing("");
    setAdding(false);
    pendingFocus.current = id
      ? () => editBtns.current.get(id)?.focus()
      : () => addBtnRef.current?.focus();
  };

  const handleSaved = (saved: ProviderProfile, first: boolean) => {
    setSavedNote({ name: saved.name, first });
    closeForm();
    if (first) pendingFocus.current = () => startChatRef.current?.focus();
  };

  const cancelDelete = (id: string) => {
    setConfirmDel(null);
    delBtns.current.get(id)?.focus();
  };

  const confirmDelete = (id: string) => {
    setSavedNote(null);
    setConfirmDel(null);
    if (editing === id) setEditing("");
    p.onDelete(id);
    // The card and its focused button unmount; keep focus inside the section.
    listTitleRef.current?.focus();
  };

  const formFor = (profile: ProviderProfile | null) => (
    <ProviderForm
      key={profile?.id ?? "new"}
      profile={profile}
      test={p.test}
      testing={p.testing}
      onSave={p.onSave}
      onTest={p.onTest}
      onSaved={(saved) => handleSaved(saved, firstRun)}
      onCancel={firstRun ? undefined : closeForm}
    />
  );

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
          {p.listError && <ErrorBanner text={p.listError} />}
          {savedNote && (
            <div className="banner banner--ok settings-saved" role="status">
              <span className="banner__icon"><IconCheck size={14} aria-hidden="true" /></span>
              <span className="banner__text">
                {savedNote.first ? `“${savedNote.name}” tersimpan dan aktif.` : `“${savedNote.name}” tersimpan.`}
              </span>
              {savedNote.first && (
                <button type="button" className="btn btn--sm btn--primary" ref={startChatRef} onClick={p.onClose}>
                  Mulai chat
                </button>
              )}
            </div>
          )}
          {!firstRun && (
            <div id="settings-list">
              {p.providers.map((pr) => {
                const active = p.activeId === pr.id;
                return [
                  <div key={pr.id} className={active ? "card card--active" : "card"} role="group" aria-label={`Provider ${pr.name}`}>
                    <div className="prov-head">
                      <span className="prov-name" title={pr.name}>{pr.name}</span>
                      {active && <span className="prov-badge">Aktif</span>}
                    </div>
                    <div className="prov-meta">
                      <span className="prov-meta__url" title={pr.baseURL}>{pr.baseURL}</span>
                      <span>
                        {pr.activeModel || pr.model} · {pr.models.length} model · {pr.hasCredential ? "key tersimpan" : "tanpa key"}
                      </span>
                    </div>
                    <div className="prov-actions">
                      {!active && <button type="button" className="btn btn--sm" onClick={() => p.onActivate(pr.id)} aria-label={`Aktifkan provider ${pr.name}`}>Aktifkan</button>}
                      <button
                        type="button"
                        className="btn btn--sm"
                        ref={register(editBtns.current, pr.id)}
                        onClick={() => startEdit(pr.id)}
                        aria-label={`Ubah provider ${pr.name}`}
                        aria-expanded={editing === pr.id}
                      >
                        Ubah
                      </button>
                      {!active && (
                        <button
                          type="button"
                          className="btn btn--sm btn--danger"
                          ref={register(delBtns.current, pr.id)}
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
                  </div>,
                  editingProfile?.id === pr.id ? formFor(editingProfile) : null,
                ];
              })}
            </div>
          )}
          {showNewForm
            ? formFor(null)
            : (
              <button type="button" className="btn settings-add" ref={addBtnRef} onClick={startAdd}>
                <IconPlus size={14} aria-hidden="true" />
                Tambah provider
              </button>
            )}
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
          {p.appearanceErr && <ErrorBanner text={p.appearanceErr} />}
        </section>
        <section className="section" aria-label="Umum">
          <h2 className="section__title">Umum</h2>
          <label className="check">
            <input type="checkbox" checked={p.autostart} onChange={p.onToggleAutostart} />
            Jalankan otomatis saat Windows menyala
          </label>
          {p.autostartErr && <ErrorBanner text={p.autostartErr} />}
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
          {p.shortcutErr && <ErrorBanner text={p.shortcutErr} />}
          {p.hotkeyErr && <ErrorBanner text={p.hotkeyErr} />}
        </section>
        <section className="section" aria-label="Tentang">
          <h2 className="section__title">Tentang</h2>
          <dl className="about">
            <dt>Versi</dt>
            <dd>Wortel AI {p.version || "dev"}</dd>
            <dt>Lisensi</dt>
            <dd>MIT © 2026 Ridho Dwi Ramadhan</dd>
            <dt>Font</dt>
            <dd>Geist &amp; Geist Mono, SIL Open Font License 1.1</dd>
          </dl>
          <p className="about-note">
            Data chat dan provider disimpan lokal di perangkat ini. Tidak ada
            telemetry, analytics, atau cloud sync.
          </p>
        </section>
      </div>
    </>
  );
}
