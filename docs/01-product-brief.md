# 01 — Product Brief

## 1. Ringkasan

**Wortel AI** adalah desktop companion Windows dengan karakter 2D kecil yang tinggal di desktop dan dapat membuka panel chat AI. Produk ini mengutamakan pengalaman interaksi yang cepat dan personal tanpa menjadi aplikasi besar yang selalu meminta perhatian.

Versi pertama berfokus pada **chat AI**. Integrasi automation, suara, membaca layar, menjalankan command, dan agent mandiri bukan bagian dari MVP.

## 2. Masalah yang ingin diselesaikan

Pengguna yang sering memakai AI harus berulang kali berpindah ke browser atau aplikasi chat. Itu memecah alur kerja dan membuat bantuan AI tidak terasa dekat dengan desktop.

Wortel AI mencoba mengurangi gesekan tersebut dengan:

- Karakter kecil yang selalu mudah ditemukan.
- Cara membuka chat yang cepat, baik melalui klik maupun shortcut global.
- Dukungan banyak endpoint yang mengikuti pola OpenAI-compatible.
- Pengalaman desktop native-ish dengan ukuran jendela kecil dan idle overhead terkontrol.

## 3. Visi produk

> AI yang terasa dekat seperti teman desktop, mudah dipanggil, bisa dikonfigurasi, dan tidak mendominasi komputer pengguna.

## 4. Target pengguna awal

### Pengguna utama

Pengguna Windows yang sering memakai AI untuk bertanya, menjelaskan konsep, menulis, merangkum, atau membantu coding; bersedia memasukkan endpoint dan API key sendiri.

### Kebutuhan utama

- Membuka chat dalam beberapa detik tanpa mencari tab browser.
- Menggunakan provider/model favorit dan bisa menggantinya.
- Menjaga percakapan dan pengaturan tersimpan secara lokal.
- Membiarkan companion menganggur tanpa penggunaan CPU yang tidak perlu.

### Bukan target MVP

- Pengguna yang butuh model AI sepenuhnya offline.
- Pengguna yang membutuhkan sinkronisasi cloud lintas perangkat.
- Pengguna yang ingin autonomous agent mengendalikan Windows.
- Pengguna yang membutuhkan voice assistant atau pemrosesan gambar langsung.

## 5. Prinsip produk

1. **Lightweight first.** Hindari proses latar belakang, timer, animasi berat, dan dependency yang tidak menambah manfaat MVP.
2. **User controls the AI.** Pengguna secara eksplisit mengirim pesan; tidak ada pengiriman otomatis atau perekaman layar.
3. **Provider-agnostic.** Informasi provider bukan hardcoded; kemampuan opsional dinegosiasikan atau dikonfigurasi.
4. **Local-first.** Riwayat serta preferensi disimpan di mesin pengguna, bukan dikirim ke server Wortel AI.
5. **Progressive enhancement.** Tambahkan fitur hanya setelah alur dasar teruji.
6. **Transparent errors.** Jelaskan masalah seperti URL salah, API key invalid, model tidak ditemukan, limit terlampaui, atau koneksi terputus.
7. **Accessible by default.** Teks mudah dibaca, fokus keyboard jelas, shortcut tidak wajib digunakan, serta karakter dapat disembunyikan.

## 6. Pengalaman inti (core loop)

1. Wortel AI berjalan dan karakter kecil tampak di desktop.
2. Pengguna mengklik karakter atau menekan shortcut global.
3. Jendela berpindah ke mode chat.
4. Pengguna memilih provider/model aktif jika perlu.
5. Pengguna menulis dan mengirim pesan.
6. Jawaban tampil bertahap dan dapat disalin.
7. Percakapan tersimpan lokal.
8. Pengguna mengecilkan chat menjadi karakter atau keluar dari aplikasi.

## 7. Sukses MVP

MVP dianggap sukses bila pengguna dapat menjalankan alur inti berkali-kali tanpa crash atau konfigurasi yang hilang, dan aplikasi terasa ringan saat idle.

Metrik yang harus diukur, bukan diasumsikan:

- Waktu startup dingin dan hangat.
- Penggunaan RAM seluruh pohon proses terkait, termasuk proses WebView2.
- Rata-rata CPU saat idle selama interval pengukuran yang tetap.
- Latensi dari kirim pesan sampai token pertama (pisahkan waktu provider dari overhead aplikasi jika memungkinkan).
- Tingkat keberhasilan pengiriman, pembatalan, dan pemulihan error.
- Banyaknya langkah untuk mengganti provider dan kembali ke chat.

## 8. Prinsip visual

- Karakter 2D imut dan sederhana, bukan avatar 3D.
- Siluet mudah dikenali pada ukuran kecil.
- Palet warna awal: charcoal/dark neutral dengan satu aksen indigo atau cyan; tetap siapkan kontras untuk light theme.
- Animasi memiliki status diskret: `idle`, `thinking`, `speaking`, `error` atau `offline`.
- Gerakan tidak boleh berulang dengan frame rate tinggi tanpa kebutuhan.
- Panel chat lebih memprioritaskan keterbacaan dibanding dekorasi.

## 9. Risiko produk

| Risiko | Dampak | Mitigasi |
|---|---|---|
| Companion menghalangi konten lain | Mengganggu pemakaian harian | Bisa dipindah, opacity/ukuran dapat diatur, dan dapat disembunyikan. |
| Transparansi window sulit di Windows | MVP tertunda atau hit-testing tidak nyaman | Buat spike teknis sebelum menyelesaikan desain karakter. |
| Perbedaan provider API | Model tertentu gagal meskipun endpoint terlihat kompatibel | Implementasi baseline minimal, pesan error jelas, uji beberapa provider nyata. |
| RAM WebView lebih tinggi dari ekspektasi | Bertentangan dengan positioning ringan | Ukur seluruh proses, kurangi komponen dan renderer yang tidak perlu. |
| API key terekspos | Kebocoran akun/API | Penyimpanan secret OS dan redaksi log. |
| Jawaban AI terlalu panjang | UI berat atau penggunaan konteks mahal | Batas output configurable dan riwayat yang dipilih dengan sadar. |

## 10. Batas dan non-goals MVP

Tidak dibuat dahulu:

- Voice input/output.
- Vision, screenshot capture, OCR.
- Eksekusi command/terminal atau automation.
- Plugin/tool marketplace.
- Sinkronisasi cloud dan akun pengguna.
- Memory AI jangka panjang di luar riwayat percakapan.
- Avatar 3D atau engine game.
- Pembaruan otomatis kompleks.
- Mode offline atau model lokal.

Scope dapat berubah, tetapi setiap penambahan harus menjawab: **apakah ini dibutuhkan untuk menguji core loop chat companion?**
