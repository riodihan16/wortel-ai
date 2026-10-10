# 06 — UI/UX Desktop Companion

## 1. Arah visual

Karakter 2D kecil dengan tampilan modern, hangat, dan tidak terlalu ramai. UI di sekelilingnya netral, dengan kontras jelas dan aksen warna tunggal.

**Tujuan:** saat idle, Wortel AI terasa seperti desktop pet; ketika diklik, ia berubah menjadi chat tool yang praktis. Karakter bukan pengganti panel chat dan bukan elemen dekorasi yang wajib mengambil perhatian.

## 2. Mode tampilan

### Mode A — Pet

Elemen:
- Karakter 2D.
- Status dibawa oleh wortel itu sendiri, tanpa pip terpisah (ADR-049,
  menggantikan pip ADR-044). Bentuk membedakan state, jadi warna bukan
  satu-satunya pembawa makna (WCAG 1.4.1): `idle` = senyum; `thinking` = mata
  melirik ke atas + gelembung pikiran "…"; `error` = alis cemas, mulut
  cemberut, daun layu, dan badge "!". Teks status tetap tersedia bagi screen
  reader lewat `.sr-only`.
- Warna karakter tetap (`--carrot-*`) dan tidak ikut tema Gelap/Terang: pet
  duduk di atas wallpaper, bukan di atas kertas aplikasi. Outline cokelat gelap
  menjaga siluet terbaca di wallpaper terang maupun gelap.
- Tooltip atau hint singkat ketika hover, bukan dialog mengganggu.
- Context menu sederhana: Open chat, Hide, Settings, Exit (menu dapat ditunda sampai tray tersedia).

Interaksi:
- Click: buka chat.
- Drag: pindahkan posisi.
- Klik kanan: menu opsional.
- Escape: tidak boleh membuat aplikasi hilang tanpa cara memulihkan.
- Shortcut global (P1, ADR-050): jendela tersembunyi atau pet → buka chat; panel
  terbuka → kembali ke pet. Shortcut tidak pernah menyembunyikan companion.
  Hint hover pet menampilkan kombinasinya ("Klik atau Ctrl+Shift+W: chat").

Batas desain:
- Tidak ada teks permanen yang menutupi desktop. Pip 14 px bukan teks dan
  tidak pernah menutupi konten pengguna.
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
- About/version (implementasi: section `Tentang` menampilkan versi build,
  lisensi MIT, atribusi font Geist/Geist Mono, dan pernyataan lokasi data).

Provider form minimum:
- Display name.
- Base URL (dengan hint: endpoint kompatibel OpenAI, biasanya berakhiran `/v1`).
- Daftar Model ID: tambah/hapus model; radio memilih model aktif.
- API key input dengan reveal/hide, plus keterangan bahwa key disimpan di
  Windows Credential Manager dan tidak ditampilkan lagi.
- `Tes koneksi` (sebelum) `Simpan`.
- Status aktif (badge `Aktif` mengikuti `ActiveProvider()` dari backend).

Implementasi (ADR-049): form provider baru tersembunyi di balik
`+ Tambah provider` begitu ada satu provider; pada run pertama form langsung
terbuka. Form provider baru menawarkan preset (OpenAI, OpenRouter, Groq,
LM Studio, Ollama) yang hanya mengisi nama dan Base URL; nama model tetap
contoh di placeholder. `Ubah` membuka form tepat di bawah kartunya dan memfokus
`Nama`. `Hapus` provider meminta konfirmasi inline (pola ADR-048) dan tidak
ditawarkan untuk provider aktif, karena backend menolaknya. Simpan pertama
menampilkan "tersimpan dan aktif" dengan tombol `Mulai chat`.

Header menampilkan judul aplikasi dan tombol kontrol (Settings, Kecilkan jadi wortel, Hide, Quit); versi aplikasi ada di Settings › Tentang. Di bawah kotak pesan, satu baris footer menampilkan tombol pemilih model (nama aksesibel `Model <aktif>, provider <nama>`) di kiri, dan di kanan status hanya selama ada pekerjaan (`Mengirim…`, `Wortel mengetik…`), sehingga model bisa diganti tanpa membuka Settings (ADR-029, ADR-034, ADR-049).

Jangan menampilkan API key lama penuh setelah settings dibuka. Tampilkan “credential tersimpan” dan sediakan aksi replace/remove.

## 3. State dan status karakter

