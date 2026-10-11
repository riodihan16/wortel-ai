# 09 — Decision Log dan Risiko yang Belum Terjawab

Dokumen ini mencatat keputusan arsitektur agar perubahan tidak terjadi secara diam-diam. Tambah entri baru saat keputusan penting diubah.

## Status keputusan

| ID | Keputusan | Status | Alasan |
|---|---|---|---|
| ADR-001 | Gunakan Wails v2.16.0 + Go 1.26.4 + Node v24.18.0 | Dikunci 2026-10-09 | v2.16.0 rilis stabil v2 terakhir; v3 masih beta. Go/Node versi terpasang, penuhi syarat (Go 1.21+, Node 15+). |
| ADR-002 | Backend aplikasi menggunakan Go | Dipilih | Cocok untuk HTTP, streaming, concurrency, dan desktop utilities. |
| ADR-003 | Frontend React + TypeScript + Vite, styling CSS token (Tailwind dihapus) | Dipilih | Ekosistem dan pola pengembangan familier serta cocok untuk UI chat. Tailwind dihapus 2026-10-10 (ADR-027); sistem token OKLCH sekarang di `frontend/src/styles/` (ADR-043). |
| ADR-004 | Tidak menjalankan model lokal di MVP | Dipilih | Fokus pada provider API dan idle footprint. |
| ADR-005 | Satu jendela/WebView berpindah mode pet/chat | Diterima 2026-10-09 | Uji manual: drag mulus, toggle M0 pet/chat OK, Center OK, Probe OK, Quit bersih. |
| ADR-006 | SQLite untuk conversation history | Dipilih sementara | Embedded, tidak membutuhkan server database terpisah. |
| ADR-007 | API key di OS credential storage | Dipilih sebagai persyaratan | Menghindari plaintext di database/config. Implementasi konkret harus diprototipekan. |
| ADR-008 | MVP mendukung chat text + SSE saja | Dipilih | Membatasi scope dan permukaan bug. |
| ADR-011 | Baseline M1 companion | Validated 2026-10-09 | Fresh `wails init` react-ts di repo; main ~30MB CPU idle 0–0.21% (12 core); window 180x180; toggle pet 180x180 / chat 400x600 via WindowSetSize. Uji manual user: drag + toggle klik lancar. |
| ADR-041 | Pet transparan sungguhan + W bisa digeser | Done 2026-10-10 | `WindowIsTranslucent` + `usePetDrag` (ambang 4 px) — klik tidak lagi dimakan drag. |
| ADR-042 | Window sadar tepi monitor: panel tidak terpotong, drag berhenti di tepi | Done 2026-10-10 | `internal/window` clamp ke `rcWork` + subclass `WM_MOVING`; `SetWindowMode` menggantikan `WindowSetSize`. |
| ADR-043 | Redesign visual frontend: token OKLCH + karakter SVG | Done 2026-10-10 | `App.css` (739 baris, 9 radius ad-hoc, 0 `:focus-visible`/motion) diganti sistem token OKLCH di `frontend/src/styles/`; audit kontras aplikasi nyata bersih pada kontrol aktif, kedua tema. |
| ADR-044 | Pip status pet, onboarding provider pertama, section Tentang | Done 2026-10-10 | Tiga item daftar kerja docs 10 §5.2/§5.4/§6 diimplementasikan; pip memakai bentuk (bukan warna) untuk membedakan state, empty state bercabang pada `hasProvider`. |
| ADR-045 | Alias tampilan per model + Settings draggable | Done 2026-10-10 | `config.ModelEntry{ID, Alias}` + migrasi v4; `#chat` Settings jadi drag region dengan `#settings-header` sebagai handle. Semula ditulis sebagai ADR-027 (duplikat). |
| ADR-047 | Aksen UI oranye wortel, menggantikan indigo | Done 2026-10-10 | Accent dark `#f68c36` (= hue brand), light `#ac4403`; neutral/danger/success tetap; semua pasangan ≥ 4.5:1. |
| ADR-048 | Dropdown percakapan dan konfirmasi hapus tanpa UI native | Done 2026-10-10 | `<select>` + `window.confirm` digambar WebView2 di luar tema; diganti `ConversationPicker` (pola `ModelPicker`) + baris konfirmasi inline. |
| ADR-049 | Perbaikan hasil critique UI: Settings, onboarding, retry di tempat, karakter pembawa status | Done 2026-10-10 | Critique dual-agent 22/40; lima masalah prioritas dikerjakan; `RetryMessage` + `PrepareRetry` menghapus duplikasi pesan user saat retry; pip diganti wajah wortel; netral hue 280 → 70. |
| ADR-050 | Perbaikan critique ulang: form provider, shortcut, konfirmasi Keluar, wortel di chat | Done 2026-10-10 | Critique ulang 27/40; tiga P1 (jebakan ID model, hasil tes basi, shortcut menyembunyikan pet) dan dua P2 diperbaiki. |
| ADR-051 | Critique delta 32/40: fokus shortcut, salin teks biasa, error per field, pencarian percakapan | Done 2026-10-11 | Satu P2 (shortcut membuka chat tanpa fokus keyboard, dinaikkan ke P1 oleh user) dan tiga P3 diperbaiki, plus celah heuristik yang tersisa. |
| ADR-052 | Critique delta 36/40: shortcut untuk chat tak fokus, picker, alias, Tes | Done 2026-10-11 | Dua P2 (regresi picker, shortcut menciutkan chat yang tak fokus) dan dua P3 + celah H5 diperbaiki. |
| ADR-053 | Critique delta 39/40: shortcut tidak lagi membuang form Settings | Done 2026-10-11 | `toggleFront` menjaga tampilan panel yang tersembunyi/tak fokus; tiga P3 kecil. |
| ADR-054 | Lampiran gambar/PDF/teks: staging Go, BLOB SQLite, 3 turn, tanpa gating | Done 2026-10-11 | `content` array standar OpenAI; bytes di-stage di Go lalu disimpan BLOB bersama pesan; thumbnail via route `/attachments/{id}`; 413/400 modality dipetakan ke pesan khusus. |

## ADR-001 — Wails v2

**Konteks:** Wortel AI harus menjadi aplikasi Windows executable dengan backend Go dan UI modern.

**Keputusan:** Wails v2.16.0, Go 1.26.4, Node v24.18.0 (dikunci 2026-10-09). Install via `go install github.com/wailsapp/wails/v2/cmd/wails@v2.16.0`. Jangan pakai v3 (beta).

**Konsekuensi positif:** backend Go, UI web, tidak perlu memaketkan browser Chromium lengkap.

**Konsekuensi negatif/risiko:** membutuhkan WebView2 runtime; native window behavior tertentu perlu integrasi tambahan; penggunaan RAM tetap bergantung pada WebView2 dan isi UI.

**Validasi:** buat clean build Windows, dokumentasikan versi Wails/Go/Node dan cara build.

## ADR-002 — Satu jendela atau dua jendela

**Konteks:** karakter perlu mengambang, sementara chat butuh panel yang lebih besar.

**Pilihan:**
- Satu jendela yang resize/berganti mode.
- Dua jendela terpisah untuk karakter dan panel chat.

**Pilihan awal:** satu jendela.

**Alasan:** mengurangi jumlah WebView, state duplikat, serta kompleksitas sinkronisasi.

**Risiko:** jendela transparan/frameless dapat menangani hit-testing dan area transparan secara berbeda dari ekspektasi. Resize mungkin memindahkan karakter/panel dan perlu desain transisi yang baik.

**Kriteria mengubah keputusan:** bila spike M0 menemukan transparansi, drag, fokus, atau hit-testing tak dapat dibuat nyaman tanpa workaround rapuh, evaluasi pendekatan alternatif yang masih menjaga jumlah webview kecil.

## ADR-003 — Penyimpanan API key

**Konteks:** provider eksternal memerlukan credential.

**Keputusan:** credential disimpan melalui OS credential storage, sedangkan DB hanya menyimpan referensinya.

**Konsekuensi:** implementasi platform dan testing perlu lebih banyak pekerjaan daripada plaintext JSON, tetapi risiko kebocoran file berkurang.

**Validasi:** uji simpan/baca/ganti/hapus credential, akun Windows berbeda, dan error permission.

## ADR-004 — Streaming di backend atau frontend

**Pilihan awal:** streaming HTTP dilakukan di Go, kemudian chunk/event diteruskan ke React melalui bridge Wails.

**Alasan:** API key tetap di backend, cancellation dan error handling tersentralisasi, serta frontend tidak mengelola header Authorization.

**Risiko:** terlalu banyak event kecil dapat meningkatkan overhead bridge/render. Lakukan batching ringan jika pengukuran menunjukkan masalah.

## ADR-010 — Baseline M0 spike-m0 (2026-10-09)

- **Status:** Measured
- **Spike:** `../spike-m0` (di luar repo companion), Wails v2.16.0, React 19.1.0, TS 5.6.3, Vite 7.3.7, `wails build`, Windows 11 Pro 25H2 Build 26200, 12 core, WebView2 154.0.4258.62.
- **Hasil:** startup ke hwnd ~587ms; main 29.8–32.3MB CPU idle ~0.08%; tree spike 7 proses (main + 6 WebView2) total ~335.9MB; WebView2 global 792MB = semua proses sistem, bukan spike; kill bersih tanpa gantung.
- **Terbukti:** frameless transparan render, window 180x180 di posisi aman, min 120 / max 460x680 dikonfigurasi, probe ukuran/posisi via runtime OK.
- **Uji manual user 2026-10-09:** drag mulus, klik M0 buka chat, Center pindah tengah, Probe tampil ukuran/posisi, Quit keluar bersih, tombol Pet kembali ke chat.
- **Sisa risiko:** click-through pixel transparan, tepi layar, DPI 100/125/150, multi-monitor. Uji saat M1 panel nyata.
- **Keputusan:** ADR-005 DITERIMA. Satu jendela lanjut ke M1.
## ADR-012 — Uji nyata M2 kenari.id (2026-10-09)

- **Status:** Measured
- **Endpoint:** `https://kenari.id/v1`, model `muse-spark-1-3-contributor:free`.
- **Hasil:** stream 200 `text/event-stream`, delta lalu `[DONE]` cocok; client Go `OpenAICompatible.StreamChat` terima "OK GO".
- **Error nyata:** 401 key salah; 400 model tak diizinkan; 404 path salah; DNS gagal jadi error network.
- **Key uji:** sekali pakai via env, tidak disimpan di repo; file `live_test.go` dihapus seusai uji; user akan rotate key.
- **Sisa M2:** guard duplikat klik cepat belum dites; markdown aman belum ada; 429/5xx/timeout belum ketemu di uji ini.

## ADR-013 — Penutupan M2 (2026-10-09)

- **Status:** Done, kecuali markdown aman (pindah ke M3/M4 bersama sanitasi render).
- **Kirim ganda:** guard `sendingRef` sync tolak klik cepat; Stop jalan saat `sending`; terminal event reset guard.
- **Error mapping teruji via httptest:** 429/500/408/400, `Retry-After` dihormati (cap 600s), putus tengah stream pertahankan parsial, timeout jadi `PublicError` retryable.
- **Verifikasi:** `go test ./...` ok, `gofmt` bersih, `tsc` bersih, `wails build` ok.

## ADR-014 — M3 provider settings + credential storage (2026-10-09)

- **Status:** Implemented, perlu smoke manual Windows.
- **Credential:** `go-keyring v0.2.6` (Credential Manager via wincred), service `WortelAI`, ref `wortel-provider-<uuid>`. Secret tulis saat Save, ambil tepat sebelum request, wipe best-effort, tidak pernah ke frontend/log/DB.
- **Storage:** `modernc.org/sqlite v1.46.0` (pure Go, tanpa CGO), DB `%LOCALAPPDATA%\WortelAI\wortel.db`, migration `schema_version`, tabel `provider_profiles` + `app_settings` tanpa kolom secret.
- **Chat wiring:** `Service.resolve` pakai provider aktif; tanpa provider aktif kirim gagal aman `no_provider` dengan pesan Settings. Mock tetap fallback saat storage belum wired (dev/test).
- **Test connection:** non-stream `max_tokens:1`, timeout 15s, mapping 401/429/5xx reuse `MapHTTPError`, flag HTTP non-loopback sebagai insecure.
- **Delete guard:** tolak hapus provider aktif (`ErrActiveDelete`); kredensial dibersihkan setelah profil dihapus.
- **Verifikasi:** `go test ./internal/...` ok (config/credentials/storage/app baru), `tsc` bersih, `wails build` ok 15MB.
- **Sisa:** smoke manual tambah/uji/hapus provider + kirim chat live; markdown aman tetap ke M4.
## ADR-019 — M5 startup opt-in via HKCU Run (2026-10-10)

- **Status:** Done.
- **Keputusan:** startup pakai kunci `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`, nilai `WortelAI` = path exe. Tidak butuh admin, tidak menyentuh HKLM, mudah dibalik. Default **off** sampai pengguna menyalakan toggle (FR-011).
- **Perilaku:** toggle di Settings → "Jalankan otomatis saat Windows menyala". `AutostartEnabled()` baca status nyata dari registry, bukan asumsi; `SetAutostart(true)` tulis, `false` hapus nilai.
- **Bukti:** `TestRoundTrip` menyentuh registry nyata (enable → cek ada → disable → cek kosong → disable ulang idempoten), `TestEnableEmptyPath` tolak path kosong. Setelah smoke, `reg query ... Run` tidak punya entri Wortel (default off benar).
- **Risiko:** jalur exe berubah (pindah/rename) membuat entri Run basi; perlu update ulang saat pengguna menyalakan lagi.

