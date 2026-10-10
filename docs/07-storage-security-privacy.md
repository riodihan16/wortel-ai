# 07 — Penyimpanan, Keamanan, dan Privasi

## 1. Prinsip

Wortel AI berjalan di komputer pengguna dan menghubungi provider AI yang dipilih pengguna. Karena API key serta percakapan dapat bersifat sensitif, keamanan bukan fitur tambahan.

Aturan mutlak untuk MVP:
- Secret tidak ada di source code, SQLite, JSON preferences, maupun log.
- Isi percakapan hanya dikirim ketika pengguna mengirim pesan.
- Tidak ada telemetry atau sinkronisasi cloud otomatis.
- AI output dianggap sebagai teks tidak tepercaya, bukan instruksi bagi sistem operasi.
- Semua data memiliki lokasi lokal yang terdokumentasi.

## 2. Lokasi data

Gunakan direktori aplikasi per-user di `%LOCALAPPDATA%` (atau lokasi yang sesuai dengan konvensi Wails/platform) untuk data yang dapat ditulis. Jangan menyimpan database di folder instalasi atau sebelah executable karena folder tersebut bisa read-only atau berada di lokasi portable yang dibagikan.

Struktur contoh:

```text
%LOCALAPPDATA%\Wortel AIAI\
├── wortel.db
├── settings.json       # hanya bila ada kebutuhan; tanpa secret
└── logs\               # minimal dan disanitasi, bila logging file diaktifkan
```

Nama direktori final harus konsisten dan terdokumentasi. Jika mendukung mode portable di masa depan, desain lokasi data sebagai opsi terpisah; jangan secara diam-diam mencampur data portable dan data install.

## 3. Skema SQLite awal

### `conversations`

Kolom kandidat:
- `id` TEXT PRIMARY KEY (UUID).
- `title` TEXT NOT NULL.
- `created_at` TEXT/TIMESTAMP NOT NULL.
- `updated_at` TEXT/TIMESTAMP NOT NULL.
- `archived_at` nullable bila archive masuk scope.

### `messages`

Kolom kandidat:
- `id` TEXT PRIMARY KEY.
- `conversation_id` TEXT NOT NULL FK.
- `role` TEXT NOT NULL (`user`, `assistant`, dan `system` jika digunakan).
- `content` TEXT NOT NULL.
- `status` TEXT NOT NULL (`pending`, `streaming`, `complete`, `cancelled`, `failed`).
- `created_at` TEXT/TIMESTAMP NOT NULL.
- `updated_at` nullable.
- `provider_id` nullable untuk jejak provider/model.
- `model` nullable.
- `sequence` INTEGER untuk urutan eksplisit bila dibutuhkan.

Tambahkan indeks pada `messages(conversation_id, created_at)` atau kombinasi yang sesuai query nyata. Aktifkan foreign keys dan gunakan transaksi.

### `provider_profiles`

Kolom kandidat:
- `id` TEXT PRIMARY KEY.
- `name` TEXT NOT NULL.
- `base_url` TEXT NOT NULL.
- `model` TEXT NOT NULL (legacy; rekan dari model aktif).
- `models` TEXT NOT NULL (JSON array model ID; ditambahkan di v3).
- `active_model` TEXT NOT NULL (model dipakai untuk chat; ditambahkan di v3).
- `credential_ref` TEXT nullable.
- `is_enabled` INTEGER NOT NULL.
- `created_at`, `updated_at`.

**Dilarang:** kolom `api_key` atau value secret lain.

### `app_settings`

Bisa berupa key/value atau kolom bertipe yang memiliki schema jelas. Contoh: theme, pet size, opacity, posisi jendela, active provider ID, shortcut. Jangan simpan kata sandi atau token di sini.

### Migration

- Simpan versi schema.
- Jalankan migration pada startup sebelum repository digunakan.
- Migration harus gagal secara eksplisit; jangan diam-diam membuat ulang DB dan menghapus chat.
- Uji upgrade dari versi DB sebelumnya saat schema berubah.

## 4. Credential storage

Preferensi utama: **Windows Credential Manager** atau solusi secret storage OS yang dipilih setelah spike dan review dependency.

Kontrak konseptual:

```go
type CredentialStore interface {
    Set(ctx context.Context, ref string, secret []byte) error
    Get(ctx context.Context, ref string) ([]byte, error)
    Delete(ctx context.Context, ref string) error
}
```

Ini hanya gambaran interface; implementasi final harus memperhatikan library/Win32 API yang dipakai.

