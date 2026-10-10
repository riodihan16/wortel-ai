# 06 — UI/UX Desktop Companion

## 1. Arah visual

Karakter 2D kecil dengan tampilan modern, hangat, dan tidak terlalu ramai. UI di sekelilingnya netral, dengan kontras jelas dan aksen warna tunggal.

**Tujuan:** saat idle, Wortel AI terasa seperti desktop pet; ketika diklik, ia berubah menjadi chat tool yang praktis. Karakter bukan pengganti panel chat dan bukan elemen dekorasi yang wajib mengambil perhatian.

## 2. Mode tampilan

### Mode A — Pet

Elemen:
- Karakter 2D.
- Indikator status pip 14 px di sudut kanan bawah karakter, selalu tampak
  (ADR-044). Bentuk membedakan state, jadi warna bukan satu-satunya pembawa
  makna (WCAG 1.4.1): cincin kosong = `idle`, cakram terisi berdenyut =
  `thinking`, cakram + ikon seru = `error`. Teks status tetap tersedia bagi
  screen reader lewat `.sr-only`.
- Tooltip atau hint singkat ketika hover, bukan dialog mengganggu.
- Context menu sederhana: Open chat, Hide, Settings, Exit (menu dapat ditunda sampai tray tersedia).

Interaksi:
- Click: buka chat.
- Drag: pindahkan posisi.
- Klik kanan: menu opsional.
- Escape: tidak boleh membuat aplikasi hilang tanpa cara memulihkan.
- Shortcut global: toggle chat (P1).

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
- Base URL.
- Daftar Model ID: tambah/hapus model; radio memilih model aktif.
- API key input dengan reveal/hide.
- Save.
- Test connection.
- Status aktif.

Header menampilkan judul aplikasi dan tombol kontrol (Settings, Pet, Hide, Quit). Di bawah kotak pesan, satu baris footer menampilkan tombol pemilih model (label `Model: <aktif>`) di kiri dan indikator `status: <state>` di kanan, sehingga model bisa diganti tanpa membuka Settings (ADR-029, ADR-034).

Jangan menampilkan API key lama penuh setelah settings dibuka. Tampilkan “credential tersimpan” dan sediakan aksi replace/remove.

## 3. State dan status karakter

| State | Trigger | Visual yang disarankan |
|---|---|---|
| `idle` | Tidak ada interaksi/request | Gerakan kecil berkala atau pose statis; pip cincin kosong. |
| `hover` | Pointer berada di atas karakter | Reaksi singkat, bukan animasi panjang. |
| `thinking` | Request diterima dan menunggu token | Pose/indikator ringan; mulai hanya saat request; pip cakram aksen berdenyut. |
| `speaking` | Delta teks masuk | Animasi sangat sederhana atau indikator pulse terbatas. |
| `error` | Request gagal | Ekspresi/indikator error; jangan hanya mengandalkan warna — pip error memakai cakram + ikon seru. |
| `offline` | Provider belum dikonfigurasi atau tes koneksi gagal | Hint saat dibuka, bukan popup berulang. |

Jangan menjanjikan bahwa state karakter mencerminkan “emosi AI” secara nyata. Ini hanya status UI.

Pemetaan pip di kode: `PetSurface.petStatus()` memetakan `sending`/`streaming` →
`thinking`, `error` → `error`, selain itu → `idle`. `speaking` belum dibedakan
dari `thinking` karena delta token dan fase tunggu memakai status chat yang sama.
Saat `prefers-reduced-motion: reduce`, animasi denyut berhenti di frame akhirnya
(`animation-iteration-count: 1` di `base.css`), sehingga cakram tetap terlihat
tanpa gerak.

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

Implementasi: `ChatPanel` bercabang pada prop `hasProvider` (dihitung dari
`provs.activeProvider`). Tanpa provider aktif, empty state menampilkan heading
“Belum ada provider”, langkah konkret, dan tombol `Buka Pengaturan`; contoh
prompt tidak ditampilkan karena belum ada yang bisa menjawab. Contoh prompt
hanya muncul saat provider sudah aktif, sehingga tidak menjanjikan hal yang
belum bisa dilakukan.

### Loading

Tampilkan state yang tenang. Hindari spinner besar atau progress persentase palsu untuk request yang durasinya tidak diketahui.

### Error

Sertakan:
- Kalimat singkat yang menjelaskan masalah.
- Kategori atau HTTP status jika aman.
- Langkah berikutnya yang spesifik.
- Retry manual bila masuk akal.

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
(hover) · `--color-well` (input/kolom dalam). Bukan hitam murni — hue netral dingin
(280) pada lightness 17%. Tema terang memakai ramp terang dengan hue sama.

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
accent). Dark memakai hue yang sama dengan `--color-brand` (`#f68c36`); light
memakai wortel gelap (`#ac4403`) agar teks accent tetap ≥ 4.5:1.

**Semantik:** `--color-danger`/`--color-danger-soft`, `--color-success`/`--color-success-soft`.
**Brand:** `--color-brand` (oranye wortel) + `--color-brand-leaf` — hanya untuk
karakter, tidak pernah untuk chrome UI.

**Shadow:** tiga tingkat (`--shadow-1/2/3`), blur tunggal per tingkat, tanpa
tumpukan berlapis.

**Motion:** `--dur-fast` 110 ms · `--dur` 170 ms · `--dur-slow` 260 ms, dengan
`--ease-out` / `--ease-in-out`; `prefers-reduced-motion` mematikan semuanya.

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