## ADR-020 — M5 shortcut global Ctrl+Shift+W (2026-10-10)

- **Status:** Done.
- **Keputusan:** `RegisterHotKey` Win32 langsung di `internal/hotkey`, tanpa CGO. Hotkey `Ctrl+Shift+W` (agar tidak menimpa shortcut sistem umum; `MOD_NOREPEAT` untuk cegah auto-repeat). Loop pesan di goroutine khusus, dihentikan `PostThreadMessage(WM_QUIT)`. Non-Windows stub mengembalikan error.
- **Toggle:** `App.ToggleChat` menyimpan state `hidden` (Wails tak punya probe visibilitas), sembunyikan bila tampil, tampilkan + emit `app:toggle` bila tersembunyi. Tray callbacks sinkronkan state yang sama.
- **Gagal registrasi:** error dikembalikan ke frontend lewat `HotkeyError()`, tampil di Settings sebagai peringatan (FR-010). Bukan fatal — aplikasi tetap jalan.
- **Bukti:** `TestRegisterAndStop` lulus (registrasi + unregister idempoten). Probe konflik dua registrasi combo sama → "Hot key is already registered", jalur peringatan benar. Simulasi tombol via SendKeys tidak memicu hotkey OS (butuh input hardware); verifikasi hotkey nyata diserahkan ke uji manual user.
- **Sisa:** tidak ada. Rebind/disable ada di ADR-023.

## ADR-021 — M5 appearance: theme + skala font (2026-10-10)

- **Status:** Done.
- **Scope dipilih user:** theme dark/light + skala font (bukan slider ukuran window).
- **Keputusan:** default dark; override `:root[data-theme="light"]` pada blok token CSS. Semua warna hardcoded diganti variabel (`--bg`, `--surface`, `--text`, `--accent`, `--danger*`, `--ok-bg`, dst). Skala font lewat `--font-scale` → `font-size: calc(14px * var(--font-scale))` di root; opsi 0.9 / 1.0 / 1.2.
- **Persist:** key `appearance_theme` + `appearance_font_scale` di `app_settings` (nonrahasia). Validasi backend menolak theme tak dikenal dan skala di luar 0.85–1.4, jadi UI tidak bisa dibuat tak terbaca.
- **Bukti:** round-trip + penolakan nilai invalid (`appearance_test.go`). Render CSS nyata di Chromium: dark bg `#1b2636`/text `#ffffff`, light bg `#f1f5f9`/text `#0f172a`; skala 1.2 → font body 16.8px. Kontras teks/latar ≥ 15:1 (dark) dan ≥ 16:1 (light) — lolos WCAG AA (docs 06 §6).

## ADR-023 — M5 shortcut rebindable + disableable (2026-10-10)

- **Status:** Done.
- **Konteks:** FR-010 baris 3: shortcut harus dapat diubah/dinonaktifkan bila API platform mendukung. `RegisterHotKey` mendukung, jadi rebind wajib, bukan opsional.
- **Keputusan:** combo disimpan sebagai string di `app_settings` (`shortcut_enabled`, `shortcut_combo`), default `Ctrl+Shift+W` + enabled. `hotkey.Parse`/`String` menangani format ("Ctrl+Shift+W"), terima huruf/angka/F1–F12, wajib ≥1 modifier, selalu tambah `MOD_NOREPEAT`. `App.SetShortcut` validasi → persist → `applyShortcut` (Stop lalu Start) sehingga binding lama selalu dibebaskan.
- **UI:** checkbox aktif + tombol "Ubah" yang menangkap keydown berikutnya (Escape batal; modifier saja diabaikan; tanpa modifier ditolak dengan pesan). Error registrasi OS tampil sebagai peringatan dan tidak menghentikan aplikasi.
- **Keamanan:** combo rusak/ tak bisa di-parse jatuh kembali ke default (`ParseOrDefault`), jadi startup tidak pernah tanpa shortcut diam-diam.
- **Bukti:** `combo_test.go` (round-trip, case-insensitive, tolak invalid/modifierless), `shortcut_test.go` (default, round-trip, tolak unparsable tanpa merusak state, fallback combo korup), `rebind_windows_test.go` (swap combo nyata: register → stop → register combo lain; Stop sebelum start aman). Render UI: state capture tampil dengan border accent.

## ADR-027 — Hapus Tailwind CSS yang tidak terpakai (2026-10-10)

- **Status:** Done.
- **Konteks:** README keputusan awal menyebut Tailwind, tetapi audit menemukan nol utility class Tailwind di seluruh JSX; hanya `@import "tailwindcss"` di `style.css` yang dipakai sebagai reset. Plugin Vite + dua dependency aktif = beban build tanpa manfaat.
- **Keputusan:** hapus `tailwindcss` + `@tailwindcss/vite`, ganti dengan reset CSS eksplisit (`box-sizing`, margin/padding nol, font inherit) di `style.css`. Styling tetap pakai token di `App.css`.
- **Bukti:** CSS bundle 11.31 kB → 7.02 kB (gzip 2.97 → 1.76 kB). `tsc` dan `npm run build` bersih; UI terverifikasi render (pet mode tombol + header chat).

## ADR-034 — Pemilih model + status dipindah ke footer bawah chat (2026-10-10)

- **Status:** Done.
- **Konteks:** ADR-032 menaruh pemilih model di baris kedua header (atas chat). User menghendaki baris itu **di bawah chat**, bukan di atas.
- **Perubahan:** `#chat-subheader` (di App, atas `ChatPanel`) dihapus. Baris baru `#chat-footer` dirender `ChatPanel` **setelah** `#chat-composer`, dengan `ModelPicker` di kiri dan `#chat-state` di kanan (`justify-content: space-between`). Panel menerima `modelPicker: ReactNode` (provider state tetap di `useProviders`) dan `status` kembali sebagai prop.
- **Menu:** picker kini di bawah, jadi `#model-menu` diset `bottom: calc(100% + 4px); top: auto` supaya terbuka ke atas. Membukanya ke bawah akan keluar jendela 400x600. Anchor `left: 0` dipertahankan (ADR-033).
- **Drag:** `#chat-footer` berada di dalam `#chat-main`, bukan region drag, jadi `--wails-draggable: none` pada `#model-picker` dihapus.
- **Bukti:** probe browser nyata — footer di bawah composer (`belowComposer: true`, `footerBottom 592` dalam `chatBottom 600`); picker kiri (`pickerLeft 12`) + status kanan (`statusRight 388`) sebaris; menu `opensUp: true`, tanpa clipping kiri/kanan/atas/bawah, 3 opsi. `tsc`, `go test ./...`, clean release build, app start bersih.

## ADR-033 — Dropdown model terpotong di sisi kiri (2026-10-10)

- **Status:** Done.
- **Konteks:** Setelah pemilih model pindah ke `#chat-subheader` (sisi kiri), dropdown tetap terpotong.
- **Akar:** `#model-menu` memakai `right: 0` relatif `#model-picker` yang berada di kiri layar; menu lebar minimum 160px melebar ke kiri hingga `x = -16`, di luar batas `#chat` yang `overflow: hidden`, sehingga sebagian item terpotong.
- **Perbaikan:** `#model-menu` diset `left: 0; right: auto`, jadi menu melebar ke kanan dari tepi kiri picker.
- **Bukti:** probe browser nyata — sebelum: `mLeft -16`, `clippedLeft true`; sesudah: `mLeft 12`, `mRight 172`, `clippedLeft/Right/Bottom false`, 3 opsi tampil. `tsc`, `go test ./...`, clean release build, app start bersih.
- **Catatan:** saat picker pindah posisi lagi, anchor menu harus ikut disesuaikan (kiri/kanan) agar tidak keluar jendela.

## ADR-032 — Baris kedua header: pemilih model dipindah ke bawah, status sejajar (2026-10-10)

- **Status:** Done.
- **Konteks:** Setelah pemilih model muncul (ADR-031), baris pertama header terlalu padat: judul + 4 tombol + pemilih. User meminta pemilih dipindah ke bawah.
- **Perubahan:** Header dipecah dua baris. `#chat-header`: judul + kontrol (Settings, Pet, Hide, Quit). `#chat-subheader` baru: `ModelPicker` di kiri, `#chat-state` (`status: <state>`) di kanan dengan `justify-content: space-between`. `#chat-state` dipindah dari dalam `ChatPanel` ke subheader, sehingga prop `status` panel dihapus (tidak lagi dipakai).
- **Drag:** `#chat-subheader` juga region drag. `#model-picker` dan `#chat-state` diset `--wails-draggable: none`, karena custom property mewarisi dan klik pertama akan memulai drag window (gotcha ADR-029/ADR-030).
- **Bukti:** probe browser nyata (dengan stylesheet dimuat) — `subDisplay: flex`, `justifyContent: space-between`, `sameRow: true`, `subOverflow: false`; status di kanan (`statusRight 388` vs `subRight 400`); `#btn-model` `--wails-draggable: none`, klik membuka daftar 3 model; label berubah saat status `streaming`. `tsc`, `go test ./...`, clean release build, app start bersih.
- **Catatan:** probe harus me-link CSS hasil esbuild (`_s.css`); tanpa itu layout diukur tanpa stylesheet dan hasilnya menyesatkan.

## ADR-031 — `ActiveProvider` tak pernah muncul di header: bug binding Wails multi-return (2026-10-10)

- **Status:** Done.
- **Konteks:** Setelah ADR-030 (header overflow diperbaiki), pemilih model tetap tidak muncul di build nyata.
- **Akar:** `App.ActiveProvider()` mengembalikan `(config.ProviderProfile, bool)`. Binder Wails (`internal/binding/boundMethod.go`, `Call`) untuk `OutputCount()==2` memperlakukan hasil kedua sebagai **slot error**: `returnValue = callResults[0]`, dan hanya mengisi `err` bila hasil kedua bertipe `error`. Hasil kedua `bool` dibuang. Jadi frontend menerima profil saja, bukan tuple `[profile, ok]`.
- **Efek:** Hook menulis `active[1] ? active[0] : null`; `active[1]` dari struct = `undefined`, sehingga `activeProvider` **selalu null** dan `ModelPicker` tak pernah render. Settings tetap tampil karena memakai `ListProviders` (single return).
- **Perbaikan:** `ActiveProvider()` mengembalikan `*config.ProviderProfile`; `nil` -> JSON `null` saat tak ada provider aktif. Hook memetakan nilai apa adanya (`active ?? null`). Binding Wails tetap terdeklarasi `Promise<config.ProviderProfile>`.
- **Bukti:** probe hook nyata — `activeProvider.name = "Wortel AI"`, `models = [deepseek-v4-1-flash, glm-5-3, qwen3-8-flash]`; jalur `null` juga terpetakan `null`. `tsc`, `go test ./...`, clean release build, app start bersih.
- **Aturan:** binding Wails **tidak** mendukung `(T, bool)` sebagai "value + ok". Gunakan satu nilai saja (`error` sebagai return terakhir bila perlu error), atau pointer untuk membedakan "tidak ada" dari nilai kosong.

## ADR-030 — Perbaikan hasil uji manual kedua: pet hint, header overflow, percakapan baru terkunci (2026-10-10)

- **Status:** Done.
- **Konteks:** Uji manual user pada build sebelumnya menemukan tiga masalah lanjutan.
- **Bug 1 — teks pet hint tak terbaca di tema terang:** `#pet-hint` punya `background: rgba(0,0,0,0.6)` (selalu gelap) tetapi mewarisi `color: var(--text)`, yang di tema terang = `#0f172a` (gelap) → gelap di gelap. Perbaikan: `#pet-hint` diset `color: #ffffff` eksplisit, independen tema (pill selalu gelap).
- **Bug 2 — dropdown model tak muncul:** `ModelPicker` merender, tapi `#chat-header` meluber: judul + tombol model + 4 tombol kontrol melebihi lebar jendela 400px (terukur `scrollWidth 452 > clientWidth 400`), mendorong tombol `Quit` keluar. Perbaikan: judul `min-width: 0` + `text-overflow: ellipsis` agar menyusut; `.header-actions` `flex-shrink: 0`; `#btn-model` diramping (`max-width` 180 → 132px). Terukur ulang: tidak ada overflow, semua tombol di dalam 400px.
- **Bug 3 — setelah "Percakapan baru" tak bisa kembali ke history:** `newConversation` membuat id baru yang belum ada di DB, jadi id itu tidak ada di `conversations`. `<select value={conversationId}>` tanpa `<option>` yang cocok → browser memaksa select ke opsi pertama secara diam-diam tanpa event `change`; memilih opsi pertama itu lalu menghasilkan `change` value yang sama → tak ada event → tampak terkunci. Perbaikan: render `<option value={conversationId}>Percakapan baru</option>` selama id aktif tidak ada di daftar. Terverifikasi: setelah "Percakapan baru", memilih history lama memicu `onSwitch` dan memuat percakapan itu.
- **Bukti:** probe browser nyata — `ModelPicker` label/daftar/pilih benar, header tanpa overflow (`allInside: true`), option placeholder ada, switch balik ke `old-1` tercatat. `tsc`, `go test ./...`, clean release build, app start bersih.
- **Risiko:** setiap pembuatan percakapan baru harus punya opsi di select selama percakapan itu belum tersimpan; jangan hapus placeholder option tanpa mengganti mekanisme pemilihan.