Aturan penggunaan:
- API key ditulis saat pengguna menekan Save.
- Secret diambil tepat sebelum request diperlukan.
- Secret tidak dimasukkan ke struct provider publik yang di-serialize.
- Secret tidak dikembalikan melalui method binding ke React.
- Saat API key diganti, update credential secara aman dan hapus credential lama bila sudah tidak direferensikan.
- Saat provider dihapus, credential terkait dibersihkan setelah memastikan tidak dipakai profile lain.
- Error pada credential store tidak boleh berisi secret.

Jika credential OS belum dapat diimplementasikan dengan aman, lebih baik tunda penyimpanan persistent API key atau gunakan solusi sementara yang jelas diberi label; jangan diam-diam menyimpan plaintext.

## 5. Privasi request

Sebelum mengirim pesan, aplikasi mengirim ke provider:
- Model ID (model aktif provider; hanya model itu yang dikirim, tidak seluruh daftar).
- Isi pesan pengguna dan konteks percakapan terpilih.
- Parameter request yang diperlukan.
- API key pada header autentikasi.

Aplikasi tidak boleh mengirim:
- Semua conversation lain.
- Path file, hostname, username Windows, screenshot, atau data perangkat kecuali memang secara eksplisit dibutuhkan dan disetujui di masa depan.
- API key di URL, query string, event frontend, atau diagnostic report.

UI Settings/About harus menjelaskan bahwa provider menerima teks yang dikirim pengguna sesuai kebijakan privasi provider itu sendiri.

## 6. Base URL dan risiko koneksi

Karena pengguna dapat memasukkan URL endpoint sendiri, validasi harus mencegah konfigurasi yang tidak sengaja berbahaya:
- Hanya scheme `http`/`https`.
- Tidak ada username/password pada URL.
- Tampilkan warning untuk HTTP non-loopback.
- Hindari meneruskan Authorization ke host berbeda setelah redirect.
- Jangan akses URL dari asset/Markdown secara otomatis.
- Jangan mengubah respons provider menjadi request lain ke tujuan yang tidak dipilih pengguna.

Permintaan ke `localhost` atau alamat privat bisa valid untuk gateway pribadi pengguna, tetapi jangan izinkan dari input yang datang dari AI output atau halaman Markdown. Hanya pengaturan provider pengguna yang boleh menentukan tujuan API.

## 7. Markdown dan konten tidak tepercaya

AI dapat menghasilkan konten yang terlihat seperti HTML atau link berbahaya.
- Nonaktifkan raw HTML atau sanitasi secara ketat.
- Buka link hanya jika scheme termasuk `https` atau `http`.
- Jangan menjalankan JavaScript dari konten chat.
- Jangan membuka `file:`, `ms-settings:`, `powershell:`, `cmd:`, atau custom URI dari hasil AI.
- Copy ke clipboard hanya lewat aksi pengguna.
- Jangan memasukkan hasil AI ke command line atau shell pada MVP.

## 8. Logging dan crash reporting

Logging lokal boleh membantu debugging, tetapi minimum:
- Log event lifecycle, durasi request, status HTTP, kategori error, dan versi aplikasi.
- Redaksi `Authorization`, API key, credential, dan query sensitif.
- Hindari menyimpan body prompt/response secara default.
- Jika log berisi path atau hostname, tinjau apakah benar-benar diperlukan.
- Rotasi log dan batasi ukuran jika file logging tersedia.
- Jangan upload log atau crash dump otomatis.

## 9. Penghapusan dan retensi

- Pengguna dapat menghapus satu percakapan.
- Pengguna dapat menghapus semua riwayat dalam versi berikutnya; bila masuk MVP, minta konfirmasi jelas.
- Hapus percakapan dari SQLite tidak menjamin secure erase dari media penyimpanan; jangan menjanjikan secure deletion.
- Fitur export/import dapat dipertimbangkan nanti dengan peringatan bahwa file ekspor berisi data sensitif.

## 10. Threat model ringkas

| Ancaman | Kontrol |
|---|---|
| API key bocor lewat DB/log | OS credential storage, redaksi, tidak menampilkan secret. |
| Redirect endpoint mencuri credential | Kontrol redirect dan larangan meneruskan Authorization ke host baru. |
| Markdown berisi script | Renderer aman, raw HTML nonaktif, link scheme allowlist. |
| AI menyuruh menjalankan command | Tidak ada eksekusi command dalam MVP. |
| Database rusak | Transaksi, migration, backup/manual recovery yang terdokumentasi. |
| Pengguna salah memasukkan endpoint HTTP | Peringatan transport tidak aman. |
| Crash saat streaming | Context cancellation, penyimpanan status parsial, cleanup resource. |

Ini bukan audit keamanan formal. Sebelum menambah automation/agent atau mendistribusikan aplikasi luas, lakukan threat model lanjutan.
