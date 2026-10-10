# 08 — Roadmap Implementasi, Testing, dan Release Checklist

## 1. Prinsip delivery

Bangun satu jalur vertikal yang berjalan dari jendela sampai provider sebelum memperluas fitur. Jangan membuat semua struktur folder dan service abstrak di awal jika belum dipakai.

## 2. Milestone

### M0 — Validasi risiko teknis

**Tujuan:** menguji Wails di Windows sebelum banyak kode dibuat.

Tugas:
- Buat project Wails minimal dengan versi yang dikunci.
- Build dan jalankan di Windows.
- Uji window frameless dan background transparan.
- Uji resize dari mode kecil ke panel chat.
- Uji drag, posisi, always-on-top, dan fokus.
- Ukur baseline RAM/CPU seluruh process tree setelah startup dan idle.
- Pastikan perilaku WebView2 pada mesin target.

Definition of done:
- Ada build runnable.
- Dicatat perilaku yang berhasil/gagal.
- Keputusan satu jendela tetap layak atau perlu alternatif ditulis di `09-decisions.md`.

### M1 — Companion statis

Tugas:
- Asset karakter lokal.
- Mode pet dan mode panel dasar.
- Drag dan posisi aman.
- Toggle pet/chat.
- Tidak ada API/network/storage yang belum diperlukan.

Definition of done:
- Karakter tampil konsisten.
- Tidak terpotong di tepi layar.
- Animasi idle tidak menyebabkan CPU tinggi.
- Bisa ditutup secara normal.

### M2 — Chat tanpa persistence kompleks

Tugas:
- Satu provider/config development yang tidak menyimpan secret ke repo.
- Request non-streaming awal untuk membuktikan koneksi.
- Tambahkan SSE streaming dan event contract.
- Tambahkan Stop/cancel.
- Error mapping dasar.

Definition of done:
- Request berhasil ke endpoint uji.
- Streaming dan cancellation lulus.
- Tidak ada API key di UI event atau log.
- Tidak ada request duplikat ketika tombol diklik cepat.

### M3 — Provider settings aman

Tugas:
- CRUD provider profile.
- Credential store OS.
- Validasi base URL/model.
- Test connection.
- Pilih provider aktif.

Definition of done:
- Credential tidak disimpan di SQLite atau settings JSON.
- Mengganti provider tidak menghilangkan riwayat.
- Kesalahan konfigurasi dapat dipahami pengguna.

### M4 — SQLite conversation history

Tugas:
- Schema dan migration.
- Repository percakapan dan pesan.
- List/open/rename/delete conversation.
- Penyimpanan state complete/cancelled/failed.

Definition of done:
- Percakapan bertahan sesudah restart.
- Request yang dibatalkan atau gagal tidak dianggap sukses.
- Migration dan kegagalan DB diuji.

### M5 — Kualitas desktop

Tugas:
- Shortcut global.
- System tray.
- Startup opsional.
- Settings appearance.
- Installer atau portable bundle.

Definition of done:
- Shortcut tidak bentrok secara diam-diam.
- Exit menutup semua resource.
- Startup behavior jelas dan opt-in.
- Installer/portable build diuji pada Windows target yang bersih.

**Status (2026-10-10):** tray + startup opt-in + shortcut global + appearance selesai (ADR-018…021). Rilis MVP memakai portable executable tunggal, bukan installer (ADR-022). Tersisa: smoke manual user untuk tray/hotkey/appearance dan uji build pada Windows bersih.

### M6 — RC dan dokumentasi rilis

Tugas:
- Regression test.
- Pengukuran RAM/CPU.
- Uji konfigurasi dan error provider.
- Review secret handling dan Markdown security.
- Tulis cara install/uninstall, lokasi database, dan cara menghapus data.

Definition of done:
- Seluruh P0 dari requirement lulus.
- Known issues dicatat.
- Versi dan hash artifact rilis dicatat bila sesuai.

## 3. Strategi testing

### Unit test backend