| State | Trigger | Visual yang disarankan |
|---|---|---|
| `idle` | Tidak ada interaksi/request | Pose statis, senyum; berkedip ~120 ms tiap 6 detik (hanya di pet). |
| `hover` | Pointer berada di atas karakter | Reaksi singkat, bukan animasi panjang. |
| `thinking` | Request diterima dan menunggu token | Lompatan kecil saat mulai, mata ke atas, gelembung "…" yang titiknya berdenyut bergantian. |
| `speaking` | Delta teks masuk | Animasi sangat sederhana atau indikator pulse terbatas. |
| `error` | Request gagal | Gelengan singkat, alis cemas, mulut cemberut, daun layu, badge "!"; bukan warna saja. |
| `offline` | Provider belum dikonfigurasi atau tes koneksi gagal | Hint saat dibuka, bukan popup berulang. |
| `reply-ready` | Balasan selesai saat user di mode pet | Gelembung "Sudah!" sampai chat dibuka (ADR-050). Error yang sudah dilihat di chat tidak lagi membuat pet cemas. |

Jangan menjanjikan bahwa state karakter mencerminkan “emosi AI” secara nyata. Ini hanya status UI.

Pemetaan di kode: `moodFor()` di `PetSurface` memetakan `sending`/`streaming` →
`thinking`, `error` → `error`, selain itu → `idle`; komponen `Carrot` menggambar
wajah per mood dan dipakai juga di empty state chat serta sebagai avatar setiap
balasan (wajahnya mengikuti status balasan itu). `speaking` belum dibedakan dari
`thinking` karena delta token dan fase tunggu memakai status chat yang sama.
Saat `prefers-reduced-motion: reduce`, lompatan, gelengan, pop, dan kedip
berhenti (`base.css`); denyut titik "…" tetap jalan karena hanya opacity dan
satu-satunya tanda bahwa balasan masih datang.

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
- Loading indicator tidak boleh menutupi teks yang sudah diterima. Saat balasan
  berjalan, footer chat menampilkan wortel kecil yang berpikir.
- Transkrip hanya mengikuti stream selama user berada di dekat bawah; bila user
  sedang membaca ke atas, chip "Balasan baru" muncul (ADR-050).
- Saat error, pertahankan teks pengguna dan teks parsial agar tidak hilang.
- `Keluar` (header chat dan rail pet) meminta konfirmasi inline, Batal
  difokus lebih dulu (ADR-050).
- Tombol destructive seperti Delete conversation dan Hapus provider diberi
  konfirmasi inline bertema (ADR-048, primitive `.confirm-row`). Jangan pakai `window.confirm`/`alert`/`prompt` maupun
  `<select>` native: WebView2 menggambar keduanya dengan chrome sistem di luar
  tema.

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
- Code block memiliki tombol Copy; setiap balasan selesai punya `Salin` (sumber
  markdown). Bila clipboard ditolak, tombol menulis "Gagal menyalin".
- Teks panjang tidak boleh memperlebar panel; gunakan wrapping atau horizontal scroll di code block.
- Token streaming harus tetap terbaca meskipun Markdown sementara belum lengkap (misalnya fenced code belum ditutup).

## 8. Empty, loading, dan error states

### Belum ada percakapan

Tampilkan sapaan pendek dan satu contoh prompt, tetapi jangan mengirimnya otomatis.

Implementasi (ADR-049): wortel + "Mau tanya apa?" + satu contoh yang bila
diklik mengisi composer (tidak terkirim sampai pengguna menekan Enter).

### Belum ada provider

Tampilkan instruksi untuk menambahkan endpoint/model/API key. Jangan tampilkan chat seolah-olah sudah siap.

Implementasi: `ChatPanel` bercabang pada prop `hasProvider` (dihitung dari
`provs.activeProvider`). Tanpa provider aktif, empty state menampilkan wortel,
heading “Wortel belum tersambung”, satu kalimat langkah, dan tombol
`Hubungkan provider`; composer dan tombol kirim dinonaktifkan dengan placeholder
"Hubungkan provider dulu untuk mulai chat" (ADR-049). Contoh prompt hanya
muncul saat provider sudah aktif, sehingga tidak menjanjikan hal yang belum
bisa dilakukan.

### Loading

Tampilkan state yang tenang. Hindari spinner besar atau progress persentase palsu untuk request yang durasinya tidak diketahui.

### Error

Sertakan:
- Kalimat singkat yang menjelaskan masalah.
- Kategori atau HTTP status jika aman.
- Langkah berikutnya yang spesifik.
- Retry manual bila masuk akal.

