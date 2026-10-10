# Wortel AI

> **Status:** MVP selesai (0.1.0), siap dirilis sebagai executable portable Windows.
> **Platform target:** Windows 10/11 x64
> **Bahasa dokumentasi:** Indonesia
> **Lisensi:** [MIT](LICENSE) © 2026 Ridho Dwi Ramadhan
>
> Desktop companion berkarakter 2D yang ringan dan dapat bercakap-cakap dengan AI melalui endpoint OpenAI-compatible.

## Fitur

- Karakter kecil mengambang di desktop; klik untuk membuka panel chat.
- Chat streaming ke endpoint OpenAI-compatible, dengan tombol berhenti yang benar-benar membatalkan request.
- Beberapa provider dan beberapa model per provider; ganti model aktif dari header chat.
- Riwayat percakapan lokal (SQLite) + multi-percakapan (buat, ganti nama, hapus).
- Markdown aman untuk respons AI (tanpa raw HTML; link hanya http/https).
- API key disimpan di Windows Credential Manager, tidak pernah ke database/log/frontend.
- Hotkey global, system tray, dan perilaku idle ringan (tanpa polling, tanpa model lokal).

## Prinsip

- **Ringan saat idle:** tidak ada polling, tidak ada loop kerja terus-menerus, tidak ada model lokal.
- **Provider bebas:** endpoint, model, dan API key dapat dikonfigurasi selama provider mendukung kontrak API yang kompatibel.
- **Local-first:** pengaturan dan riwayat disimpan lokal; percakapan hanya dikirim saat pengguna menekan tombol kirim.
- **Aman secara default:** rahasia tidak disimpan sebagai teks biasa, tidak ditulis ke log, dan respons AI tidak diperlakukan sebagai perintah sistem.

## Dokumen

Dokumen perencanaan dan teknis ada di `docs/`:

1. [Product brief](docs/01-product-brief.md) — visi, pengguna, prinsip produk, dan batas scope.
2. [Kebutuhan MVP](docs/02-mvp-requirements.md) — kebutuhan fungsional/nonfungsional dan acceptance criteria.
3. [Arsitektur teknis](docs/03-architecture.md) — komponen, alur data, keputusan arsitektur, risiko teknis.
4. [Struktur project](docs/04-project-structure.md) — tata letak direktori dan tanggung jawab paket.
5. [Integrasi AI provider](docs/05-ai-provider-integration.md) — konfigurasi provider, streaming, error handling, dan kontrak API.
6. [UI/UX companion](docs/06-ui-ux.md) — mode karakter, chat, pengaturan, dan perilaku interaksi.
7. [Penyimpanan, keamanan, dan privasi](docs/07-storage-security-privacy.md) — data lokal, API key, batas keamanan, dan retensi.
8. [Roadmap dan pengujian](docs/08-roadmap-testing.md) — tahapan implementasi, strategi tes, serta kriteria rilis.
9. [Keputusan teknis](docs/09-decisions.md) — keputusan yang sudah diambil dan hal yang masih perlu dibuktikan.
10. [Panduan implementasi untuk coding agent](AGENTS.md) — cara bekerja di repo, urutan implementasi, dan aturan kualitas.

## Keputusan awal

| Area | Pilihan awal | Catatan |
|---|---|---|
| Desktop framework | Wails v2.16.0 | Versi dikunci 2026-10-09; jangan pakai v3 beta. Validasi API Windows yang dibutuhkan. |
| Backend | Go 1.26.4 | Syarat Wails v2: Go 1.21+. Versi terpasang 1.26.4, dikunci untuk M0. |
| Frontend | React 19.1.0 + TS 5.6.3 + Vite 7.3.7 | Dari template `wails init` react-ts M0, dikunci 2026-10-09. |
| Styling | CSS biasa dengan design token (variabel tema) | Tailwind dihapus 2026-10-10: hanya preflight-nya terpakai, tidak ada utility class di JSX. Token di `App.css`. |
| Chat API | OpenAI-compatible Chat Completions | Tidak semua provider mendukung seluruh parameter dengan cara yang sama. Setiap provider menyimpan daftar model; satu ditandai aktif. |
| Render markdown | react-markdown + remark-gfm + rehype-sanitize | AI output = input tidak tepercaya. Tanpa raw HTML; sanitize allowlist; link hanya http/https. |
| Streaming | SSE yang diparse di backend Go | Frontend menerima token/chunk melalui event Wails. |
| Penyimpanan | SQLite | Percakapan dan preferensi nonrahasia. |
| Secret | Windows Credential Manager atau mekanisme OS yang sesuai | Jangan simpan API key dalam SQLite atau JSON biasa. |
| Model lokal | Tidak pada MVP | Mengurangi kompleksitas dan kebutuhan RAM. |
| Jendela | Satu jendela yang berpindah mode karakter/chat | Transparansi, resize, hit-testing, dan drag perlu prototipe Windows. |

## Cara memakai dokumen ini

- Baca `01-product-brief.md` dan `02-mvp-requirements.md` sebelum mengerjakan UI.
- Gunakan `03-architecture.md` dan `04-project-structure.md` sebagai batas modul, bukan alasan untuk membuat abstraksi berlebihan.
- Implementasikan milestone pada `08-roadmap-testing.md` secara berurutan.
- Perbarui `09-decisions.md` saat keputusan penting berubah.
- Jika ada konflik antar-dokumen, dahulukan requirements MVP dan catat keputusan baru di decision log.

## Definisi MVP selesai