## ADR-029 — Perbaikan hasil uji manual: drag inheritance, kontras, spacing, pemilih model (2026-10-10)

- **Status:** Done.
- **Konteks:** Uji manual user pada build 0.1.0 menemukan empat masalah. Tiga bug + satu permintaan fitur kecil.
- **Bug 1 — perlu klik 2x:** `#chat-header` dan `#pet` menyetel CSS custom property `--wails-draggable: drag` untuk area drag. Custom property **mewarisi** ke seluruh anak, jadi tombol di dalam header/pet ikut bernilai `drag`; `dragTest` di runtime Wails melihat `e.target` (tombol) sebagai area drag → klik pertama memulai drag window, bukan klik. Perbaikan: set `--wails-draggable: none` pada `.header-actions`, `#pet-dot`, `#pet-hint`, `#pet-actions` (nilai apa pun selain `drag` = opt-out).
- **Bug 2 — teks tak terbaca:** `style.css` menyetel `color: white` keras pada `html, body, #root` tanpa background, dan banyak kontrol (`h4`, `.prov-name`, `<button>`) tak punya warna eksplisit sehingga jatuh ke UA/default. Perbaikan: hapus `color: white`; pakai `color: var(--text)` pada root; tambah style dasar `button` berbasis token (`--surface-2`/`--text`/`--border`); tambah `color-scheme: dark`/`light` per tema. Kontras terukur (WCAG AA) semua ≥ 4.5: tema gelap terendah 8.48, tema terang terendah 5.47 setelah `--accent` terang digelapkan `#0d9488` → `#0f766e` (sebelumnya 3.74, gagal).
- **Bug 3 — spasi antar baris tinggi:** `.msg` memakai `white-space: pre-wrap`, dan react-markdown mengeluarkan newline antar elemen block; pre-wrap merender newline itu sebagai baris kosong. Perbaikan: hapus `pre-wrap` dari `.msg` (spacing blok dari margin `.md p`); pertahankan `pre-wrap` hanya pada `.msg-user` agar enter pengguna tetap tampil. Terukur: paragraf 1 baris = 18px + 6px margin (sebelumnya ~36px).
- **Fitur — tombol ganti model di header chat:** `ModelPicker` baru (tombol `Model: <aktif> ▾` + dropdown daftar model) di `#chat-header`, menggantikan `<select>` `#model-bar` yang dulu ada di dalam panel chat. Menutup saat klik luar dan Esc; memilih model aktif = no-op. Logika/backend tak berubah (`SetActiveModel`).
- **Bukti:** probe browser nyata — kontras kedua tema, spacing (`pMargin 6px`, tinggi `<p>` 18px), picker (`picked: ['p1','gpt-4o-mini']`, klik aktif no-op, klik luar menutup), property `--wails-draggable` header = `none` pada tombol. `tsc`, `go test ./...`, clean release build, app start bersih.
- **Risiko:** `--wails-draggable: none` harus dipertahankan pada tiap kontrol baru di dalam area drag; ini gotcha Wails yang mudah terulang.

## ADR-028 — Multi model per provider + markdown library (2026-10-10)

- **Status:** Done.
- **Konteks:** Dua temuan uji manual M6 oleh user: (1) respons AI tidak rapi karena `SafeMarkdown` hand-rolled menampilkan heading, blockquote, dan tabel sebagai teks literal; (2) Settings hanya mendukung satu model per provider. FR-006 semula mensyaratkan satu "Model ID"; multi model adalah penambahan scope yang disetujui user.
- **Keputusan markdown:** ganti `SafeMarkdown` hand-rolled dengan `react-markdown` + `remark-gfm` + `rehype-sanitize`. Raw HTML tidak pernah dirender (tanpa `rehype-raw`); sanitize allowlist sebagai lapis kedua; link dibatasi http/https di schema protocol + override komponen (skema non-http dirender sebagai teks). Bundle CSS+JS naik: JS 207 kB → 372 kB (gzip 65 → 116 kB).
- **Keputusan multi model:** `provider_profiles` menyimpan `models` (JSON array) + `active_model`; kolom lama `model` dipertahankan dan disinkronkan ke model aktif untuk pembaca lama. Migrasi v3 menseed baris lama (`models = json_array(model)`, `active_model = model`). Provider menyimpan daftar model; model aktif dipilih dari radio di Settings, atau dropdown di header chat. Model aktif tidak diubah per conversation (satu model aktif per provider) — percakapan tetap menyimpan `model` pada tiap pesan.
- **Bukti:** `TestMigrationV2toV3`, `TestProviderModelsRoundTrip`, `TestSaveMultiModelAndSwitch`, `TestNormalizeModels*` (dedupe, trim, cap, fallback aktif), `TestValidateSingleModelFallback`. Render UI nyata: heading/tabel/blockquote/task-list/strike/kode benar; `<img src=x onerror>` tetap teks; `[x](javascript:...)` dirender tanpa link. Settings: tambah model, tolak duplikat, hapus model memindah aktif ke entri pertama.
- **Risiko:** JSON array di satu kolom TEXT — tidak bisa di-query per model; cukup untuk MVP (daftar kecil, akses selalu lewat provider). Bundel markdown bertambah ~107 kB gzip.

## ADR-024 — M6 baseline performa + budget RAM (2026-10-10)

- **Status:** Done (terukur pada build release).
- **Konteks:** docs 08 §4 minta pengukuran RAM/CPU pada build release dan budget RAM final dari baseline M0; konsumsi WebView2 tidak boleh disembunyikan.
- **Metode:** `scripts/perf.ps1` — jalankan `build/bin/companion.exe`, tunggu window, scope proses ke pohon anak companion (companion.exe + `msedgewebview2.exe` turunannya), sampel `TotalProcessorTime` selama 30 detik idle setelah 6 detik settle.
- **Hasil (Windows 11 26200, 12 core, build re-sponsor 0.1.0, WebView2 Runtime terpasang):**
  - Startup ke window tampil: **< 250 ms**.
  - Idle CPU: **0.00%** (3 run: 0/0/0).
  - RAM companion.exe: **~33 MB**; WebView2 (6 proses): **~308 MB**; total **~341 MB** (3 run: 340.6/340.9/343.7).
- **Budget yang ditetapkan:** proses utama ≤ 60 MB; total termasuk WebView2 ≤ 450 MB; idle CPU < 1%; startup < 2 s. Semua target docs 08 terpenuhi dengan margin.
- **Catatan:** angka WebView2 bervariasi dengan versi runtime sistem; anggap total sebagai angka gabungan, bukan janji per-build. Race detector tidak tersedia (tanpa gcc/CGO).

## ADR-025 — M6 versi aplikasi + shutdown deterministik (2026-10-10)

- **Status:** Done.
- **Konteks 1 (versi):** checklist distribusi docs 08 minta "versi aplikasi ditampilkan"; sebelumnya tidak ada sumber versi di app maupun manifest.
- **Konteks 2 (shutdown):** `CancelAll` hanya memicu `cancel()` tanpa menunggu goroutine generation selesai — berpotensi menulis DB saat proses teardown (race), dan map `requests` tumbuh tanpa bound (leak).
- **Keputusan:** `var Version` di `main.go` distamp saat build (`wails build -ldflags "-X main.Version=0.1.0"`), binding `App.Version()`, ditampilkan kecil di header chat (`v0.1.0`, "dev" bila tak distamp). `wails.json` `info.productVersion` untuk metadata exe. Service menambah `sync.WaitGroup`; setiap goroutine generation `wg.Add/Done`; `CancelAll` menunggu `wg.Wait()` setelah cancel; `requests[requestID]` dihapus saat selesai.
- **Bukti:** `TestCancelAllWaitsForGoroutines` (3 stream; `CancelAll` kembali hanya setelah 3 event terminal, `active`/`requests` kosong), `TestOpenRejectsEmptyPath`/`TestOpenFailsWhenDirUncreatable`/`TestUseAfterCloseDoesNotPanic` (DB failure = error, bukan panic). `0.1.0` terverifikasi tertanam di binary. `go test ./...`, `tsc`, `wails build` bersih.

## ADR-026 — M6 redirect cross-host tidak membawa Authorization (2026-10-10)

- **Status:** Done.
- **Masalah:** client HTTP bawaan menyalin semua header (termasuk `Authorization: Bearer`) ke setiap hop redirect. Provider yang mengembalikan 3xx ke host lain → bearer token bocor ke host itu.
- **Keputusan:** `ai.SafeClient()` memasang `CheckRedirect` yang menghapus header `Authorization` bila scheme/host redirect berbeda dari request awal; same-host (trailing slash, kanonikalisasi path) tetap membawa token. Dipakai untuk stream chat dan Test Connection.
- **Bukti:** `TestSafeClientDropsAuthOnCrossHost` (redirect ke host lain → header kosong) dan `TestSafeClientKeepsAuthOnSameHost` (redirect same-host → token utuh).

## ADR-022 — M5 bundle portable + dokumentasi (2026-10-10)

- **Status:** Done, installer penuh ditunda.
- **Keputusan:** rilis MVP = executable tunggal `build/bin/companion.exe` hasil `wails build`. Aset frontend ditanam ke binary (Wails `//go:embed`), jadi tidak perlu Node/Go di mesin pengguna (NFR-006). Installer (MSI/Inno) ditunda sebagai milestone lanjutan; portable executable dinilai cukup untuk MVP.
- **Data pengguna:** `%LOCALAPPDATA%\WortelAI\wortel.db`, tidak pernah di direktori project/exe.
- **Bukti:** `frontend/dist` sementara dipindah keluar; `companion.exe` (15.5MB) tetap start ([WebView2] Environment created) — aset benar-benar tertanam. Windows 10/11 x64 didokumentasikan di README; ARM64/versi lama belum diuji.

## Pertanyaan terbuka yang perlu dijawab selama implementasi

## ADR-015 — M4 memory: kirim riwayat ke provider (2026-10-09)

- **Status:** Implemented, perlu smoke manual.
- **Masalah:** tiap `SendMessage` hanya kirim 1 pesan user terakhir; model jawab tanpa konteks sehingga "lupa" chat sebelumnya.
- **Fix:** migration v2 (`conversations` + `messages` + index), `EnsureConversation` + `AppendMessage` saat kirim, `RecentMessages(20)` oldest-first sebagai `ChatRequest.Messages`, `UpdateMessage` simpan final + status complete/cancelled/failed.
- **History UI:** binding `History(conv, n)` + load 100 pesan saat mount; restart tetap tampil.
- **Batas:** 20 pesan terakhir per request, tanpa token counting (sesuai docs 03 §8: klaim estimasi saja bila perlu).
- **Verifikasi:** `go test ./internal/...` ok, `tsc` bersih, `wails build` ok.

## ADR-016 — M2 tutup: error live terverifikasi (2026-10-10)

- **Status:** Done.
- **Bukti:** probe live `kenari.id/v1` key salah → `unauthorized/401` tepat 0.6s (probe dibuang, bukan test permanen). httptest sudah cover 400/408/429/5xx + Retry-After + putus tengah + timeout di `ce25195`.
- **Sisa M2:** tidak ada. Chat live + cancel + error mapping nyata selesai.

## Status M3 (2026-10-10)

- **FR-006 8/8:** CRUD + aktif + Test Connection + validasi URL + secret tak kembali + guard hapus aktif + ganti tak hapus riwayat (tes `TestSwitchActiveKeepsHistory`).
- **Sisa ke M4 penuh:** markdown aman, multi-conversation UI (judul/buka/hapus). Memory + persist sudah di `31bbfda`.
## ADR-017 — M4 tutup: multi-conversation + markdown aman (2026-10-10)

- **Status:** Done, perlu smoke manual Windows.
- **Backend:** `ListConversations`/`RenameConversation`/`DeleteConversation` di storage + service + binding; hapus tolak saat streaming aktif, hapus tak ada → `sql.ErrNoRows`.
- **Frontend:** `useChat` kelola daftar + `conversationId` aktif, `ChatPanel` dropdown + baru/rename/hapus (konfirmasi), `SafeMarkdown` tanpa deps: paragraf/list/code/`code`/bold/italic/link, raw HTML tak render, link allowlist http/https saja.
- **Verifikasi:** `go test ./...` ok, `tsc` bersih, `gofmt` bersih, `wails build` ok → `build/bin/companion.exe`. Tes `TestConversationCRUD` (rename/cascade/missing).
- **Sisa M4:** smoke manual Windows (baru/buka/rename/hapus + restart persist + markdown render + hapus saat streaming ditolak).
## ADR-039 — Guard cancel tidak boleh terkunci permanen (2026-10-10)