Implementasi (ADR-049): balasan gagal tampil sebagai kartu error di tempat
balasan itu (alasan + `Coba lagi` hanya pada giliran terbaru). `Coba lagi`
memanggil `RetryMessage`: balasan gagal dibuang dan pesan user yang sudah
tersimpan dipakai ulang, jadi pertanyaan tidak muncul dua kali di layar maupun
di riwayat. Banner `#chat-error` hanya untuk error yang bukan milik balasan
(mis. gagal ganti judul) dan bisa ditutup. Balasan yang dihentikan diberi
penanda "— Dihentikan" di dalam balasannya; tombol Stop adalah tombol tinta
terisi 32 px dan Escape di composer juga menghentikan.

## 9. Design tokens final

Nilai di bawah adalah implementasi nyata di `frontend/src/styles/tokens.css`
(ADR-043). Semua warna OKLCH; komponen **wajib** memakai nama token, bukan nilai
mentah. Skala mengikuti kelipatan 4 px.

**Radius:** `--radius-xs` 6 · `--radius-sm` 8 · `--radius-md` 12 · `--radius-lg` 16 ·
`--radius-xl` 22 · `--radius-pill` 999 (panel 16, bubble 12 — sesuai rentang draft).

**Spacing:** `--space-1` 4 · `--space-2` 8 · `--space-3` 12 · `--space-4` 16 ·
`--space-5` 20 · `--space-6` 24 · `--space-7` 32.

**Tipografi:** dua famili variabel yang dibundel lokal (tanpa CDN, offline-safe):
`--font-sans` = Geist (100–900), `--font-mono` = Geist Mono. Ukuran memakai
`calc(<px> * var(--font-scale))` sehingga setelan ukuran teks menskalakan seluruh
UI termasuk markdown. Draft sebelumnya menyarankan system UI stack untuk
menghindari bundling font; keputusan berubah demi konsistensi lintas mesin —
biaya file woff2 variabel (latin + latin-ext) dinilai dapat diterima.

**Permukaan (dark = default, `:root[data-theme="light"]` override):** ramp
`--color-paper` (permukaan utama) · `--color-paper-2` (kartu/isi) · `--color-paper-3`
(hover) · `--color-well` (input/kolom dalam). Bukan hitam murni — netral hangat "Tanah"
(hue 70, chroma rendah) pada lightness 17%; sebelum ADR-049 hue dingin 280.
Tema terang memakai ramp terang dengan hue sama.

**Tinta:** `--color-ink` (utama) · `--color-ink-2` (sekunder) · `--color-ink-3`
(metadata). Kontras pada `--color-paper` berturut-turut **16.5:1, 8.0:1, 6.2:1**
(dark) dan 17.0:1, 6.9:1, 5.3:1 (light) — lolos WCAG AA. Rasio **terburuk** ada
di permukaan `--color-paper-3`, bukan `--color-paper`: 12.7 / 6.1 / 4.5 (dark) dan
14.4 / 5.8 / 4.5 (light). `--color-ink-3` karena itu hanya boleh dipakai pada
permukaan tempat token itu masih ≥ 4.5:1.

**Border:** `--color-rule` (garis pemisah halus) · `--color-rule-strong` (batas
kontrol, terukur 4.1:1 pada `--color-paper` dark dan 3.6:1 light — memenuhi 3:1
untuk komponen UI, WCAG 1.4.11).

**Accent:** oranye wortel tunggal (ADR-047) — `--color-accent`, dengan
`--color-accent-soft` (latar pill/banner) dan `--color-accent-ink` (teks di atas
accent). Dark memakai hue yang sama dengan `--carrot-body` (`#f68c36`); light
memakai wortel gelap (`#ac4403`) agar teks accent tetap ≥ 4.5:1.

**Semantik:** `--color-danger`/`--color-danger-soft`, `--color-success`/`--color-success-soft`.
**Karakter:** `--carrot-body`, `--carrot-shade`, `--carrot-leaf`, `--carrot-ink`
(outline), `--carrot-cheek`, `--carrot-shine`, `--carrot-alert` — tetap, tidak
di-override tema, hanya untuk karakter dan `BrandMark` (ADR-049; menggantikan
`--color-brand*`).

**Shadow:** tiga tingkat (`--shadow-1/2/3`), blur tunggal per tingkat, tanpa
tumpukan berlapis.

**Motion:** `--dur-fast` 110 ms · `--dur` 170 ms · `--dur-slow` 260 ms, dengan
`--ease-out` / `--ease-in-out`; `prefers-reduced-motion` mematikan semuanya
kecuali denyut opacity titik "…" (lihat §3).

**Fokus:** satu perlakuan global `--focus-ring` (2 px accent, offset 2 px) pada
`:focus-visible`.

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
