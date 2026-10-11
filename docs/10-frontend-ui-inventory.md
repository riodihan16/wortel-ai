# 10 — Inventaris UI Frontend (pra-redesign)

Status: **audit, snapshot pra-redesign.** Disusun 2026-10-10 sebagai baseline
redesign visual. **Redesign sudah dieksekusi** (ADR-043); dokumen ini sengaja
dibiarkan sebagai catatan kondisi *sebelum*, bukan deskripsi kode saat ini.
Sumber: `frontend/src/**`, `frontend/index.html`, `docs/06-ui-ux.md`,
ADR-021/027/029/030/032/033/034.

Tujuan: daftar lengkap permukaan UI yang ada, yang belum ada, dan yang harus
diubah — supaya redesign tidak melebar ke luar scope dan tidak menghapus fitur
yang sudah jalan.

---

## 1. Peta permukaan (surfaces)

Aplikasi = **satu window Wails**, ukuran berubah per mode (`App.tsx:22-23`).

| # | Surface | Root selector | Ukuran window | Sumber |
|---|---|---|---|---|
| S1 | Pet | `#pet` | 180 × 180 | `App.tsx:156-165` |
| S2 | Chat | `#chat` | 400 × 600 | `App.tsx:181-197` |
| S3 | Settings | `#chat` (reuse root) | 400 × 600 | `App.tsx:167-179` |

Belum ada: onboarding provider pertama (docs 06 §8 minta instruksi saat belum ada
provider — sekarang hanya paragraf kecil `#chat-empty`), context menu klik-kanan
karakter (docs 06 §2, ditunda), state `thinking`/`speaking` visual pada karakter
(docs 06 §3 — karakter saat ini hanya tombol `W` 96 px).

---

## 2. Inventaris region per surface

### S1 — Pet
| Region | Selector | Ada | Catatan desain |
|---|---|---|---|
| Karakter | `#pet-dot` | ✔ | Tombol bulat 96 px, teks `W`. Bukan aset karakter. |
| Hint | `#pet-hint` | ✔ | **Teks permanen** — docs 06 §2 melarang teks permanen di desktop. Harus jadi hover-only. |
| Aksi | `#pet-actions` | ✔ | Tombol `Hide`, `Quit` selalu tampil — seharusnya hover/focus-reveal. |
| Indikator status | — | ✘ | Belum ada. |

### S2 — Chat
| Region | Selector | Ada | Catatan desain |
|---|---|---|---|
| Header | `#chat-header` | ✔ | Judul + versi + 4 tombol teks (`Settings`/`Pet`/`Hide`/`Quit`) berdesakan di 400 px. |
| Baris percakapan | `#conv-bar`, `#conv-select` | ✔ | `<select>` native + 3 tombol glyph (`+`, `✎`, `🗑`). Emoji sebagai ikon = tidak konsisten lintas font. |
| Rename inline | `#conv-rename` | ✔ | Muncul di bawah conv-bar. |
| Daftar pesan | `#chat-list` | ✔ | Scroll container. |
| Empty state | `#chat-empty` | ✔ | Satu paragraf, tanpa contoh prompt (docs 06 §8 minta satu contoh prompt). |
| Bubble user | `.msg-user` | ✔ | Accent fill. |
| Bubble assistant | `.msg-assistant` | ✔ | `--surface-2` fill. |
| Status pesan | `.msg-status` | ✔ | Teks kecil di bawah bubble. |
| Error banner | `#chat-error` | ✔ | Hanya untuk error yang bukan milik balasan; bisa ditutup. Balasan gagal memakai kartu inline + `Coba lagi` (ADR-049). |
| Composer | `#chat-composer` + `textarea` | ✔ | Autosize belum ada (fixed `rows={2}`). |
| Tombol kirim/stop | `#btn-send`, `#btn-stop` | ✔ | Toggle by `busy`. |
| Footer model + status | `#chat-footer`, `#chat-footer-model`, `#chat-state` | ✔ | `#chat-footer-model` **tanpa style** (0 rule). |
| Baris lampiran | `#chat-pending`, `.pending-item` | ✔ | Thumbnail 56 px / chip nama+ukuran, tombol × per item (ADR-054). |
| Tombol lampir | `#btn-attach` | ✔ | Kiri textarea; nonaktif saat busy/tanpa provider. |
| Drop overlay | `#chat-drop` | ✔ | Hanya saat drag berkas di atas panel chat. |
| Lampiran di transkrip | `.att-list`, `.att-thumb`, `.att-chip` | ✔ | Klik thumbnail = perbesar inline, tanpa lightbox. |
| Tombol Stop saat generation berjalan di latar | — | ✘ | docs 06 §2 minta cara membatalkan eksplisit saat panel ditutup. |