- **Status:** Done.
- **Konteks:** `useChat.cancel()` menjalankan `await CancelGeneration(...)` lalu melepas guard (`sendingRef`, `activeRequest`) **tanpa** `try/finally`. Audit pra-rilis menemukan: bila `CancelGeneration` reject (IPC/backend error), kedua guard tertinggal `true`/non-null → `busyNow()` selamanya true → tombol +, rename, delete, select, dan `send()` terkunci sampai restart aplikasi.
- **Keputusan:** bungkus dengan `try/catch/finally`; `finally` selalu reset `sendingRef.current = false`, `activeRequest.current = null`, `status = "idle"`. Error cancel dilaporkan lewat `setError`, tidak ditelan.
- **Bukti:** `useChat.test.ts` — dua tes (`cancel` gagal saat fase sending; gagal saat fase streaming dengan `started` sudah tiba) memverifikasi guard lepas dan `send()` berikutnya jalan. Dihapus sementara → 2 tes gagal (menangkap regresi).
- **Konsekuensi:** guard selaras dengan FR-005 (cancel selalu mengembalikan UI ke idle).

## ADR-038 — Kegagalan start menampilkan pesan, bukan hilang di stdout (2026-10-10)

- **Status:** Done.
- **Konteks:** `main.go` memanggil `println("Error:", ...)` bila `wails.Run` gagal. Pada build GUI tidak ada console, jadi error (termasuk WebView2 Runtime tidak terpasang) tak terlihat. FR-001 mensyaratkan "ketergantungan runtime ditangani dengan pesan yang dapat dimengerti".
- **Pilihan:** (a) biarkan; (b) `MessageBoxW` Win32 dengan pesan khusus WebView2; (c) log ke file.
- **Keputusan:** (b). `fatal_windows.go` memakai `user32!MessageBoxW` (`flags = MB_OK|MB_ICONERROR|MB_SETFOREGROUND`). Bila error memuat `webview2`/`runtime version not available`, teks berisi tautan unduhan WebView2. Di non-Windows, `fatal_other.go` mencetak ke stdout.
- **Bukti:** probe memanggil `fatalDialog` dengan error WebView2 → proses **memblokir** menunggu klik (dialog tampil); `isWebView2Missing` lulus unit test (2 positif, 2 negatif). Aplikasi normal tetap start dan keluar bersih.
- **Konsekuensi:** kegagalan start selalu terlihat user; pesan menyertakan detail error asli (tanpa secret).

## ADR-040 — Publikasi open source: lisensi MIT, CI, dan jalur signing SignPath (2026-10-10)

- **Status:** Done (belum dijalankan di GitHub — repo belum di-push).
- **Konteks:** Proyek akan dipublikasikan open source. Repo belum punya lisensi, README masih berstatus "draft perencanaan MVP", tidak ada CI, dan perilaku SmartScreen belum punya jalur penyelesaian. Distribusi publik nyata memerlukan otomatisasi build dan kebijakan penandatanganan yang jelas.
- **Pilihan lisensi:** MIT (paling permisif, umum untuk app desktop; pemilik `Ridho Dwi Ramadhan`). Alternatif Apache-2.0 (klausa patent) atau GPL-3.0 (copyleft) ditolak sebagai berlebih untuk proyek ini.
- **Pilihan signing:** (a) rilis tanpa sertifikat + catatan "Run anyway"; (b) SignPath Foundation gratis untuk OSS; (c) OV CA berbayar; (d) Azure Trusted Signing; (e) Microsoft Store (MSIX). **Keputusan: (b) sebagai jalur utama**, dengan (a) sebagai default sampai akun SignPath dan secrets ada. Alasan: SignPath Foundation memberi penandatanganan OV-level tanpa biaya untuk proyek OSS, cocok dengan kenyataan bahwa **Azure Trusted Signing membatasi individu ke AS/Kanada** sehingga kemungkinan tertutup untuk pengelola di Indonesia. EV ditolak: sejak 2024 EV tidak lagi bypass SmartScreen instan, jadi tidak sepadan biayanya.
- **Implikasi:** Kepemilikan repositori tidak boleh berubah di masa depan karena SignPath mensyaratkan nama organisasi/pemilik tetap.
- **Implementasi:** `LICENSE` (MIT), `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, `SECURITY.md`, `.gitattributes` (LF, kecuali `*.ps1`/`*.bat` CRLF), `.editorconfig`. README ditulis ulang dari draft menjadi README proyek nyata (fitur, build, distribusi, signing, lisensi). `.github/workflows/ci.yml`: gofmt, `go vet`, `go test`, `go test -race` (MinGW di `windows-latest`), typecheck + `npm test` + build frontend, lalu `wails build` + upload artifact. `.github/workflows/release.yml`: tag `v*` → build (versi dari tag → `-ldflags "-X main.Version="`) → job `sign` (environment `release`, `vars.SIGNPATH_ENABLED`, aksi `signpath/github-action-submit-signing-request@v3`, `archive: false` karena SignPath menolak wrapper ZIP) → release draft + `SHA256SUMS.txt`.
- **Batas yang diketahui:** Workflow belum pernah dieksekusi di runner GitHub (repo belum memiliki remote). Validasi hanya lokal: parse YAML, `gofmt`/`go vet`/`go test`/`go test -race`, `tsc --noEmit`/`npm test`/`npm run build`, dan asumsi (`main.Version`, `npm test` = `vitest run`) diverifikasi terhadap file repo. Job `sign` aktif hanya setelah secrets SignPath diisi dan `vars.SIGNPATH_ENABLED=true`; sampai itu, rilis mempublikasikan exe belum ditandatangani.
- **Konsekuensi:** Rilis publik pertama tetap memicu SmartScreen sampai akun SignPath lolos kualifikasi. `.gitattributes` mengubah default checkout ke LF; file yang ada sudah LF sehingga tidak ada churn.

## ADR-044 — Pip status pet, onboarding provider pertama, section Tentang (2026-10-10)

- **Status:** Done (terverifikasi pada aplikasi nyata via `wails dev`).
- **Konteks:** Tiga item daftar kerja docs 10 §5.2/§5.4/§6 belum ada kode: indikator status pet (item 11), onboarding provider pertama, dan section About/version (item 27). Sebelumnya pet tidak punya sinyal apa pun saat request berjalan, dan chat kosong menampilkan contoh prompt walau belum ada provider yang bisa menjawab — UI menjanjikan hal yang belum bisa dilakukan.
- **Keputusan pip status:** `PetSurface` menerima prop `status: ChatStatus` dan memetakan `sending`/`streaming` → `thinking`, `error` → `error`, lainnya → `idle`. Pip 14 px di `bottom: var(--space-2); right: var(--space-3)`, `role="status"` + `.sr-only`. Bentuk membawa makna, bukan warna (WCAG 1.4.1): cincin kosong vs cakram terisi vs cakram + `IconAlert`. Dipilih `bottom-right` karena hint hover menempati `top` pada lebar 180 px dan rail aksi `bottom-center`; pip di kanan bawah tidak beririsan dengan keduanya (diukur di 180×180: hint `y 4–23`, rail `x 56–124`, pip `x 154–168, y 158–172`).
- **Revisi posisi pip (2026-10-10):** uji manual: pip di sudut window terlalu jauh dari karakter dan tidak terbaca sebagai status wortel. Pip dipindah ke samping sisi kanan badan wortel: `left: calc(50% + 17px); top: calc(50% + 2px)`, offset dari pusat window tempat wortel berada, jarak ~6 px ke sisi badan. Pip juga diberi animasi `pet-float` yang sama agar jaraknya tetap saat wortel melayang. Kandidat lain yang dirender lalu ditolak: bahu kanan atas (cakram oranye `thinking` menyatu dengan badan) dan dekat ujung (terbaca seperti menetes). Tetap tidak beririsan: pip `x 107–121, y 92–106` vs hint `y 4–23` dan rail `y ≥ 134`. `pointer-events: none` tetap ada, jadi pip tidak menghalangi hit area wortel.
- **Keputusan `speaking`:** state `speaking` (docs 06 §3) tidak dibedakan dari `thinking`. Backend mengirim status chat yang sama (`streaming`) untuk fase tunggu token dan fase delta masuk, jadi membedakannya akan butuh state baru di `useChat` tanpa sinyal backend yang mendukung. Dicatat sebagai gap, bukan diimplementasikan setengah.
- **Keputusan onboarding:** `ChatPanel` menerima `hasProvider: boolean` (dari `provs.activeProvider !== null`) dan `onOpenSettings`. Tanpa provider aktif: heading “Belum ada provider” + langkah konkret + tombol `Buka Pengaturan`; **contoh prompt tidak ditampilkan**. Provider pertama otomatis menjadi aktif (`Providers.Save` sudah menyetel `ActiveProviderKey` bila kosong), sehingga tombol itu mengarah ke alur yang benar-benar selesai.
- **Keputusan About:** section `Tentang` menampilkan versi (prop `version` yang sudah dipakai `ChatHeader`), lisensi MIT + pemegang hak dari `LICENSE`, atribusi font Geist/Geist Mono (SIL OFL 1.1), dan pernyataan lokasi data. Link keluar (BrowserOpenURL) **tidak** ditambahkan: di luar item 27 dan menambah permukaan baru tanpa kebutuhan. Grup `Data & privacy` masih belum ada sebagai section tersendiri (docs 10 §6) — hanya pernyataan satu baris di sini.
- **Bukti:** `wails dev` di-drive via CDP. Pip: idle `oklch(0.56 0.022 280)` cincin kosong; saat pesan nyata dikirim → `--thinking` `oklch(0.7 0.15 280)` + sr “Sedang menulis balasan”, lalu kembali `--idle`; dengan base URL provider diarahkan ke port mati → `--error` `oklch(0.66 0.19 22)` + ikon seru, sr “Permintaan terakhir gagal”. Empty state dengan provider aktif: heading/hint/contoh prompt, 0 tombol. Empty state tanpa provider (dipaksa sementara lewat prop, lalu direvert): heading “Belum ada provider”, 0 contoh prompt, tombol 28×115 px satu baris, 0 overflow; klik tombol membuka `#settings`. Section Tentang: 3 baris + catatan, `scrollWidth == clientWidth` di 320/375/414/768, 0 elemen overflow. `tsc --noEmit` + `npm test` (14 tes) bersih.
- **Risiko:** (a) pip dirender hanya di mode pet; di mode chat tidak ada sinyal status karakter — status chat sudah dibawa indikator di footer (ADR-034). (b) `hasProvider` berasal dari daftar provider aktif; bila `ListProviders` gagal, empty state akan menampilkan onboarding walau provider ada — kegagalan itu sudah tampil sebagai error list di Settings. (c) `--color-rule-strong` untuk cincin idle adalah elemen dekoratif saat idle; makna state tetap dibawa `.sr-only` dan bentuk.

## ADR-043 — Redesign visual frontend: sistem token OKLCH + karakter SVG (2026-10-10)

- **Status:** Done (terverifikasi pada aplikasi nyata via `wails dev`).
- **Konteks:** Inventaris pra-redesign (`docs/10-frontend-ui-inventory.md`) mencatat 739 baris `App.css` berbasis ID tanpa layer, 18 token ad-hoc, **9 nilai radius berbeda**, 3 warna literal di luar token, 0 `transition`/`@keyframes`/`:focus-visible`/`prefers-reduced-motion`, 0 scrollbar styling, dan **tidak satu pun kontrol dengan 8 state lengkap**. Dijalankan dengan skill desain `hallmark`; scope mengikuti daftar kerja docs 10 §5, tanpa mengubah logika (docs 10 §5.7).
- **Keputusan styling:** `App.css` + `style.css` dihapus, diganti 9 file di `frontend/src/styles/` (`tokens` · `base` · `shell` · `controls` · `chat` · `settings` · `markdown` · `pet` · `index`). Semua warna OKLCH bernama; komponen **wajib** memakai `var(--token)`, tidak boleh nilai mentah. Radius jadi satu skala `xs/sm/md/lg/xl/pill`; spacing `space-1..7`; tipografi dua famili variabel lokal (Geist + Geist Mono, latin + latin-ext) menggantikan `system-ui`; ukuran memakai `calc(px * var(--font-scale))` sehingga setelan ukuran teks ikut menskalakan markdown (dulu `.md` px keras). Nunito + logo PNG dihapus sebagai aset mati.
- **Keputusan gerak & aksesibilitas:** token `--dur-fast/--dur/--dur-slow` + `--ease-out/--ease-in-out`, transisi hanya pada transform/opacity, `prefers-reduced-motion` mematikan semuanya, satu perlakuan global `:focus-visible` (`--focus-ring` 2 px + offset 2 px), dan scrollbar di-style agar panel 400 px tidak memakai scrollbar UA.
- **Keputusan karakter:** `#pet-dot` (tombol teks `W` 96 px) diganti `PetSurface` — SVG wortel inline (3 daun + badan + shading + ridge + mata/mulut) dengan idle float. Hint dan rail aksi jadi hover/`focus-within`-reveal murni CSS (tanpa state React), sehingga tombol tetap tabbable. Glyph emoji `+`/`✎`/`🗑` diganti set ikon SVG inline (`components/ui/Icon.tsx`).
- **Kontras (terukur di aplikasi nyata, WCAG 2.x):** rasio terburuk per token di seluruh permukaan tempat token itu dipakai — dark `ink` 12.7 · `ink-2` 6.1 · `ink-3` 4.5; light `ink` 14.4 · `ink-2` 5.8 · `ink-3` 4.5. `--color-rule-strong` (batas kontrol, butuh 3:1 per WCAG 1.4.11): 4.1 dark / 3.6 light. `--color-accent` pada `--color-paper` 6.9 dark / 5.8 light. Audit kontras menyeluruh pada aplikasi nyata: **0 kegagalan pada kontrol aktif** di kedua tema; satu-satunya temuan adalah tombol dalam keadaan `disabled` (dikecualikan WCAG 1.4.3 — komponen nonaktif). Angka di `docs/06-ui-ux.md` §9 (`ink` 16.5/8.0/5.4) adalah rasio **pada `--color-paper` saja**; rasio terburuk ada di `--color-paper-3`.
- **Bukti:** `wails dev` (backend Go nyata + runtime Wails nyata + SQLite nyata, bukan harness stub) di `http://localhost:34115`, di-drive lewat CDP dari Chromium: percakapan nyata ter-render (4 pesan, 11 `h2`, 1 tabel, 16 `ul`, 63 `strong`, 0 markdown literal); picker model menampilkan 3 model nyata dan terbuka ke atas tanpa terpotong; tema gelap/terang via segmented control dengan persist `GetAppearance()`; klik tunggal pada tombol Pengaturan dan tombol kembali; guard `Tambah` model (disabled → aktif saat field diisi); region drag Wails dipetakan ulang setelah merge dengan ADR-041: `drag` hanya pada `#chat-header`/`#settings-header` (mode chat/settings), sedangkan mode pet tidak punya region `drag` sama sekali — drag pet dipegang hook `usePetDrag`, dan `#pet` maupun setiap kontrolnya `none`/unset sehingga tidak ada drag ganda. `tsc --noEmit` dan `npm test` bersih.
- **Risiko:** (a) `--wails-draggable: none` wajib dipertahankan pada setiap kontrol baru di dalam area drag — gotcha ADR-029 yang mudah terulang; kini `.pet__actions` dan `.pet__hint` juga di-opt-out eksplisit karena keduanya non-kontrol di dalam `#pet` dan hanya tertutup oleh `pointer-events: none`. (b) Bundle font variabel menambah ~3 file woff2 ke build; diambil sebagai trade-off untuk konsistensi lintas mesin (tanpa CDN, offline-safe). (c) `--color-ink-3` hanya lolos 4.5 di `--color-paper-3` — jangan pakai token itu untuk metadata di permukaan yang lebih gelap. (d) Redesign memindahkan drag pet dari `--wails-draggable` polos ke hook `usePetDrag` yang sudah ada (ADR-041): ambang 4 px dipertahankan agar klik tunggal tidak dimakan drag (ADR-029), dan `SetWindowMode` (ADR-042) tetap dipakai, bukan `WindowSetSize`.