- URL normalization.
- Provider config validation.
- SSE parser (frame terfragmentasi, banyak baris data, komentar, `[DONE]`, EOF, malformed JSON).
- Error mapping HTTP.
- Conversation/message repository.
- Credential reference lifecycle (mock credential store untuk test).
- Cancellation dan request ID mapping.

Gate backend: `gofmt -l .`, `go vet ./...`, `go test ./...`, dan `CGO_ENABLED=1 go test -race ./...`
(MinGW-w64 terpasang; lihat Known issues untuk asal gcc).

### Integration test

Gunakan HTTP test server lokal atau mock provider untuk:
- Non-stream response.
- SSE dengan beberapa delta.
- Stream lambat yang dibatalkan.
- 401, 403, 404, 429, 500.
- Response JSON malformed.
- Server disconnect di tengah stream.
- Timeout.

Jangan menjadikan test suite wajib bergantung ke API berbayar/internet. Smoke test provider nyata dijalankan secara manual dengan credential lokal yang tidak dicatat.

### Frontend test

Runner: `vitest` + `jsdom` (`frontend/vitest.config.ts`). Jalankan `npm test` di `frontend/`.
Binding Go dan event bus Wails diganti test double di `src/test/wails.ts`; tidak ada network
atau backend Go yang dibutuhkan.

- Composer mengirim/menolak input kosong. → `useChat.test.ts` (spasi saja tidak dikirim).
- Enter vs Shift+Enter. → belum (butuh test komponen composer; ditunda).
- State Send/Stop. → status `streaming`/`complete`/`cancelled`/`error` di `useChat.test.ts`.
- Delta dengan request ID yang salah diabaikan. → `useChat.test.ts` (guard terbukti: menghapus guard membuat tes gagal).
- Error notice dan retry state. → pesan error dari event `failed` muncul di state `useChat.test.ts`.
- Provider key tidak muncul dalam state visible/persisted. → belum di runner; dibuktikan lewat pemeriksaan DB nyata + `ChatEvent` tanpa secret.
- Markdown link scheme dibatasi. → `SafeMarkdown.test.tsx` (4 tes; `javascript:`/`data:` jadi teks, raw HTML inert).
- Delta untuk conversation lain diabaikan. → `useChat.test.ts`.

### Manual test Windows

- Windows 10/11 yang didukung.
- WebView2 tersedia/tidak tersedia jika memungkinkan diuji.
- Scaling 100%, 125%, 150%.
- Satu monitor dan multi-monitor bila tersedia.
- Resolusi rendah dan tinggi.
- Window berpindah ke berbagai tepi layar.
- Sleep/resume atau display change bila relevan.
- Aplikasi keluar ketika request aktif.
- Restart komputer untuk menguji startup hanya setelah fitur dibuat.

## 4. Pengujian performa

Ukur pada build release, bukan hanya dev mode. Catat OS, hardware, resolusi, scaling, dan versi WebView2 bila dapat diperoleh.

Skenario:
1. Jalankan aplikasi tanpa provider.
2. Diamkan 60 detik pada mode pet.
3. Buka/tutup panel chat beberapa kali.
4. Buka riwayat panjang.
5. Streaming jawaban panjang.
6. Tutup aplikasi ketika request berjalan.

Metrik:
- RAM process utama + proses WebView2 yang berkaitan.
- CPU average dan peak selama idle.
- Startup time.
- Responsivitas window saat streaming.
- Pertumbuhan RAM setelah membuka/tutup beberapa conversation.

Target awal yang disarankan (bersifat **target eksperimen**, bukan jaminan):
- CPU idle rata-rata mendekati 0–1% setelah startup dan aktivitas berhenti.
- Tidak ada pertumbuhan RAM terus-menerus selama idle.
- Startup pada SSD dalam beberapa detik di mesin referensi.
- Tetapkan budget RAM final berdasarkan pengukuran baseline M0; jangan menyembunyikan konsumsi proses WebView2.

## 5. Release checklist