### S3 — Settings
| Region | Selector | Ada | Catatan desain |
|---|---|---|---|
| Header | `#settings-header` | ✔ | Drag region + tombol `Kembali`. |
| Umum | `#settings-general` | ✔ | Autostart checkbox, shortcut enable, combo + `Ubah`. |
| Error umum | `#settings-general-error` | ✔ | |
| Shortcut | `#settings-hotkey`, `.shortcut-row`, `.hotkey-combo` | ✔ | Capture state hanya ganti border. |
| Tampilan | `#settings-appearance`, `.appearance-row`, `.seg` | ✔ | Tema gelap/terang + 3 skala font. |
| Daftar provider | `#settings-list` | ✔ | **Tanpa style** (0 rule) — bergantung `.prov` saja. |
| Empty provider | `#settings-empty` | ✔ | **Tanpa style** (0 rule). |
| Kartu provider | `.prov`, `.prov-name`, `.prov-meta`, `.prov-actions` | ✔ | `.prov-name` **tanpa style** (0 rule). |
| Form provider | `#settings-form`, `.keyrow`, `.model-add`, `.model-list`, `.model-item`, `.model-alias` | ✔ | Alias pakai lebar px keras (110 px). |
| Hasil tes koneksi | `#settings-test` | ✔ | |
| Grup Data & privacy | — | ✘ | docs 06 §2 menyebutnya; belum ada konten. |
| About/version | — | ✘ | Versi hanya di header chat (`#app-version`). |

---

## 3. Inventaris komponen (kode)

| File | Peran | Baris | Di-redesign? |
|---|---|---|---|
| `App.tsx` | Router mode + drag region + state appearance/shortcut/autostart | 197 | Struktur markup ya, logika tidak |
| `components/chat/ChatPanel.tsx` | Conv bar, list, error, composer, footer | 178 | Ya |
| `components/chat/ImagePromptBar.tsx` | Prompt gambar + model gambar, state menggambar, cancel, error | — | Ya |
| `components/chat/MessageItem.tsx` | Turn transcript; merender gambar hasil pembuatan via `/attachments/{id}` | — | Ya |
| `components/chat/ModelPicker.tsx` | Dropdown model (button + list) | 76 | Ya |
| `components/chat/SafeMarkdown.tsx` | Renderer markdown tersanitasi | 55 | Style saja (`.md`), logika tidak |
| `components/settings/SettingsPanel.tsx` | 5 grup settings + CRUD provider | 258 | Ya |
| `hooks/useChat.ts` | Reducer streaming/persist | — | **Tidak** |
| `hooks/useImageGen.ts` | State pembuatan gambar (pending/error/cancel) + event `image:event` | — | **Tidak** |
| `hooks/useProviders.ts` | CRUD provider | — | **Tidak** |

---

## 4. Inventaris token & style

| Item | Kondisi |
|---|---|
| `frontend/src/App.css` | 739 baris, ~70 blok rule, ID-based, tanpa layer |
| `frontend/src/style.css` | 30 baris reset + font stack sistem |
| Blok token | 2 (`:root` dark, `:root[data-theme="light"]`), 18 variabel masing-masing |
| Radius | **9 nilai berbeda**: 4, 6, 8, 10, 12 px + 50% — tidak ada skala |
| Spacing | Ad-hoc (`4/6/8/12/20 px`), bukan skala bernama |
| Warna di luar token | `rgba(0,0,0,.65)` (`#pet-hint`), `rgba(0,0,0,.35)` (shadow `#model-menu`), `#ffffff` literal |
| Ukuran font | `calc(<px> * var(--font-scale))` tersebar; `.md` pakai px keras tanpa scale |
| Lebar px keras | `.model-alias 110px`, `#btn-model 132px`, `#model-menu 160px` |
| `transition` / `@keyframes` | **0** di seluruh frontend |
| `:focus-visible` | **0** |
| `prefers-reduced-motion` | **0** |
| Scrollbar styling | **0** |
| Aset mati | `assets/fonts/nunito-v16-latin-regular.woff2` (19 KB) + `OFL.txt`, `assets/images/logo-universal.png` (140 KB) — **tidak direferensikan** di mana pun |