## ADR-042 — Window sadar tepi monitor: panel tidak terpotong, drag berhenti di tepi (2026-10-10)

- **Status:** Done; drag live di tepi perlu uji manual user.
- **Konteks:** Uji manual: pet di bawah layar → klik W → panel chat 400x600 tembus ke bawah monitor. Lalu: pet sendiri bisa digeser melewati tepi layar. Checklist release sebelumnya menyebut "logika clamp dari M0/M5", tetapi kode tidak punya clamp apa pun; `WindowSetSize` Wails selalu membesar dari pojok kiri-atas.
- **Keputusan:** package `internal/window` (Go, sesuai batas "operasi sistem di backend").
  - `SetWindowMode("pet"|"panel")` menggantikan `WindowSetSize` di frontend. Backend memegang ukuran (`PetSize`/`PanelSize`, juga dipakai `main.go`), menolak mode lain.
  - `Placer`: saat pet → panel, posisi digeser seminimal mungkin agar masuk work area monitor (taskbar dikecualikan, `GetMonitorInfo.rcWork`); geseran diingat. Saat panel → pet, geseran dibatalkan sehingga pet kembali ke tempat user menaruhnya, atau ikut offset yang sama bila panel sempat digeser. Mode sama berulang (tray "Buka" saat chat) hanya memastikan di layar, tanpa mereset geseran.
  - `KeepOnScreen`: subclass window (`SetWindowLongPtrW(GWLP_WNDPROC)`, chain `CallWindowProcW`) menangani `WM_MOVING` dan mengoreksi rect usulan ke work area **monitor di bawah kursor** (`GetCursorPos` + `MonitorFromPoint`), jadi drag berhenti di tepi, dan window ikut pindah monitor begitu kursor menyeberang (versi awal pakai `MonitorFromRect`: tepi bersama jadi dinding sampai > 50% window menyeberang). Berlaku untuk drag W dan header chat/settings.
  - State `Placer` dihitung di salinan dan baru di-commit setelah `SetWindowPos` berhasil, agar kegagalan tidak meninggalkan geseran yang tak pernah diterapkan. Frontend menunggu `SetWindowMode` selesai sebelum mengganti view.
  - HWND dicari sekali via `EnumWindows` (PID sendiri + class `WortelAIWindow`, diset eksplisit lewat `WindowClassName`). Ukuran fisik = DIP × `GetDpiForWindow` / 96 (rumus sama dengan Wails).
- **Fallback:** gagal menempatkan → `runtime.WindowSetSize` biasa (perilaku lama); gagal subclass → log warning, drag tetap jalan tanpa batas.
- **Bukti:** `go test ./internal/window/` 13 tes (Fit: muat/kanan/bawah/kiri-atas/monitor offset negatif/oversize; Placer: pet kembali, ikut panel digeser, mode berulang, clamp saat undo; Constrain; mode tak dikenal). `go vet` Windows + `GOOS=linux`, `gofmt`, `tsc`, vitest 24/24, release build OK. Probe app nyata (data dir terisolasi, tanpa mouse/keyboard user): pet diparkir di `1685,785..1910,1010` (work `0,0..1920,1020`), klik tray kiri → panel `1420,270..1920,1020` = `panelInsideWork=True`; screenshot panel utuh di atas taskbar.
- **Risiko:** subclass berada di bawah wndproc Wails; aman selama Wails tidak mengganti `GWLP_WNDPROC` sendiri (v2.16.0 tidak). `WM_MOVING` tidak bisa diuji otomatis tanpa menggerakkan mouse user.

## ADR-041 — Pet transparan sungguhan + W bisa digeser (2026-10-10)

- **Status:** Done, drag native perlu uji manual user.
- **Konteks:** Uji manual: mode pet tampil dengan kotak hitam di sekitar W, dan W hanya bisa diklik (drag hanya dari area kosong).
- **Akar masalah latar hitam:** `WebviewIsTransparent: true` hanya membuat WebView2 alpha 0, sehingga yang terlihat adalah window host. Dengan `WindowIsTranslucent: false`, window host tetap opaque dan dicat dengan class brush dari `BackgroundColour` RGB(0,0,0) (`win32.SetBackgroundColour` → `CreateSolidBrush`); alpha `BackgroundColour` diabaikan untuk brush itu. Hasilnya hitam. Konfigurasi ini sudah ada sejak M1 (`d1fc247`).
- **Keputusan transparansi:** `WindowIsTranslucent: true` (Wails menambah `WS_EX_NOREDIRECTIONBITMAP`, jadi brush tidak tampil) + `BackdropType: windows.None` (tanpa Mica/acrylic, desktop tembus apa adanya). Mode chat/settings tidak berubah karena `#chat` punya latar solid `--bg`; sudut membulatnya kini ikut transparan.
- **Akar masalah W tidak bisa digeser:** `#pet-dot` sengaja `--wails-draggable: none` (ADR-029). Bila diset `drag`, runtime Wails (`deferDragToMouseMove`) memulai drag pada `mousemove` pertama, termasuk getaran 1px, sehingga klik hilang lagi.
- **Keputusan drag:** hook `usePetDrag` di W: `pointerdown` mencatat posisi layar, `pointermove` dengan tombol utama tertekan dan jarak ≥ 4px memanggil `window.WailsInvoke('drag')` (pesan IPC yang sama yang dikirim runtime; Go → `ReleaseCapture` + `WM_NCLBUTTONDOWN/HTCAPTION`). `setPointerCapture` saat press + `pointercancel` = reset, agar lepas di luar W tidak meninggalkan origin basi. Klik setelah drag ditelan; klik keyboard (`detail 0`) tetap jalan. W tetap `--wails-draggable: none`. Hint pet: `klik W: chat · geser W: pindah`.
- **Bukti:** `usePetDrag.test.ts` 10 tes (klik, jitter < threshold, drag sekali, klik pasca-drag ditelan, press berikutnya normal, tanpa tombol, tombol kanan, capture, pointercancel, keyboard). `tsc` bersih, vitest 24/24, `gofmt`/`go vet` bersih, release build OK. Probe piksel release build (DPI-aware, data dir terisolasi): 4 sudut window 225x225 identik dengan piksel desktop sebelum launch, tengah = `(27,38,54)` = `--bg`.
- **Revisi hit area (2026-10-10, pasca ADR-043):** uji manual: drag dan klik masih aktif di area kosong sekitar wortel. Penyebabnya, area hit `.pet__btn` adalah kotak 116×116; `border-radius` tidak memotong hit test anak (`.pet__float`/`svg` tetap kotak). Selain itu `#pet` memakai `cursor: move` di seluruh window 180×180, sehingga semua area tampak bisa digeser. Perbaikan di `pet.css`: `.pet__btn { pointer-events: none }` + `path, circle { pointer-events: visiblePainted }`, jadi hanya piksel wortel yang menerima pointer. Event tetap bubble ke tombol, jadi `usePetDrag`, `:hover`, dan fokus keyboard tidak berubah. `cursor: move` di `#pet` dihapus. Bukti (Edge headless + CDP, CSS asli): `elementFromPoint` di sudut kotak, sela daun, dan sisi ujung wortel → `#pet` (sebelumnya `svg`); press di badan wortel lalu geser ke (10,10) → `setPointerCapture` sukses pada tombol meski `pointer-events: none`, `pointermove` tetap diterima; press/geser dari sudut kotak → 0 event.
- **Risiko:** `window.WailsInvoke` adalah API internal runtime Wails v2 (bukan bagian `wailsjs/runtime`); aman selama versi dikunci v2.16.0, cek ulang saat upgrade. Pada Windows < 22621 (`SupportsBackdropTypes` false) Wails memakai `SetTranslucentBackground` (blur accent) sehingga latar pet bisa tampak buram, bukan bening. Area transparan tetap menangkap klik (bounds window persegi, sesuai docs 03 §4).

## ADR-037 — Persistensi percakapan best-effort, bukan blocker stream (2026-10-10)

- **Status:** Done.
- **Konteks:** `internal/app/service.go` menulis conversation + pesan (`EnsureConversation`, `AppendMessage`, `UpdateMessage`) di dalam goroutine streaming dan mengabaikan error (`_ =`). Audit pra-rilis menanyakan apakah error ini harus menggagalkan request.
- **Pilihan:** (a) propagasi error tulis → batalkan stream + emit `failed`; (b) best-effort, abaikan, andalkan user melihat history kosong; (c) retry dengan backoff.
- **Keputusan:** (b). Stream ke provider adalah operasi jaringan panjang; kegagalan tulis lokal (disk penuh, handle hilang) tidak boleh membuang respons yang sudah diterima user. MVP tidak punya retry queue; (c) ditunda.
- **Batas yang diketahui:** bila penulisan gagal, pesan hilang saat restart tanpa notifikasi. `EnsureConversation` pada DB tertutup sudah dipastikan mengembalikan error, bukan panic (`TestEnsureConversationFailsAfterClose`), sehingga layanan tetap aman. Kondisi `resolve` sukses **dan** tulis gagal (mis. disk penuh tepat di tengah) sulit direproduksi dalam unit test tanpa hook test-only di produksi; karenanya tidak ada tes otomatis untuk jalur emisi, hanya jaminan tidak-panic.
- **Konsekuensi:** error persistensi tidak terlihat user pada 0.1.0. Kandidat perbaikan M7: surface `storage` error via `ChatEvent` bila sebuah flag diset, atau retry queue.

## ADR-036 — M6 hardening: race detector aktif (MinGW-w64) (2026-10-10)

- **Status:** Done.
- **Konteks:** `go test -race` butuh CGO + compiler C; mesin build sebelumnya tanpa gcc, sehingga kebersihan konkurensi hanya dibuktikan manual. Service goroutine-heavy (streaming, cancel, `CancelAll`, tray, hotkey).
- **Pilihan:** (a) biarkan `-race` tak dijalankan, (b) pasang MinGW-w64 permanen, (c) runner CI terpisah.
- **Keputusan:** (b). `winget install BrechtSanders.WinLibs.POSIX.UCRT` (gcc 16.2.0, UCRT). PATH user menunjuk ke `…\WinGet\Packages\…\mingw64\bin`, dikelola winget (bisa uninstall/upgrade). Rilis **tetap** `CGO_ENABLED=0` — artefak tidak memakai CGO, jadi compiler hanya dibutuhkan untuk verifikasi dev.
- **Temuan:** `-race` menemukan 2 data race, keduanya di **kode test** (bukan produksi): emitter `events = append(...)` di `TestSendAndCancel` dan `TestCancelAllStopsStreams` ditulis dari goroutine stream layanan lalu dibaca loop test tanpa sinkronisasi. Diperbaiki dengan `eventSink` ber-mutex (`emit`/`terminal`/`snapshot`).
- **Bukti:** `CGO_ENABLED=1 go test -race -count=1 ./...` hijau (8 paket). Detektor dibuktikan berfungsi: menghapus mutex di `eventSink.emit` → race terdeteksi (`DATA RACE`, rc=1), restore → hijau. `go vet` bersih, `gofmt -l` kosong.
- **Konsekuensi:** `-race` masuk gate backend. Toolchain mesin berubah (gcc permanen); didokumentasikan di README + known issues.