Status diisi 2026-10-10 untuk rilis 0.1.0 (build release di mesin referensi Windows 11 26200, 12 core).
Tanda: `[x]` terverifikasi (bukti disebut), `[~]` terverifikasi sebagian / otomatis saja, `[ ]` belum.

### Fungsional
- [x] Build release berhasil dari clean checkout. → `wails build -ldflags "-X main.Version=0.1.0"` → `build/bin/companion.exe`, app start (WebView2 Environment created).
- [x] Companion bisa tampil, drag, hide/show. → tampil + hide/show via tombol & tray API terverifikasi; drag fisik dan hide/show via ikon tray diuji manual oleh user pada release build (lulus).
- [x] Panel chat bisa dibuka dan ditutup. → mode pet/chat/settings; draft bertahan lewat switch (probe browser live).
- [x] Pesan terkirim ke provider yang dipilih. → mock provider E2E lulus; endpoint OpenAI-compatible nyata diuji manual dengan credential lokal pada release build (streaming + cancel, lulus).
- [x] Streaming berfungsi dan dapat dibatalkan. → `TestStreamDeltasFromServer`, `TestSendAndCancel`, `TestCancelAllStopsStreams`; UI Stop↔Kirim toggle terverifikasi.
- [x] Error provider tidak membuat aplikasi crash. → error mapping 400/401/403/404/408/429/5xx + timeout + diskonek tengah + malformed JSON (`provider_test.go`, `sse_test.go`); DB failure = error bukan panic (`db_test.go`).
- [x] Percakapan bertahan setelah restart. → `TestChatHistoryRoundTrip`, `TestMigrationV1toV2`, `TestSwitchActiveKeepsHistory`; DB nyata di `%LOCALAPPDATA%\WortelAI\wortel.db`.
- [x] Pengaturan provider dapat diubah dan dites. → `TestProviderCRUD`, `TestTestConnectionMapping`, `TestValidate*`.
- [x] Multi model per provider + ganti model aktif. → `TestMigrationV2toV3`, `TestProviderModelsRoundTrip`, `TestSaveMultiModelAndSwitch`, `TestNormalizeModels*`; UI: editor daftar model di Settings + tombol pemilih model di header chat (ADR-028, ADR-029).
- [x] Kontrol di dalam area drag tidak butuh klik ganda. → `.header-actions` dan kontrol pet menyetel `--wails-draggable: none`; terukur pada probe (ADR-029).
- [x] Kontras teks/latar memenuhi WCAG AA di kedua tema. → terukur: gelap terendah 8.48, terang terendah 5.47 (ADR-029).
- [x] Spasi antar baris respons wajar. → hapus `white-space: pre-wrap` dari `.msg` (react-markdown newline tak lagi jadi baris kosong); `<p>` = 18px + 6px margin (ADR-029).
- [x] Teks pet hint terbaca di kedua tema. → `#pet-hint` warna teks putih tetap di pill gelap (ADR-030).
- [~] Latar pet transparan, W bisa diklik dan digeser. → probe piksel release build: sudut window = piksel desktop, tengah = `--bg`; klik vs geser di `usePetDrag.test.ts` (10 tes). Drag native di app nyata menunggu uji manual user (ADR-041).
- [x] Semua kontrol header muat di jendela chat 400px. → judul menyusut (`min-width: 0`), `.header-actions` `flex-shrink: 0`, `#btn-model` max 132px; tanpa overflow (ADR-030).
- [x] Percakapan baru tidak mengunci pemilihan history. → placeholder `<option>` untuk id aktif yang belum tersimpan (ADR-030).
- [x] Pemilih model di header tampil pada build nyata. → `ActiveProvider()` ganti ke `*config.ProviderProfile` (binding Wails buang return kedua non-error); hook memetakan profile/null apa adanya (ADR-031).
- [x] Pemilih model dan status ada di footer bawah chat, sejajar. → `#chat-footer` setelah composer, `space-between`; dropdown buka ke atas (`bottom: 100%`) dan `left: 0` agar tidak terpotong (ADR-033, ADR-034).

