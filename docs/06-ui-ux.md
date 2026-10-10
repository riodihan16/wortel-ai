# 06 — UI/UX Desktop Companion

## 1. Arah visual

Karakter 2D kecil dengan tampilan modern, hangat, dan tidak terlalu ramai. UI di sekelilingnya netral, dengan kontras jelas dan aksen warna tunggal.

**Tujuan:** saat idle, Wortel AI terasa seperti desktop pet; ketika diklik, ia berubah menjadi chat tool yang praktis. Karakter bukan pengganti panel chat dan bukan elemen dekorasi yang wajib mengambil perhatian.

## 2. Mode tampilan

### Mode A — Pet

Elemen:
- Karakter 2D.
- Indikator status yang sangat kecil bila diperlukan.
- Tooltip atau hint singkat ketika hover, bukan dialog mengganggu.
- Context menu sederhana: Open chat, Hide, Settings, Exit (menu dapat ditunda sampai tray tersedia).

Interaksi:
- Click: buka chat.
- Drag: pindahkan posisi.
- Klik kanan: menu opsional.
- Escape: tidak boleh membuat aplikasi hilang tanpa cara memulihkan.
- Shortcut global: toggle chat (P1).

Batas desain:
- Tidak ada teks permanen yang menutupi desktop.
- Tidak memutar suara.
- Tidak menunjukkan “thinking” terus-menerus saat idle.
- Status idle tidak memerlukan request jaringan.

### Mode B — Chat

Elemen utama:
1. Header kecil: karakter/avatar, nama, model aktif, tombol minimize/back dan settings.
2. Area pesan yang dapat scroll.
3. Status streaming yang tidak mengubah layout secara besar.
4. Composer multiline.
5. Send/Stop button yang stateful.
6. Pemilih conversation sederhana; boleh menjadi sidebar kecil atau dropdown tergantung ukuran.

Perilaku:
- Fokus input setelah chat dibuka jika tidak menimbulkan perpindahan fokus yang mengejutkan.
- Enter kirim; Shift+Enter newline, kecuali setting alternatif ditambahkan.
- Saat request berjalan, tombol Send berubah menjadi Stop atau Stop tersedia bersebelahan.
- Menekan Escape dapat mengecilkan panel ke pet, tetapi tidak boleh membatalkan request secara diam-diam.
- Ketika menutup panel dan generation masih berlangsung, generation tetap berjalan secara default. Pengguna dapat kembali melihatnya. Sediakan cara untuk membatalkan secara eksplisit.

### Mode C — Settings

Kelompok:
- Provider & model.
- Appearance.
- Behavior/keyboard shortcut.
- Data & privacy.
- About/version.

Provider form minimum:
- Display name.
- Base URL.
- Daftar Model ID: tambah/hapus model; radio memilih model aktif.
- API key input dengan reveal/hide.
- Save.
- Test connection.
- Status aktif.

Chat header menampilkan pemilih model (dropdown) bila provider aktif punya lebih dari satu model, sehingga model bisa diganti tanpa membuka Settings.

Jangan menampilkan API key lama penuh setelah settings dibuka. Tampilkan “credential tersimpan” dan sediakan aksi replace/remove.

## 3. State dan status karakter

| State | Trigger | Visual yang disarankan |
|---|---|---|
| `idle` | Tidak ada interaksi/request | Gerakan kecil berkala atau pose statis. |
| `hover` | Pointer berada di atas karakter | Reaksi singkat, bukan animasi panjang. |
| `thinking` | Request diterima dan menunggu token | Pose/indikator ringan; mulai hanya saat request. |
| `speaking` | Delta teks masuk | Animasi sangat sederhana atau indikator pulse terbatas. |
| `error` | Request gagal | Ekspresi/indikator error; jangan hanya mengandalkan warna. |
| `offline` | Provider belum dikonfigurasi atau tes koneksi gagal | Hint saat dibuka, bukan popup berulang. |

Jangan menjanjikan bahwa state karakter mencerminkan “emosi AI” secara nyata. Ini hanya status UI.

## 4. Ukuran dan jendela

