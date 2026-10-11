# 02 — Kebutuhan dan Acceptance Criteria MVP

## 1. Cara membaca dokumen

Prioritas:

- **P0:** syarat rilis MVP.
- **P1:** penting tetapi dapat menyusul bila ada blocker teknis.
- **P2:** eksplorasi setelah MVP.

Acceptance criteria sebaiknya dijadikan tes manual atau automated test bila memungkinkan. Perilaku UI spesifik masih boleh berubah setelah prototipe, tetapi hasil yang harus dicapai jangan dihilangkan tanpa keputusan tertulis.

## 2. Functional requirements

### FR-001 — Menjalankan aplikasi Windows (P0)

Aplikasi dapat dibangun menjadi executable Windows. Pengguna tidak perlu memasang Node.js atau Go untuk menjalankan build rilis.

**Kriteria penerimaan**
- Build release menghasilkan executable yang dapat dijalankan pada instalasi Windows yang didukung.
- Ketergantungan runtime seperti WebView2 ditangani dengan pesan yang dapat dimengerti jika tidak tersedia.
- Aplikasi dapat ditutup sepenuhnya; tidak meninggalkan proses Wortel AI yang tidak diperlukan.
- Build development dan release dipisahkan dengan jelas.

### FR-002 — Tampilan companion (P0)

Saat startup, aplikasi menampilkan satu karakter 2D di jendela frameless dengan ukuran kecil.

**Kriteria penerimaan**
- Karakter terlihat pada background terang maupun gelap.
- Jendela dapat dipindahkan dengan drag pada area karakter.
- Karakter tidak terpotong pada batas layar setelah dipindahkan.
- Companion dapat dikecilkan/disembunyikan dan dipulihkan melalui cara yang terdokumentasi.
- Status animasi tidak memicu proses CPU yang terus tinggi saat idle.
- Transparansi, klik pada area transparan, dan always-on-top harus diuji di Windows; jangan menyatakan fitur ini selesai sebelum dites di build nyata.

### FR-003 — Membuka dan menutup chat (P0)

Klik karakter atau shortcut yang dikonfigurasi membuka mode chat. Pengguna dapat kembali ke mode karakter tanpa kehilangan draft secara tidak sengaja.

**Kriteria penerimaan**
- Satu jendela dipakai untuk mode karakter dan mode chat pada MVP, kecuali spike membuktikan keputusan ini tidak layak.
- Perubahan mode tidak menggandakan jendela atau membuat koneksi chat duplikat.
- Draft dipertahankan saat panel ditutup/diciutkan, kecuali pengguna secara eksplisit menghapusnya.
- Fokus keyboard ditempatkan ke input chat saat panel dibuka bila sesuai.

### FR-004 — Kirim dan tampilkan pesan (P0)

Pengguna dapat mengirim pesan teks ke provider aktif dan menerima jawaban.

**Kriteria penerimaan**
- Pesan pengguna tampil setelah dikirim.
- Input kosong/whitespace tidak dikirim.
- Enter mengirim pesan; Shift+Enter membuat baris baru, atau pola alternatif yang konsisten dan dapat diketahui pengguna.
- Tombol kirim tidak menghasilkan duplikasi ketika diklik berkali-kali selama request aktif.
- Pesan berurutan dijaga per percakapan dan role-nya benar.
- Markdown umum (paragraf, list, code block, link) ditampilkan dengan aman.

### FR-005 — Streaming dan pembatalan (P0)

Jawaban streaming tampil seiring diterima. Pengguna dapat membatalkan request berjalan.

**Kriteria penerimaan**
- UI memiliki state `sending`, `streaming`, `complete`, `cancelled`, dan `error` atau padanan yang jelas.
- Token/chunk diteruskan ke frontend melalui event Wails.
- Tombol Stop memicu pembatalan `context.Context` di backend dan menutup HTTP stream jika memungkinkan.
- Pembatalan tidak membuat pesan asisten palsu yang dianggap selesai.
- Bila koneksi putus saat streaming, teks parsial dipertahankan dan status diberi label gagal/terputus.
- UI tetap responsif selama streaming.

### FR-006 — Provider dan model (P0)

Pengguna dapat membuat, mengedit, memilih, menguji, dan menghapus konfigurasi provider.

**Field minimum**
- Nama tampilan.
- Base URL.
- Daftar Model ID (minimal satu; model aktif dipilih dari daftar).
- API key (secret).
- Provider aktif.