## ADR-035 — M6 hardening: regression suite frontend (2026-10-10)

- **Status:** Done.
- **Konteks:** Known issue lama: "UI frontend belum punya test runner permanen"; regresi UI hanya terjaga oleh probe browser sekali pakai. Logika paling rawan adalah reducer event streaming di hook (matching `conversationId`, adopsi `requestId` dari event `started`, guard kirim ganda) dan kebijakan link markdown (input tak tepercaya).
- **Pilihan:** (a) probe browser sekali pakai (status quo), (b) `vitest`+`jsdom` dengan binding Wails di-mock, (c) tambah runner E2E (Playwright) atas binary Wails.
- **Keputusan:** (b). Paling kecil yang menjangkau logika berisiko tanpa binary/WebView2. `frontend/vitest.config.ts` terpisah dari `vite.config.ts` supaya build app tak menarik plugin test. `src/test/wails.ts` mengganti `wailsjs/runtime` (bus event) dengan emitter in-memory; `wailsjs/go/main/App` di-`vi.mock` per file test. Assertion menargetkan invariant perilaku (0 elemen berbahaya, link jahat jadi teks), bukan snapshot.
- **Bukti:** 12 tes lolos (`npm test`). Deteksi regresi diverifikasi dua arah: menghapus guard `requestId` di `useChat.ts` → 1 tes gagal; menetralkan guard link `SafeMarkdown` → 2 tes gagal; keduanya kembali hijau setelah restore. `tsc --noEmit` bersih; `npm run build` sukses; bundle produksi tidak memuat kode test.
- **Konsekuensi:** `npm test` masuk gate. E2E atas binary Wails tetap manual.
- **Sisa:** Enter vs Shift+Enter, busy state kontrol, dan render provider key belum otomatis.

## ADR-045 — Alias tampilan per model + Settings draggable (2026-10-10)

> Catatan: entri ini semula ditulis sebagai ADR-027, ID yang sudah dipakai
> "Hapus Tailwind CSS" (di atas). Dinomori ulang 2026-10-10 tanpa mengubah isi.

- **Status:** Done, perlu smoke manual Windows.
- **Konteks:** Nama Model ID sering panjang/teknis; user ingin label singkat tanpa mengubah ID yang dikirim ke endpoint. Juga: window Settings tidak bisa dipindah (hanya pet/chat yang bisa).
- **Pilihan:** (a) ubah ID jadi label, (b) field alias terpisah per entri model.
- **Keputusan:** (b). `config.ModelEntry{ID, Alias}` menggantikan `[]string` di input/profile/row, storage, dan binding. Alias kosong → UI pakai ID. ID tetap identitas dan tetap dikirim ke API. Migrasi v4 (idempoten) mengubah `models` JSON `["a","b"]` → `[{"id":"a","alias":""},...]`; baris objek tak disentuh; `active_model` dipertahankan.
- **Drag Settings:** `#chat` root Settings = drag region, `#settings` scroll area di-opt-out (`--wails-draggable: none`), `#settings-header` jadi handle (`cursor: move`), tombol/input di header opt-out agar klik pertama kena kontrol. Pola sama dengan header chat/pet.
- **Verifikasi:** `go test ./...` ok, `go vet` bersih, `gofmt` bersih, `tsc --noEmit` bersih, `wails build` ok → `build/bin/companion.exe`. Probe browser: picker tampil alias + fallback ID, `onModelChange` kirim ID (bukan alias); Settings root/header `drag`, scroll `none`, tombol `none`. Migrasi diuji pada salinan DB v3 user → v4, array string jadi entri objek, `active_model` utuh. Tes baru: `TestNormalizeModelsAliases`, `TestMigrationV3toV4`.
- **Sisa:** smoke manual Windows (rename alias lalu Simpan → picker chat menampilkan alias; drag Settings; DB user ter-migrasi saat buka app).

## ADR-046 — Metadata exe: string `FileVersion` wajib ada (2026-10-10)

- **Status:** Done.
- **Konteks:** Uji pra-rilis: Explorer menampilkan Company/Copyright/File version dengan benar, tetapi `System.Diagnostics.FileVersionInfo` (PowerShell `(Get-Item x.exe).VersionInfo`, .NET, dan tool yang memakainya) mengembalikan `CompanyName`/`ProductName`/`FileDescription`/`LegalCopyright` **kosong**; `FileVersionRaw`/`ProductVersionRaw` justru terisi.
- **Investigasi:** `version.dll` (`VerQueryValue`) membaca `\StringFileInfo\040904b0\CompanyName` dengan benar pada exe yang sama; dua jalur Win32 tidak sepakat. Diisolasi dengan exe probe minimal dari `winres` langsung (`/tmp/wrprobe`): tabel string tanpa key `FileVersion` → .NET mengosongkan **semua** field string; menambahkan `FileVersion` → semua terisi. Jadi .NET memperlakukan absennya key `FileVersion` sebagai tabel tidak valid dan membuang seluruh blok, sementara Win32/Explorer hanya mengabaikannya.
- **Pilihan:** (a) abaikan karena Explorer sudah benar, (b) tambahkan key string `FileVersion` ke tabel `0409`.
- **Keputusan:** (b). `build/windows/info.json` `info.0409` ditambah `"FileVersion": "{{.Info.ProductVersion}}"`. Ini juga menutup kesenjangan dengan `git.exe` (punya `FileVersion` string) dan membuat properti exe konsisten di semua pembaca (installer, signature tooling, AV heuristics yang memakai .NET).
- **Bukti:** probe `c_withfv.exe` (dengan `FileVersion`) → `Co='Wortel AI' Prod='Wortel AI' Desc='Wortel AI' FV='0.1.0'`; probe `d_origorder.exe` (tanpa) → string kosong, `FVraw=0.1.0.0`. Setelah `wails build` ulang, `companion.exe` → `Co='Wortel AI' Prod='Wortel AI' Desc='Wortel AI' Copy='MIT' FV='0.1.0' PV='0.1.0'`. `scripts/smoke.ps1` → `SMOKE_OK` (window "Wortel AI", 35 MB).
- **Risiko:** `info.0409` masih `0409` (en-US) hardcoded lewat template Wails; UI berbahasa Indonesia tetapi metadata tetap en-US. Bukan blocker rilis.

## ADR-047 — Aksen UI oranye wortel, menggantikan indigo (2026-10-10)

- **Status:** Done.
- **Konteks:** Permintaan user: aksen UI harus oranye seperti wortel. Aksen ADR-043 adalah indigo (hue 280), satu keluarga dengan neutral, sehingga UI tidak membawa identitas produk; oranye hanya muncul di karakter (`--color-brand`).
- **Keputusan:** hanya `--color-accent`/`--color-accent-soft` yang berubah; neutral Iris (hue 280), `danger`, `success`, dan `brand` tetap. Dark: `oklch(74% 0.16 55)` (`#f68c36`), hue dan lightness sama dengan `--color-brand` dark; soft `oklch(28% 0.05 55)`. Light: `oklch(52% 0.15 44)` (`#ac4403`); wortel terang gagal 4.5:1 di kertas terang, jadi lightness diturunkan dan hue digeser ke merah agar tetap oranye, bukan cokelat. Soft `oklch(95.5% 0.028 62)`. Semua nilai dicek di dalam gamut sRGB.
- **Kontras (dihitung OKLCH → sRGB, WCAG 2.x):** dark: accent pada `paper` 7.9, terburuk `paper-3` 6.1, pada `accent-soft` 6.1, `accent-ink` pada accent 7.9. Light: accent pada `paper` 5.6, terburuk `paper-3` 4.7, pada `accent-soft` 5.1, `accent-ink` pada accent 5.6. Angka accent di ADR-043 digantikan angka ini.
- **Risiko:** accent oranye dan `danger` (hue 22) sekarang sama-sama hangat (beda hue ~25–33°). Keduanya tetap terbedakan (dark `#f68c36` vs `#f0555b`, light `#ac4403` vs `#be2132`), dan state error sudah membawa ikon/teks (ADR-044: pip error = cakram + `IconAlert`; status dot busy berdenyut). Jangan pakai warna sebagai satu-satunya pembeda accent vs danger pada komponen baru.

## ADR-048 — Dropdown percakapan dan konfirmasi hapus tanpa UI native (2026-10-10)

- **Status:** Done.
- **Konteks:** Uji manual: popup konfirmasi hapus percakapan dan dropdown percakapan tampak rusak. Keduanya memakai UI native: `window.confirm()` di `ChatPanel` dan `<select id="conv-select">`. WebView2 menggambar dialog confirm dan popup opsi `<select>` dengan chrome sistem (judul origin, tombol OK/Cancel sistem, daftar opsi sistem) di luar token tema, sehingga tidak ikut dark/light maupun aksen.
- **Keputusan dropdown:** `ConversationPicker` baru (`components/chat/ConversationPicker.tsx`), kontraknya sama dengan `ModelPicker` (ADR-033): tombol trigger + `role="menu"` berisi `menuitemradio`, tutup dengan Escape/klik di luar tanpa handler global saat tertutup, terbuka ke bawah selebar trigger, `max-height: 50vh` + scroll. Tambahan: fokus ke item aktif saat dibuka, ArrowUp/Down/Home/End berpindah item, fokus kembali ke trigger setelah memilih atau Escape, dan menu ditutup otomatis saat `busy`. Percakapan baru yang belum tersimpan tetap muncul sebagai entri aktif "Percakapan baru" (mempertahankan perbaikan Bug 3 untuk `<select>`). Chrome menu dipindah dari `.model-menu`/`.model-option` ke primitive bersama `.menu`/`.menu-option` di `controls.css`; masing-masing picker hanya mengatur posisi.
- **Keputusan konfirmasi:** baris inline `#conv-delete` di bawah `#conv-bar`, bukan modal: menghapus satu percakapan cukup butuh tekan kedua yang disengaja, tanpa perlu focus trap. Baris ini memakai latar `danger-soft` + border `danger` + `IconAlert`, teks "Hapus permanen “judul”?" dengan judul terpotong ellipsis, serta tombol `Hapus` (danger) dan `Batal`. `Batal` mendapat fokus awal, jadi Enter tidak menghapus. Escape membatalkan dan mengembalikan fokus ke tombol tong sampah. Baris rename dan baris hapus saling meniadakan, dan baris hapus tertutup saat percakapan aktif berganti agar tidak menghapus percakapan yang tidak sedang dilihat.
- **Bukti:** selama pengembangan, 10 tes komponen (test-first) mencakup: tidak ada `combobox`; trigger berlabel judul aktif; item aktif `aria-checked` + fokus; pilih item lain → `onSwitch`; item aktif tidak re-switch; panah + Escape; Tab keluar menutup menu; entri percakapan baru; `window.confirm` tidak dipanggil; fokus awal `Batal`; `Hapus` → `onDelete` lalu fokus ke composer; `Batal`/Escape tidak menghapus. Tes konfirmasi gagal sebelum implementasi. File tes itu (`ChatPanel.test.tsx`) **tidak di-commit** atas keputusan user, jadi perilaku ini belum dijaga regresi otomatis. `tsc --noEmit` bersih, `npm test` 24/24 (suite yang di-commit). Render visual memakai `ChatPanel` + CSS asli (esbuild, Edge headless) di dark/light, untuk menu terbuka dan baris konfirmasi. Interaksi nyata via CDP (mouse/keyboard sungguhan): klik trigger → terbuka + fokus item aktif; tekan padding menu → tetap terbuka (menu `tabIndex=-1`, jadi fokus tidak jatuh ke `<body>`); klik di luar → tertutup; Tab berpindah antar-item lalu keluar → tertutup, fokus ke tombol `+`; pilih item → tertutup, fokus kembali ke trigger.
- **Review (cavecrew-reviewer):** diterapkan: fokus ke composer setelah hapus, baris hapus dirender dari `armed = deleting === conversationId` (bukan hanya `deleting`) agar tidak sempat satu render menampilkan judul percakapan lain, menu tutup saat fokus keluar (`onBlur` + `relatedTarget`). Tidak diterapkan: baris hapus tetap terbuka bila klik di tempat lain (konsisten dengan baris rename; ditutup oleh Batal/Escape/ganti percakapan), dan refocus saat menu ditutup oleh `busy` (trigger juga disabled, jadi tidak ada target fokus). `ModelPicker` punya celah Tab-keluar yang sama; belum diubah.
- **Catatan test:** vitest berjalan tanpa `globals`, jadi Testing Library tidak memasang `cleanup` otomatis; test komponen yang ditambahkan nanti wajib `afterEach(cleanup)`.

## ADR-049 — Perbaikan hasil critique UI (2026-10-10)

