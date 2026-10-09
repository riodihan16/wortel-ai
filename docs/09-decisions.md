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