**Kriteria penerimaan**
- Base URL hanya menerima `http` atau `https`; `https` disarankan sebagai default.
- Validasi URL dan field wajib dilakukan sebelum request.
- Pengguna dapat menyimpan lebih dari satu model per provider dan mengganti model aktif tanpa mengedit ulang provider (ditambahkan 2026-10-10; lihat ADR-028).
- Tombol Test Connection memberi sukses/gagal yang bermakna, tanpa menampilkan API key.
- Mengganti provider tidak menghapus percakapan yang ada.
- Secret tidak dikembalikan ke frontend setelah disimpan; UI hanya menampilkan status configured/not configured atau mask.
- Provider yang gagal dihapus ketika sedang aktif harus mengharuskan pengguna memilih provider lain atau mengonfirmasi dampaknya.

### FR-007 — Penyimpanan percakapan (P0)

Percakapan disimpan secara lokal.

**Kriteria penerimaan**
- Percakapan dapat dibuat, diberi judul, dibuka, dan dihapus.
- Pesan tersimpan dengan role, timestamp, urutan, status, dan isi.
- Perubahan state tidak merusak percakapan terdahulu.
- Pesan yang gagal/cancel memiliki status yang jelas.
- Pengguna tidak diwajibkan login atau membuat akun Wortel AI.

### FR-008 — Pengaturan aplikasi (P1)

Pengaturan minimal: theme, ukuran karakter, preferensi startup, dan shortcut bila implementasi platform tersedia.

**Kriteria penerimaan**
- Pengaturan tersimpan lokal dan berlaku kembali saat aplikasi dibuka.
- Nilai ukuran/opacity memiliki batas minimum-maksimum yang aman.
- Bila startup belum didukung, UI tidak boleh menampilkan toggle palsu.

### FR-009 — Error handling (P0)

Aplikasi memberi pesan yang bisa ditindaklanjuti untuk error umum.

**Error minimum**
- Tidak ada provider aktif.
- API key belum diisi / invalid (jangan pernah mengklaim invalid berdasarkan status yang ambigu).
- URL tidak valid atau koneksi gagal.
- HTTP 401/403, 404, 429, 5xx.
- Model tidak ditemukan atau parameter tidak didukung.
- Timeout, koneksi terputus, dan stream berakhir sebelum selesai.
- Database gagal dibuka atau ditulis.

**Kriteria penerimaan**
- Detail teknis yang aman dapat masuk log; secret dan isi sensitif tidak otomatis masuk log.
- Pengguna dapat mencoba kembali secara eksplisit.
- Request yang telah mulai menghasilkan token tidak diulang otomatis tanpa persetujuan.

### FR-010 — Shortcut global (P1)

Shortcut global membuka/menyembunyikan chat saat aplikasi berjalan.

**Kriteria penerimaan**
- Shortcut default tidak boleh menimpa shortcut sistem umum tanpa peringatan.
- Registrasi yang gagal diberitahukan kepada pengguna.
- Shortcut dapat diubah atau dinonaktifkan jika API platform mendukungnya.
- Shortcut dibatalkan ketika aplikasi keluar.

### FR-011 — System tray dan startup (P1)

Aplikasi dapat tetap berjalan dengan ikon tray, dan pengguna dapat mengatur startup bersama Windows.

**Kriteria penerimaan**
- Menu tray memiliki Open/Show, Hide, Settings, dan Exit (label dapat disesuaikan).
- Exit menutup semua jendela dan membatalkan request aktif.
- Startup default mengikuti keputusan produk yang eksplisit; disarankan **off** sampai pengguna mengaktifkannya.

### FR-012 — Lampiran gambar dan berkas (P1)

Pengguna dapat melampirkan gambar (png/jpeg/webp), PDF, dan berkas teks (txt/md/csv/json) pada pesan, lewat tombol klip, Ctrl+V gambar, atau drag-drop ke panel chat (ditambahkan 2026-10-11; lihat ADR-054).

**Kriteria penerimaan**
- Maksimal 4 lampiran per pesan; gambar ≤ 8 MB, PDF ≤ 15 MB, teks ≤ 256 KB, total ≤ 20 MB; tipe ditentukan dari isi berkas, bukan ekstensi.
- Pesan boleh tanpa teks bila ada lampiran; tanpa keduanya tidak dikirim.
- Lampiran tersimpan di riwayat lokal dan tampil kembali saat percakapan dibuka; hapus percakapan menghapus lampirannya.
- Hanya 3 pesan user terakhir yang mengirim lampiran penuh ke model, dengan total ≤ 20 MB per request; yang lebih lama diganti placeholder teks.
- Provider yang menolak gambar/berkas menghasilkan pesan "Model ini tidak menerima gambar/berkas", bukan error generik.
- Berkas hanya dibaca dari aksi eksplisit pengguna; path penuh dan isi berkas tidak masuk log.