Nilai awal adalah hipotesis untuk diuji, bukan requirement pixel-perfect:
- Mode pet: window compact, sekitar 120–200 px per sisi sesuai aset karakter.
- Mode chat: panel sekitar 360–460 px lebar dan 480–680 px tinggi.
- Panel harus tetap muat pada layar beresolusi rendah dan scaling Windows tinggi.
- Simpan posisi dengan koordinat aman; jika monitor/area kerja berubah atau koordinat tidak valid, posisikan kembali ke area yang terlihat.

Mendukung multi-monitor dan DPI scaling perlu dites. Jangan mengasumsikan koordinat satu monitor selalu valid setelah perangkat display berubah.

## 5. Prinsip micro-interaction

- Animasi pendek (kurang lebih 100–250 ms untuk transisi UI, kecuali gerakan karakter khusus).
- Hindari efek glow/blur besar yang membutuhkan repaint intensif.
- Hover feedback jelas tetapi tidak mengubah layout.
- Semua tombol memiliki label/tooltip atau accessible name.
- Loading indicator tidak boleh menutupi teks yang sudah diterima.
- Saat error, pertahankan teks pengguna dan teks parsial agar tidak hilang.
- Tombol destructive seperti Delete conversation diberi konfirmasi.

## 6. Aksesibilitas

- Navigasi keyboard untuk seluruh fungsi penting.
- Focus ring terlihat.
- Kontras teks memadai dalam dark dan light mode.
- Ukuran font dapat dibaca dan scaling UI mengikuti DPI bila memungkinkan.
- State error/success tidak disampaikan melalui warna saja.
- Reduced-motion preference dapat dipertimbangkan untuk versi setelah MVP; setidaknya jangan memakai animasi yang tidak bisa dihentikan dan berlebihan.

## 7. Markdown dan code block

- Gunakan renderer yang tidak mengeksekusi HTML mentah.
- Sanitasi atau nonaktifkan raw HTML.
- Link eksternal dibuka melalui browser default setelah validasi skema (`https`/`http`); jangan mengaktifkan protokol arbitrer seperti `file:`.
- Code block memiliki tombol Copy.
- Teks panjang tidak boleh memperlebar panel; gunakan wrapping atau horizontal scroll di code block.
- Token streaming harus tetap terbaca meskipun Markdown sementara belum lengkap (misalnya fenced code belum ditutup).

## 8. Empty, loading, dan error states

### Belum ada percakapan

Tampilkan sapaan pendek dan satu contoh prompt, tetapi jangan mengirimnya otomatis.

### Belum ada provider

Tampilkan instruksi untuk menambahkan endpoint/model/API key. Jangan tampilkan chat seolah-olah sudah siap.

### Loading

Tampilkan state yang tenang. Hindari spinner besar atau progress persentase palsu untuk request yang durasinya tidak diketahui.

### Error

Sertakan:
- Kalimat singkat yang menjelaskan masalah.
- Kategori atau HTTP status jika aman.
- Langkah berikutnya yang spesifik.
- Retry manual bila masuk akal.

## 9. Draft design tokens

Bukan keputusan final; sesuaikan setelah tampilan awal dibuat.

- Radius panel: 12–16 px.
- Radius bubble: 10–14 px.
- Spacing utama: kelipatan 4/8 px.
- Font UI: system UI stack untuk mengurangi kebutuhan bundling font.
- Background dark: charcoal netral; hindari hitam murni pada semua permukaan.
- Accent: indigo/cyan.
- Border: tipis dan kontras rendah.
- Shadow: sederhana, tanpa blur berlapis berlebihan.

## 10. Uji UX sebelum rilis

Minta pengguna melakukan tugas tanpa panduan:
1. Membuka chat dari karakter.
2. Menambah provider.
3. Menguji koneksi.
4. Mengirim pesan dan membatalkan jawaban.
5. Membuka percakapan lama.
6. Menyembunyikan dan memulihkan companion.
7. Menemukan cara keluar sepenuhnya dari aplikasi.

Catat kebingungan, klik yang tidak terlihat, perilaku window di tepi layar, dan apakah karakter mengganggu kerja biasa.
