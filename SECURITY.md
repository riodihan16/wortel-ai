# Kebijakan Keamanan

## Melaporkan kerentanan

**Jangan buka issue publik** untuk kerentanan keamanan. Laporkan secara privat melalui fitur **Security → Report a vulnerability** di GitHub (Security Advisories) pada repo ini, atau hubungi pengelola lewat alamat yang tercantum di profil pemilik (`Ridho Dwi Ramadhan`).

Sertakan bila memungkinkan:

- Deskripsi masalah dan dampaknya.
- Langkah reproduksi atau bukti konsep.
- Versi (commit SHA atau tag) dan lingkungan Windows.
- Saran perbaikan bila ada.

Kami berupaya membalas dalam beberapa hari kerja dan akan mengoordinasikan pengungkapan sebelum perbaikan dirilis.

## Model ancaman

Wortel AI adalah aplikasi desktop lokal yang mengirim permintaan ke endpoint AI yang dikonfigurasi pengguna. Batas kepercayaan utama:

- **API key** adalah rahasia. Disimpan di Windows Credential Manager, **tidak** di SQLite, log, localStorage, atau event frontend. Ambil tepat sebelum request, lalu wipe salinannya dari memori.
- **Respons AI adalah input tidak tepercaya.** Markdown dirender lewat `react-markdown` + `rehype-sanitize` tanpa raw HTML; link hanya `http`/`https`.
- **Redirect lintas host tidak membawa header `Authorization`** (di-drop oleh HTTP client yang di-hardening).
- Aplikasi **tidak** menjalankan perintah shell, membaca screenshot, clipboard, atau file pengguna secara otomatis, dan tidak memiliki telemetry/analytics/cloud sync.
- Semua request jaringan memakai `context`/cancellation dan selalu menutup body respons.

## Di luar cakupan

- Endpoint AI yang dikonfigurasi pengguna (ketersediaan, kebijakan privasi, atau penanganan data oleh pihak ketiga).
- Kebocoran yang berasal dari kompromi sistem operasi atau akun pengguna.
- Serangan yang membutuhkan akses fisik ke perangkat.

## Versi yang didukung

Hanya rilis terbaru pada branch `master` yang didukung dengan patch keamanan.

Catatan jujur: ini **bukan** audit keamanan formal. Sebelum menambahkan automation/agent atau mendistribusikan aplikasi secara luas, lakukan threat model lanjutan; lihat `docs/07-storage-security-privacy.md`.