### Kelengkapan state per kontrol
| Kontrol | default | hover | focus-visible | active | disabled | loading | error | success |
|---|---|---|---|---|---|---|---|---|
| `button` generik | ✔ | ✔ (border) | ✘ | ✘ | ✔ | ✘ | ✘ | ✘ |
| `#btn-send` | ✔ | ✘ | ✘ | ✘ | ✔ (opacity .4) | ✘ | ✘ | ✘ |
| `#btn-stop` | ✔ | ✘ | ✘ | ✘ | ✘ | ✘ | ✘ | ✘ |
| `textarea` | ✔ | ✘ | ✘ | ✘ | ✘ | ✘ | ✘ | ✘ |
| `.seg button` | ✔ | ✘ | ✘ | ✘ | ✘ | ✘ | ✔ (`.seg-on`) | ✘ |
| `.model-option` | ✔ | ✔ | ✘ | ✘ | ✘ | ✘ | ✘ | ✔ (`.model-option-on`) |
| `input` (settings) | ✔ | ✘ | ✘ | ✘ | ✘ | ✘ | ✘ | ✘ |
| `#conv-select` | ✔ | ✘ | ✘ | ✘ | ✔ | ✘ | ✘ | ✘ |
| `.prov` | ✔ | ✘ | ✘ | ✘ | ✘ | ✘ | ✘ | ✔ (`.prov-active`) |

**Tidak ada satu pun kontrol dengan 8 state lengkap.**

---

## 5. Daftar kerja redesign

Legenda: **[R]** redesign visual · **[F]** perbaikan cacat · **[A]** tambahan yang
diminta docs 06 · **[K]** dipertahankan apa adanya.

### 5.1 Fondasi (harus lebih dulu)
1. **[R]** Ganti 36 deklarasi token mentah → sistem token OKLCH bernama (paper, ink, rule, accent, focus, + tint), dengan skala radius & spacing bernama.
2. **[F]** Hapus 9 radius ad-hoc → 1 skala (mis. `--radius-sm/md/lg/pill`).
3. **[F]** Ganti 3 warna literal di luar token.
4. **[R]** Tipografi: 2 famili (display + body) dari aset yang sudah ada (Nunito 400 tersedia; tambah weight/keluarga lain saat implementasi), menggantikan `system-ui`.
5. **[A]** Lapisan motion: token durasi + easing, `prefers-reduced-motion`, `transition` pada transform/opacity saja.
6. **[F]** `:focus-visible` untuk semua kontrol interaktif.
7. **[F]** Scrollbar styling agar panel 400 px tidak memakai scrollbar UA.

### 5.2 Pet
8. **[R]** `#pet-dot` → komposisi karakter/avatar berlapis (ring + glow lembut + gerak idle halus).
9. **[F]** `#pet-hint` permanen → hover-only (docs 06 §2).
10. **[R]** `#pet-actions` → muncul saat hover/focus di dalam karakter, bukan selalu tampil.
11. **[A]** Indikator status kecil (idle/thinking/error) tanpa teks permanen.

### 5.3 Chat
12. **[R]** Header: judul + versi + aksi → aksi jadi ikon (bukan 4 tombol teks), pemisah jelas antara drag area dan kontrol.
13. **[F]** Tombol glyph emoji (`+`, `✎`, `🗑`) → ikon SVG inline agar konsisten lintas font dan bisa diberi state.
14. **[R]** Conv bar → pill/tab dengan label, atau compact header row.
15. **[R]** Bubble: radius asimetris, jarak antar-turn lebih lega, meta (status) dipisah dari isi.
16. **[F]** `#chat-footer-model` tanpa style.
17. **[R]** Composer: focus ring, autosize (1–6 baris), tombol kirim bulat, hint shortcut.
18. **[F]** Empty state: tambah 1 contoh prompt (tidak auto-kirim, docs 06 §8).
19. **[R]** Error banner: struktur sebab → langkah → retry (docs 06 §8).
20. **[A]** Stop yang tetap terlihat saat panel ditutup di tengah generation.