MVP dianggap layak diuji ketika pengguna dapat membuka aplikasi Windows, melihat dan memindahkan karakter, membuka chat, mengirim pesan ke provider OpenAI-compatible, menerima jawaban streaming, membatalkan respons, menyimpan/membuka percakapan lokal, mengatur provider dengan aman, serta menutup aplikasi tanpa proses menggantung. Kriteria rinci berada di `02-mvp-requirements.md`.

## Menjalankan dan membangun

Prasyarat untuk **development**: Go 1.26.4, Node.js + npm, Wails CLI v2.16.0 (`go install github.com/wailsapp/wails/v2/cmd/wails@v2.16.0`). WebView2 Runtime diperlukan untuk menjalankan aplikasi.

Development (hot reload frontend):

```bash
wails dev
```

Verifikasi:

```bash
go test ./...        # unit test backend
CGO_ENABLED=1 go test -race ./...   # race detector (butuh gcc/MinGW-w64 di PATH)
gofmt -l app.go main.go internal/
cd frontend && npm test        # unit test frontend (vitest + jsdom)
cd frontend && npm run build   # typecheck + bundle frontend
```

Build release (menghasilkan satu executable):

```bash
wails build -ldflags "-X main.Version=0.1.0"
```

Hasil: `build/bin/companion.exe`. Aset frontend dibundel ke dalam binary, sehingga **pengguna rilis tidak perlu memasang Go atau Node.js**. Jika WebView2 tidak tersedia, aplikasi menampilkan pesan yang dapat dimengerti.

`-ldflags "-X main.Version=..."` menstamp versi rilis; binding `Version()` menampilkannya di header chat (build tanpa flag menampilkan `dev`). `wails.json` `info.productVersion` mengisi metadata exe Windows.

Lokasi data pengguna: `%LOCALAPPDATA%\WortelAI\wortel.db` (SQLite). API key **tidak** disimpan di file ini; disimpan di Windows Credential Manager.

### Distribusi dan uninstall

Aplikasi MVP dirilis sebagai **executable portable** — tidak ada installer, tidak ada entri registry, tidak ada service. Salin `build/bin/companion.exe` ke folder mana pun lalu jalankan.

Cara uninstall:
1. Tutup aplikasi (menu tray → Keluar, atau tombol Quit).
2. Hapus file `companion.exe`.
3. Hapus data pengguna: hapus folder `%LOCALAPPDATA%\WortelAI\` (berisi `wortel.db` dan file WAL-nya).
4. Hapus credential API key: buka **Credential Manager** (Panel Kontrol → Kredensial Windows), cari entri dengan target berawalan `WortelAI`, hapus satu per satu. API key disimpan di sini, bukan di folder data.

Tidak ada langkah lain yang diperlukan; tidak ada komponen yang tertinggal di sistem.

### Penandatanganan rilis (code signing)

Executable saat ini **belum ditandatangani**, sehingga Windows SmartScreen menampilkan "Windows protected your PC". Untuk menjalankan: **More info → Run anyway**. Tanda tangan tidak menghilangkan peringatan pada rilis pertama; ia menggantinya dengan nama penerbit yang terverifikasi dan membangun reputasi seiring unduhan.

Opsi untuk proyek ini (open source):

- **SignPath Foundation** (gratis untuk OSS yang lolos kualifikasi): penandatanganan OV-level lewat pipeline terkelola. Jalur gratis paling realistis untuk proyek ini; lihat `.github/workflows/release.yml`.
- **Azure Trusted Signing** (~$10/bulan): termurah yang kredibel dan ramah CI, tetapi **individu dibatasi Amerika Serikat & Kanada**, sehingga kemungkinan tidak tersedia untuk pengelola di Indonesia.
- **OV certificate** dari CA (Certum/Sectigo/DigiCert, ~$99–300/tahun): berlaku global; sejak Juni 2023 private key wajib di hardware token/HSM.
- **EV certificate** (~$400+/tahun): **jangan** dibeli demi SmartScreen — sejak 2024 EV tidak lagi bypass instan.

Menandatangani secara lokal (setelah punya sertifikat/token):

```bash
signtool sign /fd SHA256 /tr http://timestamp.digicert.com /td SHA256 /n "Ridho Dwi Ramadhan" build/bin/companion.exe
```


### Platform yang didukung

- Windows 10/11 x64. Ini satu-satunya target MVP.
- Belum diuji pada Windows ARM64 atau Windows versi lebih lama.

## Referensi teknis resmi

- Wails: <https://wails.io/>
- Dokumentasi Wails v2: <https://wails.io/docs/v2.0/introduction>
- Go: <https://go.dev/doc/>
- React: <https://react.dev/>
- SQLite: <https://www.sqlite.org/docs.html>
- OpenAI API reference: <https://platform.openai.com/docs/api-reference/chat>

Endpoint dan format spesifik harus dikonfirmasi terhadap dokumentasi provider yang benar-benar digunakan. Istilah “OpenAI-compatible” bukan jaminan bahwa semua parameter, endpoint, tool calls, atau streaming identik.

## Berkontribusi

Kontribusi diterima. Baca [CONTRIBUTING.md](CONTRIBUTING.md) untuk alur kerja dan gate verifikasi, lalu [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md). Untuk kerentanan keamanan, ikuti [SECURITY.md](SECURITY.md) — jangan buka issue publik.

## Lisensi

Dirilis di bawah [MIT License](LICENSE) © 2026 Ridho Dwi Ramadhan.