- **Status:** Done, perlu smoke manual Windows.
- **Konteks:** `/impeccable critique` dual-agent atas `frontend/src` (HEAD `3090e83`) memberi skor 22/40 dengan lima masalah prioritas: (P1) tombol Stop nyaris tak terlihat saat streaming; (P1) area provider di Settings rusak (badge `Aktif` tidak pernah muncul setelah reload karena `refresh()` tidak mengisi `activeId`, Hapus provider tanpa konfirmasi, form/list/Tentang tanpa CSS layout, Ubah memuat form di luar layar); (P1) first run penuh jargon, composer aktif tanpa provider; (P2) karakter tidak membawa status dan ikut warna tema; (P2) giliran gagal/batal berantakan (bubble kosong, "gagal" 3×, `Coba lagi` menambah bubble user baru, teks status 10 px). User memilih: Settings + first run dulu, karakter lebih hadir, kerjakan kelima issue.
- **Keputusan Settings:** `useProviders.refresh()` menyalin `ActiveProvider()` ke `activeId`; status aktif selalu mengikuti backend. Form diekstrak ke `ProviderForm` (di-key per target, jadi reset tanpa effect), tersembunyi di balik `+ Tambah provider` bila sudah ada provider, dan terbuka di bawah kartu yang di-`Ubah`. Hapus provider memakai `.confirm-row` (pola ADR-048) dan tidak ditawarkan untuk provider aktif (backend menolak, `ErrActiveDelete`).
- **Keputusan onboarding:** preset OpenAI, OpenRouter, Groq, LM Studio, Ollama hanya mengisi nama + Base URL; nama model hanya contoh di placeholder karena katalog berubah. Keterangan penyimpanan key (Windows Credential Manager, diverifikasi di `internal/credentials/osstore.go`). `Tes koneksi` di depan `Simpan`; simpan pertama → banner "tersimpan dan aktif" + `Mulai chat` (fokus). Composer + kirim nonaktif tanpa provider. Empty state memakai wortel; contoh prompt mengisi composer, tidak mengirim.
- **Keputusan retry:** backend `RetryMessage(conversationID, text)` (service + binding `wailsjs`). `storage.PrepareRetry` membaca dua baris terbaru (`ORDER BY created_at DESC, rowid DESC`, karena pesan user dan balasannya ditulis di detik yang sama): bila `[user(text), assistant failed]`, balasan gagal dihapus dan baris user dipakai ulang; bila baris terbaru user(text), dipakai ulang; selain itu tidak mengubah apa pun dan generate menyimpan pesan user seperti biasa (mis. percobaan pertama gagal sebelum menyentuh DB). Frontend `useChat.retry()` membuang bubble gagal lokal dan memanggil `RetryMessage`, tanpa bubble user baru.
- **Keputusan status chat:** balasan gagal = kartu inline (alasan + `Coba lagi` hanya pada giliran terbaru, `role="alert"`); banner `#chat-error` hanya untuk error yang bukan milik balasan dan bisa ditutup (`clearError`). Batal = penanda "— Dihentikan" di dalam balasan. Stop = `.btn--stop` tinta terisi 32 px, Escape di composer menghentikan. Footer hanya bicara selama sibuk. Transkrip `role="log"` + `aria-busy`. `ModelPicker` mengikuti kontrak keyboard `ConversationPicker` (helper `moveMenuFocus`). Rename/hapus nonaktif untuk percakapan yang belum tersimpan.
- **Keputusan tipe:** `--text-2xs` (10 px, 9 px pada "Kecil") dihapus; `--text-xs`/`--text-sm` memakai `max(11px, …)`. Hint field naik ke 12 px; judul section 16 px di atas judul form 12 px. Pill versi pindah dari header ke Tentang.
- **Keputusan karakter:** komponen `Carrot` (mood `idle`/`thinking`/`error` lewat wajah dan daun) dipakai pet, empty state, dan avatar setiap balasan. Pip status ADR-044 dihapus; `thinking` = gelembung "…", `error` = badge "!". Palet tetap `--carrot-*` + outline gelap, tidak ikut tema (menggantikan `--color-brand*`). Balasan tanpa bubble di samping avatar; bubble user netral (`paper-3`), jadi oranye milik companion, bukan gema user. Netral pindah dari hue 280 ke hue 70 dengan chroma ×0.6 ("Tanah").
- **Keputusan motion:** float idle tak berujung dihapus (jendela always-on-top tidak boleh makan biaya saat diam); diganti kedip ~120 ms tiap 6 detik di pet. Mulai balasan: lompatan squash-and-rise + gelembung pop; gagal: gelengan singkat. Panel chat masuk dengan fade + naik 8 px saat datang dari pet. Reduced motion mematikan semuanya kecuali denyut opacity titik "…".
- **Kontras (OKLCH → sRGB, WCAG 2.x, kasus terburuk di keempat permukaan):** netral hue 70 vs 280 praktis sama: dark ink 12.70 / ink-2 6.14 / ink-3 4.72 / rule-strong 3.15 / accent 6.09; light ink 14.36 / ink-2 5.84 / ink-3 4.51 / rule-strong 3.04 / accent 4.71.
- **Bukti:** test-first di tiap langkah. Vitest 75/75 (hook test di-commit: `useProviders.test.ts`, tambahan `useChat.test.ts`; test komponen `SettingsPanel.test.tsx`, `ChatPanel.test.tsx`, `PetSurface.test.tsx` **tidak di-commit**, mengikuti keputusan user untuk `ChatPanel.test.tsx` di ADR-048). Go: `TestPrepareRetry*` (DB `t.TempDir()`), `TestRetryStreamsLikeSend`; `go vet`, `gofmt`, `go test ./internal/... .` bersih. `tsc --noEmit` dan `npm run build` bersih. Visual: harness mock-backend (Vite + shim Wails, di `node_modules`, tidak di-repo) di-drive lewat CDP Edge headless, kedua tema, skala teks 0.9/1.2, state interaksi (form edit, konfirmasi hapus, streaming, gagal, menu terbuka, fokus keyboard). Detector runtime sebelum → sesudah: `undersized-ui-text` ×3, `tiny-text` ×2, `flat-type-hierarchy`, dan `nested-cards` pada bubble balasan hilang; tersisa `overused-font` (Geist, ADR-043), `nested-cards` di Settings (false positive: shell jendela `#chat` dibaca sebagai kartu), dan `monotonous-spacing` (advisory: padding mikro 2–5 px pada titik/badge; ritme layout memakai skala 8/12/16/20).
- **Review (cavecrew-reviewer, 3🔴 9🟡 2🔵), diterapkan dengan tes lebih dulu:** (1) kegagalan cepat bisa mengirim `started`+`failed` sebelum binding mengembalikan `requestId`, lalu `begin()` memasang lagi request yang sudah selesai sehingga `send()` terkunci diam-diam; kini id yang selesai dicatat (`settled`) dan tidak dipasang ulang. Bug ini sudah ada sebelum ADR ini. (2) Event `cancelled` dari backend datang setelah guard dilepas dan terbuang, sehingga balasan yang dihentikan tetap "streaming" (titik "…" selamanya); kini `cancel()` menandai balasan itu `cancelled` secara lokal. (3) Alasan gagal disimpan di pesannya (`errorText`), bukan di `error` global: error lain (ganti judul, hapus) tetap tampil di banner walau balasan terakhir gagal, dan kartu gagal tidak lagi bisa menampilkan alasan yang salah. (4) `retry()` membuang bubble gagal hanya setelah backend menerima retry. (5) `Simpan` punya guard + state `Menyimpan…` (klik ganda tidak lagi membuat dua profil). (6) Baris hapus provider ter-disarm saat provider itu aktif / `Aktifkan` diklik. (7) `PrepareRetry` sekarang satu transaksi, DELETE dibatasi `conversation_id`+`role`+`status`, dan balasan `streaming` sisa crash diperlakukan seperti gagal; rekonsiliasi retry berjalan sebelum resolve provider, dan error DB tidak lagi menggandakan baris user. Tidak diterapkan: jalur `SendMessage` yang gagal resolve (mis. belum ada provider) tetap tidak menyimpan pesan user, perilaku lama di luar cakupan ini.
- **Risiko:** (a) `RetryMessage` belum diuji end-to-end dengan provider nyata + DB nyata; jalur DB dicakup unit test storage. (b) Binding `wailsjs` diedit tangan mengikuti format generator; `wails build` berikutnya meregenerasinya dan hasilnya harus sama. (c) Preset bisa usang bila provider mengganti endpoint; hanya Base URL yang diisi dan tetap bisa diedit. (d) Animasi CSS di WebView2 yang tersembunyi bisa tertunda sampai jendela terlihat; tidak memengaruhi state akhir.

## ADR-050 — Perbaikan dari critique ulang (2026-10-10)

- **Status:** Done, perlu smoke manual Windows.
- **Konteks:** critique ulang setelah ADR-049 memberi 27/40 (naik dari 22) dengan tiga P1, ketiganya perilaku lama: (1) ID model yang diketik tetapi tidak di-`Tambah` ditolak saat Simpan; (2) satu state hasil `Tes koneksi` global tidak pernah dibersihkan, sehingga provider sehat bisa menampilkan 401 palsu; (3) `ToggleChat` menyembunyikan jendela setiap kali terlihat, jadi shortcut "buka chat" di mode pet membuat companion hilang. Dua P2: wortel bisu saat menunggu di chat dan saat balasan selesai di mode pet; tidak bisa menyalin seluruh balasan dan transkrip ditarik ke bawah di setiap delta. User memilih: kerjakan semua, `Keluar` diberi konfirmasi, test komponen tetap tidak di-commit.
- **Keputusan form:** `commitDraft()` memasukkan draft ID model ke daftar saat Simpan/Tes. Field yang menggagalkan simpan diberi `aria-invalid`, fokus, dan `aria-describedby` ke banner error. Hasil tes dibersihkan saat form dibuka/ditutup atau URL/key/model berubah (`onDirty`), dan `useProviders` memakai nomor urut sehingga tes yang selesai setelah dibersihkan dibuang. Preset menampilkan satu baris sumber key + ID model.
- **Keputusan shortcut:** `toggleFor(hidden, mode)` (unit test): tersembunyi atau pet → tampil + chat (`app:toggle`); panel terbuka → pet (`app:pet`). `App` menyimpan nilai `SetWindowMode` terakhir. Menyembunyikan tetap hanya lewat aksi `Sembunyikan`. Hint pet: "Klik atau <combo>: chat" (satu baris supaya tidak menutup daun).
- **Keputusan Keluar:** konfirmasi inline di bawah header chat (`.confirm-row`) dan di dalam rail pet; Batal difokus, Escape membatalkan dan mengembalikan fokus ke tombol power.
- **Keputusan karakter:** footer chat menampilkan wortel 26 px yang berpikir (melompat masuk) selama balasan berjalan, menggantikan status dot berdenyut. `usePetSignal`: balasan yang selesai saat user di pet memunculkan gelembung "Sudah!" sampai chat dibuka; error yang sudah dilihat di chat tidak lagi membuat pet cemas.
- **Keputusan transkrip:** `Salin` di bawah balasan selesai (sumber markdown), hook `useCopy` bersama dengan status "Gagal menyalin". Auto-scroll hanya bila user berada ≤ 48 px dari bawah; selain itu chip "Balasan baru". Composer terkunci tetap terbaca (placeholder ink-2, border putus-putus); tombol kirim disabled berupa outline tenang; `--leading-read` 1.5 untuk transkrip; banner Settings tidak lagi memotong kata.
- **Bukti:** test-first per langkah. Vitest 109/109 (hook test di-commit: `useProviders`, `usePetSignal`, `useCopy`; test komponen tetap untracked sesuai keputusan user). Go `TestToggleFor` + seluruh `go test ./internal/... .` bersih. Visual: harness mock-backend lewat CDP Edge headless (pet "Sudah!", konfirmasi Keluar di rail dan header, chip "Balasan baru", wortel footer, composer terkunci, form dengan field API key `aria-invalid`).
- **Review (cavecrew-reviewer, 0🔴 5🟡 6🔵 1❓), diterapkan:** shortcut dari Settings menuju chat (seperti `Kembali`), bukan pet; Settings tidak dihitung sebagai "error sudah dilihat" dan tidak menghapus "Sudah!" (hanya chat); banner error form hilang begitu field yang ditandai diperbaiki; `setForm` fungsional; chip "Balasan baru" hanya untuk teks baru (bukan retry/perubahan status); `useCopy` tidak set state setelah unmount. Diukur lewat CDP pada skala "Besar": hint pet 183/170 px terpotong dan rail konfirmasi 178/180 px, jadi teks pet dikunci 11/12 px (`#pet` me-redefinisi `--text-xs/--text-sm`; jendela pet tetap 180 px) → 155 px dan 158 px. Diterima: `clearTest` mengaktifkan tombol Tes lagi walau request lama masih berjalan (hasilnya dibuang oleh nomor urut); sisi `ToggleChat` yang memanggil runtime Wails hanya diuji manual. Vitest kini 109/109.
- **Risiko:** (a) `toggleFor` bergantung pada frontend memanggil `SetWindowMode` sebelum mode berganti (sudah dijamin `switchTo`); bila panggilan IPC gagal, mode backend bisa tertinggal satu langkah. (b) Hint pet tidak lagi menyebut "Geser"; geser tetap dijelaskan di tooltip tombol wortel.

## ADR-051 — Perbaikan dari critique delta (2026-10-11)