### 5.4 Settings
21. **[R]** Struktur 5 grup → section bertajuk, jarak antar-section konsisten.
22. **[F]** `#settings-list`, `#settings-empty`, `.prov-name` tanpa style.
23. **[R]** Kartu provider → kartu ringkas dengan badge status + aksi ikon.
24. **[R]** Form provider: label di atas field, validasi inline, tidak ada `.model-alias` lebar px keras.
25. **[R]** Segmented control (tema, skala font) → indikator geser (sliding pill) yang dianimasikan.
26. **[R]** Shortcut capture: state capturing yang jelas (border + pulse, bukan hanya warna).
27. **[A]** Section About (versi + lisensi) — opsional, docs 06 §2.

### 5.5 Markdown & pesan
28. **[R]** `.md` tipografi ikut token (saat ini px keras, tidak ikut `--font-scale`).
29. **[F]** Tabel/code block: scrollbar styled, lebar tidak memaksa panel melebar.
30. **[A]** Tombol Copy pada code block (docs 06 §7 — belum ada).

### 5.6 Kebersihan
31. **[F]** Hapus aset mati (Nunito woff2 + OFL + logo PNG 140 KB) **atau** pakai Nunito sebagai font display.
32. **[F]** `docs/06 §9` token draft diperbarui ke nilai final.

### 5.7 Tidak boleh berubah (logika)
33. **[K]** Semua binding Wails, reducer streaming `useChat`, CRUD provider, persist appearance/shortcut, `--wails-draggable` opt-out di setiap kontrol dalam drag region (gotcha ADR-029 — mudah terulang).
34. **[K]** Sanitasi markdown & allowlist link `http/https` (`SafeMarkdown.tsx`).
35. **[K]** Tidak ada API key yang kembali ke UI.

---

## 6. Yang belum ada sama sekali (tidak ada kode)

- Karakter 2D sebagai aset (sekarang tombol `W`).
- Onboarding provider pertama.
- Section Data & privacy.
- Halaman About/version.
- Context menu klik-kanan.
- Ikon set (semua ikon sekarang emoji/teks).
- Motion apa pun.
- Test UI untuk composer Enter vs Shift+Enter, busy-state kontrol, render key provider
  (dicatat sebagai sisa di ADR-035).

### Status pasca-redesign (2026-10-10)

Daftar di atas adalah snapshot pra-redesign. Yang sudah punya kode sejak itu:

| Item | Status |
|---|---|
| §5.2 no. 11 — indikator status kecil pet | Selesai (ADR-044) |
| §5.3 no. 18 — contoh prompt di empty state | Selesai, dan contoh prompt hanya tampil bila provider aktif |
| §5.4 no. 27 — section About (versi + lisensi) | Selesai (ADR-044) |
| §5.6 no. 31 — hapus aset mati (Nunito + logo PNG) | Selesai (ADR-043) |
| §5.6 no. 32 — `docs/06 §9` token draft → final | Selesai (docs 06 §9) |
| §6 — karakter 2D, ikon set, motion | Selesai (ADR-043) |
| §6 — onboarding provider pertama | Selesai (ADR-044): empty state bercabang pada `hasProvider` + tombol `Buka Pengaturan` |
| §6 — halaman About/version | Selesai sebagai section `Tentang` di Settings, bukan halaman terpisah |

Masih belum ada kode: section `Data & privacy` tersendiri, context menu
klik-kanan karakter, dan test UI sisa ADR-035.

**Tidak dikerjakan meski ada di daftar kerja §5:** §5.1 no. 3 (kelengkapan state
per kontrol) — audit menemukan kontrol yang memang punya jalur state di kode
sudah lengkap; sisanya belum punya state di React sehingga menambahkannya berarti
mengubah logika, di luar batas §5.7. §5.6 no. 30 (dead CSS) selesai lewat
penghapusan `App.css` penuh.
