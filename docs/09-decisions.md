# 09 — Decision Log dan Risiko yang Belum Terjawab

Dokumen ini mencatat keputusan arsitektur agar perubahan tidak terjadi secara diam-diam. Tambah entri baru saat keputusan penting diubah.

## Status keputusan

| ID | Keputusan | Status | Alasan |
|---|---|---|---|
| ADR-001 | Gunakan Wails v2.16.0 + Go 1.26.4 + Node v24.18.0 | Dikunci 2026-10-09 | v2.16.0 rilis stabil v2 terakhir; v3 masih beta. Go/Node versi terpasang, penuhi syarat (Go 1.21+, Node 15+). |
| ADR-002 | Backend aplikasi menggunakan Go | Dipilih | Cocok untuk HTTP, streaming, concurrency, dan desktop utilities. |
| ADR-003 | Frontend React + TypeScript + Tailwind | Dipilih | Ekosistem dan pola pengembangan familier serta cocok untuk UI chat. |
| ADR-004 | Tidak menjalankan model lokal di MVP | Dipilih | Fokus pada provider API dan idle footprint. |
| ADR-005 | Satu jendela/WebView berpindah mode pet/chat | Diterima 2026-10-09 | Uji manual: drag mulus, toggle M0 pet/chat OK, Center OK, Probe OK, Quit bersih. |
| ADR-006 | SQLite untuk conversation history | Dipilih sementara | Embedded, tidak membutuhkan server database terpisah. |
| ADR-007 | API key di OS credential storage | Dipilih sebagai persyaratan | Menghindari plaintext di database/config. Implementasi konkret harus diprototipekan. |
| ADR-008 | MVP mendukung chat text + SSE saja | Dipilih | Membatasi scope dan permukaan bug. |
| ADR-011 | Baseline M1 companion | Validated 2026-10-09 | Fresh `wails init` react-ts di repo; main ~30MB CPU idle 0–0.21% (12 core); window 180x180; toggle pet 180x180 / chat 400x600 via WindowSetSize. Uji manual user: drag + toggle klik lancar. |

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