- **Status:** Done, perlu smoke manual Windows (fokus shortcut tidak bisa diuji di harness).
- **Konteks:** critique delta (hanya perubahan sejak `c272317`) memberi 32/40; kelima masalah ADR-050 terselesaikan. Temuan baru: (P2→P1 atas keputusan user, alurnya "tekan combo lalu ketik") shortcut dari pet yang terlihat membuka chat tanpa fokus keyboard; (P3) hasil tes bertahan setelah keluar Settings; (P3) fokus field gagal sebelum re-render dan banner jauh dari field; (P3) rail konfirmasi Keluar di pet menempel setelah fokus pergi. User juga memilih "Salin" menyalin teks biasa, dan meminta semua celah heuristik ditutup.
- **Keputusan fokus:** `ToggleChat` selalu `WindowShow` + `window.Activate()` (`SetForegroundWindow`; penerima `WM_HOTKEY` boleh mengambil foreground) pada jalur buka chat; frontend menaikkan `focusSignal` sehingga composer menerima caret walau chat sudah ter-mount.
- **Keputusan form:** pesan validasi tampil di bawah field-nya (`.field__error`, `role="alert"`, `aria-describedby`), fokus + `scrollIntoView({block:"center"})` dijalankan di effect setelah render; duplikat ID model memakai mekanisme yang sama; Settings membersihkan hasil tes saat di-unmount; baris model menampilkan ID sekali (alias hanya di input).
- **Keputusan pet:** rail "Keluar?" batal saat fokus keluar atau jendela blur; hint jatuh ke "Klik: chat · Geser: pindah" bila combo > 14 karakter; kursor wortel `grab`/`grabbing` mengajarkan geser.
- **Keputusan salin:** `markdownToText` (unit test) membuang tanda `** _ \` # >`, menyimpan isi code block tanpa fence, link menjadi "teks (url)"; tombol berlabel "Salin balasan" (beda dari "Salin" code block).
- **Keputusan konsistensi & bantuan:** oranye hanya untuk fokus dan aksi utama; pilihan (kartu aktif, chip preset, baris model aktif) memakai fill + badge/radio. Picker percakapan punya pencarian mulai 7 entri. Empty chat mengajarkan shortcut; section Provider punya satu kalimat pengantar; "Alamat API (Base URL)"; ikon model berupa kubus (globe terbaca sebagai bahasa/web).
- **Bukti:** test-first per langkah; vitest 72/72 (hook + `markdownToText` di-commit; test komponen untracked). `go vet`, `go test ./internal/... .`, `GOOS=linux go vet ./internal/window` bersih. Catatan: empat file test komponen untracked dari ADR-049/050 terhapus di luar sesi ini (bersamaan dengan commit `.gitignore`), tidak dibuat ulang.
- **Review (cavecrew-reviewer, 0🔴 8🟡 3🔵 1❓), diterapkan dengan test lebih dulu:** `markdownToText` menahan URL, inline code, dan escape backslash sebagai placeholder sebelum pass emphasis, dan emphasis hanya cocok di batas kata (`2*3*4`, `(1)*(2)`, `__init__.py`, `https://x/_a_/` utuh; `\*a\*` → `*a*`); kotak cari dan catatan kosong dipindah keluar `role="menu"`, panah dari kotak cari diteruskan ke daftar, Enter tanpa query tidak berpindah percakapan; menekan label/celah rail "Keluar?" tidak lagi memindah fokus (jadi tidak menutup rail); pilihan aktif (kartu, chip, baris model) memakai garis `ink-3` (≥ 3:1, bukan oranye); `Activate` menyertakan `GetLastError` dan jendela yang diminimalkan di-restore lebih dulu (`WindowUnminimise`); batas combo hint jadi konstanta, combo panjang tetap ada di tooltip. Tidak diterapkan: kursor `grab` pada wortel dipertahankan dengan sengaja (hint mengajarkan klik, kursor mengajarkan geser). Vitest 72/72.
- **Risiko:** `SetForegroundWindow` bisa ditolak Windows bila aturan foreground lock berlaku (mis. user sedang mengetik di app lain dengan input terakhir di sana); kegagalan hanya dicatat (`LogWarning`), jendela tetap tampil.

## ADR-052 — Perbaikan dari critique delta kedua (2026-10-11)

- **Status:** Done, perlu smoke manual Windows.
- **Konteks:** critique delta setelah ADR-051 memberi 36/40. Sisa: (P2, regresi) menu percakapan menutup saat padding atau catatan kosongnya ditekan, karena kotak luar yang men-scroll tidak bisa difokus; (P2) shortcut menciutkan chat yang terlihat tetapi tidak fokus (jendela AlwaysOnTop, jadi ini cara umum kembali); (P3) empat gaya "terpilih" di Settings; (P3) alias bisa diisi di dua tempat; H5: `Tes koneksi` bisa jalan tanpa model.
- **Keputusan:** `toggleFor(hidden, mode, focused)` dengan `window.IsForeground()` (`GetForegroundWindow`): panel di belakang app lain maju dengan caret di composer, hanya panel yang sedang dipakai yang menciut ke pet (unit test 6 kasus). Menu percakapan: daftar yang men-scroll, permukaan menu `tabIndex=-1` sehingga tekanan di padding tetap di dalam picker; ArrowUp dari item pertama kembali ke kotak cari. `Tes koneksi` tanpa model menampilkan alasan di bawah field Model. Alias hanya di baris modelnya (baris tambah = ID + Tambah). Chip preset yang dipilih memakai fill netral seperti pilihan lain.
- **Bukti:** vitest 76/76; `go vet`, `go test . ./internal/window`, `GOOS=linux go vet ./internal/window` bersih.
- **Risiko:** `GetForegroundWindow` membandingkan HWND utama; bila Wails membuat jendela anak yang memegang fokus, perbandingan bisa keliru dan shortcut memilih "maju" alih-alih "menciut" (aman, bukan kehilangan data).

## ADR-053 — Shortcut menjaga tampilan panel (2026-10-11)

- **Status:** Done, perlu smoke manual Windows.
- **Konteks:** critique delta setelah ADR-052 = 39/40. Celah H5: shortcut pada Settings yang tersembunyi/tak fokus mengirim `app:toggle` → chat, sehingga form provider yang belum disimpan hilang. Tiga P3: chip hover mirip chip terpilih; hint alias pada run pertama menyebut baris yang belum ada; menekan catatan "Tidak ada…" memindah fokus dari kotak cari.
- **Keputusan:** `toggleFront` untuk panel yang tersembunyi atau tak fokus: tampil + foreground + `app:front` tanpa ganti tampilan; frontend memberi caret ke composer hanya bila mode chat. Panel yang sedang dipakai tetap menciut ke pet; pet/jendela tersembunyi di mode pet tetap membuka chat. Hover chip hanya mempertegas outline; hint alias muncul setelah ada baris model; tekanan pada permukaan menu/catatan kosong tidak mengambil fokus dari kotak cari.
- **Bukti:** `TestToggleFor` (6 kasus) hijau; vitest 79/79.

## ADR-018 — M5 tray Win32 murni + exit bersih (2026-10-10)

- **Status:** Implemented, perlu smoke manual Windows.
- **Keputusan:** tray via Win32 langsung (`Shell_NotifyIcon` + popup menu) di `internal/tray`, tanpa CGO. Wails v2.16 tidak punya runtime tray; `fyne.io/getlantern systray` butuh CGO + gcc yang tidak tersedia (CGO_ENABLED=0).
- **Menu:** Buka (show chat) / Settings / Sembunyikan / Keluar. Klik kiri = Buka. Explorer restart ditangani (TaskbarCreated → re-add icon).
- **Exit:** `OnShutdown` → `Service.CancelAll` + `tray.Stop`. Frontend ikut event `app:show`/`app:settings`, tombol Hide di pet + chat.
- **Verifikasi:** `go test ./...` ok (termasuk `TestCancelAllStopsStreams` + tray mapping 3 tes), `gofmt` bersih, `tsc+vite` ok, `wails build` ok → `build/bin/companion.exe`.
- **Sisa M5:** smoke manual (icon muncul, menu 4 aksi, hide/show, exit saat streaming → cancel + proses bersih) + shortcut global, startup opsional, appearance, installer.



- Versi Windows minimum yang resmi didukung: Windows 10 22H2 dan/atau Windows 11? Tentukan sesuai lingkungan build dan kebutuhan pengguna.
- Apakah companion harus click-through pada pixel transparan, atau cukup window bounds ketat?
- Apakah posisi karakter perlu disimpan per monitor?
- Credential store library/API apa yang dipakai dan bagaimana license/dependency-nya?
- Apakah SQLite library yang dipilih mendukung target build dan packaging yang diperlukan?
- Apakah portable mode diperlukan pada MVP atau cukup executable dengan data di `%LOCALAPPDATA%`?
- Apakah conversation auto-title menggunakan judul lokal sederhana atau request tambahan ke provider? Default yang disarankan: judul dari potongan pesan pertama secara lokal untuk menghindari request biaya ekstra.
- Parameter API mana yang configurable pada UI dan mana yang ditetapkan pada default?

## ADR-054 — Lampiran gambar dan berkas (2026-10-11)

- **Status:** Done, perlu smoke manual Windows (dialog, drop, paste, model vision Kenari).
- **Konteks:** Provider yang dipakai (Kenari) meneruskan model multimodal; pengguna ingin bertanya tentang screenshot dan dokumen tanpa pindah ke browser. Spec: `docs/superpowers/specs/2026-10-11-attachments-design.md`.
- **Pilihan:** (A) staging bytes di Go + BLOB SQLite + route asset-server; (B) staging base64 di state React; (C) seperti A tetapi berkas di disk.
- **Keputusan:** (A). Dialog Wails hanya memberi path, jadi Go harus membaca; satu tabel `attachments` dengan cascade membuat kirim + simpan atomik tanpa sweep file; `<img src="/attachments/{id}">` menghindari base64 lewat IPC. Format request memakai `content` array standar OpenAI (`image_url` data URL, `file` untuk PDF, teks disisipkan sebagai part `text`) sehingga tidak ada parameter khusus provider. Hanya 3 turn user terakhir membawa lampiran (`AttachmentTurns`); lebih lama jadi placeholder teks karena gambar ditagih ulang tiap request. Tanpa gating kemampuan model: 400 yang menyebut image/file/vision/modal/document → `unsupported_input`, 413 → `payload_too_large`. Batas: 4 lampiran, gambar 8 MB, PDF 15 MB, teks 256 KB, total 20 MB; tipe dari magic bytes. Staging adalah isi pesan yang sedang disusun, batas ditegakkan saat menambah.
- **Perbaikan ikutan:** `RecentContext` tidak lagi mengirim baris assistant kosong (`streaming`) yang dulu ikut sebagai `{"role":"assistant","content":""}`.
- **Keputusan saat implementasi (review per task):** (1) pragma SQLite (`foreign_keys`, `busy_timeout`, `journal_mode`) kini dipasang lewat DSN `_pragma=` sehingga berlaku di setiap koneksi pool, bukan hanya koneksi pertama — tanpa ini `DELETE conversations` pada koneksi kedua tidak meng-cascade dan BLOB lampiran yatim (tes dua koneksi `TestForeignKeysOnEveryConnection`); (2) saat provider gagal di-resolve, turn user beserta lampirannya tetap disimpan dan baris assistant ditandai `failed`, sehingga "Coba lagi" membawa lampiran lagi; (3) retry teks kosong diizinkan bila DB ada (turn hanya-lampiran), ditolak tanpa DB; (4) `main.tsx` membatalkan default hanya untuk drag berkas (`types` memuat `Files`), drag teks ke composer tetap jalan; overlay drop di-reset pada `drop`/`dragend` window agar tidak macet; (5) `CleanName` memotong stem dan mempertahankan ekstensi pada nama > 120 rune.
- **Validasi:** test-first di `internal/ai`, `internal/attach`, `internal/storage`, `internal/app` (marshal parts, klasifikasi, staging, migrasi v5 + cascade + atomik, jendela 3 turn, retry membawa lampiran, route 404/405, error map) dan vitest (`useChat`, `useAttachments`, `ChatPanel`, `MessageItem`, `formatSize`).
- **Risiko:** screenshot PNG besar tanpa downscale → token mahal (downscale jadi kandidat lanjutan). `OnFileDrop` butuh `--wails-drop-target` di elemen dan belum diuji di build nyata sebelum smoke. Provider OpenAI-compatible lain mungkin tidak mengenal part `file`; dipetakan ke `unsupported_input`. Retry setelah penolakan sinkron `SendMessage` (bukan kegagalan provider) berjalan tanpa lampiran. Ukuran DB bertambah sebesar lampiran. Di `wails dev`, fallback SPA Vite berpotensi menjawab `/attachments/{id}` dengan `index.html` sehingga thumbnail tidak tampil di mode dev; build produksi memakai asset server Wails yang jatuh ke handler kustom — diverifikasi saat smoke.

## Template entri keputusan baru

```markdown
## ADR-XXX — Judul keputusan

- **Status:** Proposed / Accepted / Superseded
- **Tanggal:** YYYY-MM-DD
- **Konteks:** Masalah dan constraint.
- **Pilihan:** Opsi relevan yang dievaluasi.
- **Keputusan:** Opsi yang dipilih.
- **Konsekuensi:** Trade-off, risiko, dan pekerjaan lanjutan.
- **Validasi:** Tes/bukti yang dibutuhkan.
```
