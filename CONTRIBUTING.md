# Kontribusi ke Wortel AI

Terima kasih sudah tertarik berkontribusi. Dokumen ini menjelaskan cara kerja di repo ini.

## Sebelum mulai

1. Baca `AGENTS.md` — berisi batas scope dan aturan kualitas yang berlaku untuk kontributor manusia maupun AI agent.
2. Baca `docs/02-mvp-requirements.md` (acceptance criteria) dan `docs/09-decisions.md` (keputusan yang sudah diambil).
3. Untuk perubahan besar, buka issue dulu dan sepakati pendekatan sebelum menulis kode.

## Batas scope (tidak boleh dilanggar tanpa persetujuan)

- Tidak ada eksekusi shell/PowerShell oleh agent.
- Tidak ada pembacaan screenshot, clipboard, atau file pengguna secara otomatis.
- Tidak ada telemetry, analytics, account system, atau cloud sync.
- Tidak ada model AI lokal.
- API key tidak boleh masuk source code, database, localStorage, atau log.
- Tidak ada tombol UI tanpa fungsi.

## Alur kerja

1. Fork repo dan buat branch dari `master`: `feat/...`, `fix/...`, atau `docs/...`.
2. Buat satu perubahan vertikal kecil yang bisa dijalankan dan diuji.
3. Jaga gaya kode yang ada; jangan menambah abstraksi atau dependency baru tanpa alasan.
4. Perbarui dokumen yang terpengaruh (termasuk `docs/09-decisions.md` bila keputusan berubah).
5. Buka pull request dengan deskripsi: apa yang berubah, mengapa, dan cara memverifikasi.

## Perintah verifikasi

Semua gate ini harus hijau sebelum PR di-review (jalankan dari root repo):

```bash
gofmt -l .                              # harus kosong
go vet ./...
go test ./...
CGO_ENABLED=1 go test -race ./...       # butuh gcc/MinGW-w64 di PATH
cd frontend && npx tsc --noEmit
cd frontend && npm test
cd frontend && npm run build
```

Build aplikasi:

```bash
wails build -ldflags "-X main.Version=0.1.0"
```

## Gaya kode

- **Go:** `context.Context` untuk operasi yang dapat dibatalkan; tutup `http.Response.Body`; jangan tinggalkan goroutine tanpa mekanisme stop; jangan log request/header yang mungkin membawa secret; pisahkan business logic dari adapter Wails.
- **TypeScript/React:** tipe eksplisit untuk payload backend dan event; hindari `any` tanpa justifikasi; cocokkan setiap delta dengan `requestId` dan `conversationId`; bersihkan listener pada unmount; jangan render raw HTML dari AI.
- Markdown dari AI adalah **input tidak tepercaya**.

## Commit

Gunakan Conventional Commits (`feat:`, `fix:`, `docs:`, `test:`, `chore:`). Pesan singkat, jelaskan "apa" dan "mengapa", bukan "bagaimana".

## Melaporkan bug keamanan

Jangan buka issue publik untuk kerentanan. Ikuti `SECURITY.md`.