### Keamanan
- [x] Tidak ada API key di repository, log, SQLite, dan event frontend. → `provider_profiles` tanpa kolom secret; grep pola `sk-` pada DB nyata = tidak ada; `ChatEvent` tanpa secret; tidak ada `log.Printf` pada request/header; secret di-wipe (`secret[i]=0`).
- [x] Markdown raw HTML dimatikan/disanitasi. → `react-markdown` tanpa `rehype-raw` + `rehype-sanitize` allowlist. Render live: `<img src=x onerror=alert(1)>` tetap teks literal, 0 elemen img/script (ADR-028).
- [x] Link eksternal menggunakan allowlist scheme. → schema protocol `href: [http, https]` + override komponen; live: `[jahat](javascript:alert(1))` dirender sebagai teks tanpa `<a>`, hanya `https://` jadi link.
- [x] Redirect cross-host tidak membawa Authorization. → `TestSafeClientDropsAuthOnCrossHost` + `TestSafeClientKeepsAuthOnSameHost` (ADR-026).
- [x] Tidak ada command execution atau screenshot capture. → tidak ada kode shell/screenshot/clipboard di repo.
- [x] HTTP endpoint non-TLS mendapat warning yang sesuai. → `config.Validate` set `InsecureTransport` untuk `http://` non-loopback (`TestValidateHTTPFlagsInsecure`).

### Performa dan stability
- [x] Idle CPU/RAM diukur pada release build. → `scripts/perf.ps1`: idle CPU 0.00%, app ~33 MB, WebView2 ~308 MB, total ~341 MB (ADR-024).
- [x] Tidak ada goroutine/request yang tertinggal setelah cancel/exit. → `sync.WaitGroup` + `CancelAll` menunggu; `TestCancelAllWaitsForGoroutines` (state `active`/`requests` kosong).
- [x] UI tetap merespons saat streaming. → event-driven (`chat:event`); render live busy: Stop tampil, kontrol lain disabled, delta render.
- [x] Window tidak hilang di luar area layar setelah monitor berubah. → display-change fisik diuji manual oleh user pada monitor nyata (lulus). Koreksi 2026-10-10: klaim "logika clamp dari M0/M5" keliru, kode tidak punya clamp; clamp nyata ditambah di ADR-042.
- [~] Panel chat tidak terpotong tepi monitor; drag berhenti di tepi. → `internal/window` (`Placer`, `KeepOnScreen`), 13 unit test; probe app nyata: pet di pojok kanan-bawah → panel `panelInsideWork=True`. Drag live di tepi menunggu uji manual user (ADR-042).
- [x] Database failure memberikan pesan yang jelas. → `TestOpenRejectsEmptyPath`, `TestOpenFailsWhenDirUncreatable`, `TestUseAfterCloseDoesNotPanic`; `errNoDB` = "storage belum tersedia, restart aplikasi".

### Distribusi
- [x] Versi aplikasi ditampilkan. → `App.Version()` di header chat; `0.1.0` terverifikasi tertanam di binary (ADR-025).
- [x] Cara install/run dan uninstall didokumentasikan. → README "Distribusi dan uninstall".
- [x] Lokasi data lokal dijelaskan. → `%LOCALAPPDATA%\WortelAI\wortel.db` (+ `-wal`/`-shm`); README + docs 07.
- [x] WebView2 requirement dijelaskan. → README "Menjalankan dan membangun"; kegagalan start menampilkan `MessageBoxW` dengan tautan unduh WebView2 bila runtime absen (ADR-038).
- [x] Signing/SmartScreen behavior dipertimbangkan sebelum distribusi publik. → **Belum ditandatangani.** Executable tanpa Authenticode memicu peringatan SmartScreen "Windows protected your PC". Untuk MVP portable pribadi, pengguna menekan "More info → Run anyway". Proyek kini open source (MIT) sehingga jalur gratis tersedia: **SignPath Foundation** (OV-level, pipeline terkelola) via `.github/workflows/release.yml`, digerbangi environment `release` dan `vars.SIGNPATH_ENABLED`; lihat ADR-040 dan README "Penandatanganan rilis". Opsi berbayar (Azure Trusted Signing, OV/EV CA) dan langkah `signtool` lokal juga didokumentasikan di README.
- [x] CI dan rilis otomatis. → `.github/workflows/ci.yml` (gofmt, `go vet`, `go test`, `go test -race`, typecheck/test/build frontend, `wails build` + upload artifact) dan `.github/workflows/release.yml` (tag `v*` → build → SignPath opsional → GitHub release draft + `SHA256SUMS.txt`). Belum pernah dijalankan di GitHub (repo belum di-push); validasi dilakukan lokal (YAML parse + gate yang sama).

