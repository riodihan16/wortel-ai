# Panduan untuk Coding Agent — Wortel AI

Dokumen ini memberi batas kerja untuk AI coding agent atau kontributor manusia. Ikuti requirements MVP dan decision log; jangan memperluas scope sendiri.

## Tujuan

Membangun aplikasi desktop companion Windows yang ringan dengan karakter 2D, chat AI, serta integrasi endpoint OpenAI-compatible. Output rilis utama adalah executable Windows; installer dapat menjadi milestone lanjutan.

## Baca sebelum coding

1. `README.md`
2. `docs/01-product-brief.md`
3. `docs/02-mvp-requirements.md`
4. `docs/03-architecture.md`
5. `docs/05-ai-provider-integration.md`
6. `docs/07-storage-security-privacy.md`
7. `docs/08-roadmap-testing.md`
8. `docs/09-decisions.md`

## Batas scope yang tidak boleh dilanggar tanpa persetujuan

- Jangan menambah fitur agent yang mengeksekusi shell/PowerShell.
- Jangan membaca screenshot, clipboard, atau file pengguna secara otomatis.
- Jangan menambah telemetry, analytics, account system, atau cloud sync.
- Jangan memasukkan model AI lokal.
- Jangan menyimpan API key dalam source code, database, localStorage, atau log.
- Jangan mengimplementasikan banyak abstraksi/plugin framework yang belum diperlukan.
- Jangan membuat tombol UI yang belum memiliki fungsi; jika fitur ditunda, jangan ditampilkan sebagai aktif.

## Prinsip implementasi

- Buat satu perubahan vertikal kecil yang bisa dijalankan dan diuji.
- Ikuti versi toolchain dan dependency yang sudah dikunci repo; jangan upgrade mayor tanpa alasan dan pengujian.
- Backend Go mengelola request AI, API key, database, cancellation, dan operasi sistem.
- Frontend React hanya memanggil binding yang tersedia dan menerima event terstruktur.
- Validasi input di backend meski frontend sudah memvalidasi.
- Berikan error yang aman dan dapat ditindaklanjuti; jangan menyertakan secret di pesan error.
- Semua request jaringan harus punya context/cancellation dan cleanup yang jelas.
- Jangan memblokir UI selama streaming atau query.
- Gunakan transaksi database untuk perubahan yang harus atomik.
- Markdown yang berasal dari AI adalah input tidak tepercaya.

## Urutan pengerjaan yang disarankan

1. M0: buktikan window Wails, frameless/transparency/drag/resize dan catat baseline resource.
2. M1: karakter statis dan mode pet/chat.
3. M2: chat ke mock provider, kemudian endpoint OpenAI-compatible dan streaming/cancel.
4. M3: provider settings dan credential storage yang aman.
5. M4: SQLite history dan migration.
6. M5: shortcut, tray, startup opsional, dan packaging.
7. M6: hardening, regression, pengukuran, dan dokumentasi rilis.

Jangan melompati M0; risiko window Windows dapat memengaruhi desain keseluruhan.

## Kualitas kode

### Go
- Gunakan `context.Context` untuk request dan operasi yang dapat dihentikan.
- Tutup `http.Response.Body` di semua jalur yang sesuai.
- Hindari goroutine tanpa mekanisme stop.
- Jangan menggunakan `log.Printf` pada object request/header yang mungkin membawa secret.
- Tangani error; jangan mengabaikan error database, stream, file, atau credential storage.
- Pisahkan business logic dari Wails adapter.
- Tambahkan unit test untuk URL normalization, SSE, error mapping, dan cancellation.

### TypeScript/React
- Gunakan tipe eksplisit untuk payload backend dan event.
- Hindari `any` tanpa justifikasi.
- Cocokkan setiap delta dengan `requestId` dan `conversationId`.
- Bersihkan listener event ketika komponen/unmount lifecycle berakhir.
- Jangan menyimpan API key di browser storage.
- Jangan render raw HTML dari AI.
- Pastikan state loading/error/cancel tidak membuat tombol terkunci selamanya.

## Definition of done untuk setiap perubahan

Perubahan baru dianggap selesai bila:

- Perilaku implementasi sesuai requirement yang relevan.
- Build/format/test yang tersedia telah dijalankan.
- Jalur error utama telah dipertimbangkan.
- Secret dan data pengguna tidak bocor ke log/event.
- Tidak ada UI palsu atau TODO yang menyaru sebagai fitur selesai.
- Dokumen yang terpengaruh diperbarui.
- Risiko baru dicatat dalam `docs/09-decisions.md` jika relevan.

## Perintah verifikasi

Gunakan perintah yang disediakan oleh versi Wails dan package manager project. Setelah scaffold dibuat, dokumentasikan perintah resmi yang berfungsi di README. Umumnya, periksa setidaknya:

- Go formatting (`gofmt`).
- Go tests (`go test ./...`).
- Frontend lint/typecheck/test jika dikonfigurasi.
- Build development.
- Build release Windows dari clean checkout.

Jangan mengklaim tes lulus jika tidak benar-benar dijalankan.

## Pelaporan pekerjaan

Setelah tiap milestone, berikan ringkasan:
1. File/fitur yang berubah.
2. Perintah build/test yang dijalankan dan hasilnya.
3. Yang belum dikerjakan atau diketahui gagal.
4. Keputusan yang memerlukan input pengguna.

Jangan mengerjakan perubahan besar yang tidak terkait dalam satu commit. Jangan menghapus database/log pengguna saat proses development tanpa konfirmasi dan backup yang sesuai.