### FR-013 — Pembuatan gambar (P1)

Pengguna dapat membuat gambar dari sebuah prompt memakai model gambar yang dikonfigurasi pada provider aktif; hasilnya tersimpan lokal dan tampil di dalam percakapan (ditambahkan 2026-10-11; lihat ADR-055).

**Kriteria penerimaan**
- Model gambar dipilih di Pengaturan (kolom "Gambar" pada daftar model provider); tidak ada ID model yang di-hardcode. Model yang tidak ada di daftar ditolak/dibuang.
- Prompt dikirim ke `POST /v1/images/generations` dari backend Go dengan timeout panjang (5 menit) dan pembatalan via context.
- Kedua bentuk respons provider ditangani: `b64_json` didekode lokal, `url`/data URL diunduh backend (dibatasi ukurannya, hanya png/jpeg/webp) — WebView tidak pernah mengambil URL provider atau menerima base64 lewat IPC.
- Hasil disimpan lokal sebagai pesan asisten (kolom `image_mime`/`image_data` di `messages`) dan disajikan lewat route `/attachments/{id}`; bertahan setelah restart.
- Model gambar di chat tidak pernah mengirim ulang gambar asisten ke model chat (gambar hanya pada baris asisten), sehingga tidak ditagih ulang.
- Prompt bar menampilkan state menggambar yang berbeda, tombol hentikan (dan Escape), serta pesan error inline yang aman.
- 402/400/429/413/503 dan MIME tak didukung dipetakan ke pesan Indonesia yang spesifik; tidak ada retry otomatis (reservasi biaya mungkin sudah terjadi).
- Biaya pembuatan gambar selalu dari saldo PAYG, **bukan** paket chat; UI menyatakan hal ini.
- Menghapus percakapan menghapus gambar hasilnya; CancelAll saat shutdown menghentikan pembuatan yang berjalan.

## 3. Non-functional requirements

### NFR-001 — Performa idle (P0)

- Tidak ada polling jaringan saat idle.
- Tidak ada model AI yang diload lokal.
- Tidak ada timer render berfrekuensi tinggi hanya untuk animasi karakter.
- Ukur seluruh pohon proses aplikasi, termasuk WebView2.
- Angka budget RAM ditetapkan setelah baseline prototipe; target awal bersifat sementara, bukan janji produk.

### NFR-002 — Responsivitas (P0)

- UI tidak memblokir input saat request berjalan.
- I/O jaringan dan database dilakukan di backend/asynchronous path yang sesuai.
- Streaming di-batch jika menerima chunk sangat kecil agar render tidak berlebihan, tanpa menambah latensi yang terasa.

### NFR-003 — Privasi dan keamanan (P0)

- API key disimpan di storage secret OS bila tersedia.
- API key tidak disimpan di database konfigurasi, source code, atau file log.
- Percakapan hanya dikirim ke endpoint yang dipilih ketika pengguna menekan kirim.
- Tidak ada screenshot capture, telemetry, analytics, atau cloud sync pada MVP.
- Rendering Markdown disanitasi dan protokol tautan dibatasi.

### NFR-004 — Ketahanan data (P0)

- Gunakan transaksi untuk perubahan yang terkait.
- Skema SQLite memiliki versi/migration.
- Kerusakan DB atau permission error dijelaskan dengan aman.
- Penghapusan percakapan meminta konfirmasi jika tindakan sulit dipulihkan.

### NFR-005 — Maintainability (P0)

- Backend AI, storage, platform Windows, dan UI dipisahkan.
- Tidak membuat abstraksi plugin/agent sebelum dibutuhkan.
- Fungsi yang berpotensi gagal mengembalikan error eksplisit.
- Formatting, lint, unit test, dan build test masuk checklist sebelum rilis.

### NFR-006 — Portabilitas build (P1)

- Hasil build rilis tidak memerlukan Node.js/Go.
- Aset frontend dibundel ke binary bila sesuai mekanisme Wails.
- Lokasi data pengguna tidak hardcoded ke direktori project atau executable.
- Dokumentasi menyebut arsitektur Windows yang didukung.

## 4. Kriteria “MVP done”

Semua P0 lulus di satu mesin Windows target dan diuji ulang setelah clean build. Tidak ada API key di log/database, streaming bisa dibatalkan, percakapan bertahan setelah restart, dan idle behavior telah diukur. Isu P1 boleh ditunda dengan alasan tertulis.