### Artifact rilis 0.1.0

- **File:** `build/bin/companion.exe`
- **Ukuran:** 15,687,680 byte (~15.0 MB)
- **SHA-256:** `8293dcc63700f0b3dc1fbba793abe986350ac90de5cae22c0f2eed2561b2cbfa`
- **Perintah build:** `wails build -ldflags "-X main.Version=0.1.0"` dari clean checkout (setelah `rm -rf frontend/dist build/bin`).
- **Verifikasi:** versi `0.1.0` tertanam di binary (mekanisme `-X main.Version` dibuktikan balik lewat probe `go run`); aplikasi start dan tercatat `[WebView2] Environment created successfully`. Migrasi DB user nyata v2→v3 dan v3→v4 terverifikasi (kolom `models`/`active_model`; migrasi v4 mengubah daftar model string menjadi entri `{id,alias}`).
- **Catatan:** hash diperbarui pada M6 hardening (ADR-045 alias + ADR-035 regression suite frontend) setelah build ulang dari clean checkout.

### Known issues (0.1.0)

- **Bug redirect diperbaiki sebelum rilis:** sebelumnya redirect cross-host membawa `Authorization`; kini dihapus (ADR-026).
- Executable belum ditandatangani → peringatan SmartScreen (lihat Distribusi).
- Beberapa item diuji manual oleh user pada release build dan lulus: drag karakter, hide/show tray (ikon), hotkey global `Ctrl+Shift+W`, perilaku display-change, dan streaming ke endpoint OpenAI-compatible sungguhan.
- `go test -race ./...` hijau. Mesin build punya MinGW-w64 (WinLibs POSIX/UCRT, `gcc` di PATH user via `winget install BrechtSanders.WinLibs.POSIX.UCRT`). Jalankan race dengan `CGO_ENABLED=1 go test -race ./...`. Dua data race di **kode test** ditemukan dan diperbaiki: emitter tidak sinkron di `TestSendAndCancel` dan `TestCancelAllStopsStreams` kini lewat `eventSink` ber-mutex. Produksi tidak terpengaruh (`s.mu` sudah menjaga map; buffer streaming lokal per goroutine). Build rilis tetap `CGO_ENABLED=0` (artefak tidak memakai CGO).
- UI frontend kini punya runner permanen: `vitest`+`jsdom`, 12 tes (`useChat` reducer/guard, `SafeMarkdown` allowlist link). Sisa yang belum ditutup otomatis: composer Enter vs Shift+Enter, busy state kontrol, dan render provider key; masih mengandalkan probe/manual.
- Kegagalan persistensi percakapan (disk penuh / handle DB hilang) **tidak terlihat user** pada 0.1.0: stream tetap jalan, pesan hilang saat restart. Best-effort by design (ADR-037); tidak ada tes otomatis untuk jalur emisi karena sulit direproduksi tanpa hook test-only.
- SmartScreen: executable belum ditandatangani (lihat Distribusi) — jalur signing OSS (SignPath Foundation) sudah disiapkan di `.github/workflows/release.yml` tapi belum aktif; rilis publik pertama tetap memicu peringatan sampai akun SignPath lolos kualifikasi (ADR-040).


