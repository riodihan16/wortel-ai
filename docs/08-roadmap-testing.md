# 08 — Roadmap Implementasi, Testing, dan Release Checklist

## 1. Prinsip delivery

Bangun satu jalur vertikal yang berjalan dari jendela sampai provider sebelum memperluas fitur. Jangan membuat semua struktur folder dan service abstrak di awal jika belum dipakai.

## 2. Milestone

### M0 — Validasi risiko teknis

**Tujuan:** menguji Wails di Windows sebelum banyak kode dibuat.

Tugas:
- Buat project Wails minimal dengan versi yang dikunci.
- Build dan jalankan di Windows.
- Uji window frameless dan background transparan.
- Uji resize dari mode kecil ke panel chat.
- Uji drag, posisi, always-on-top, dan fokus.
- Ukur baseline RAM/CPU seluruh process tree setelah startup dan idle.
- Pastikan perilaku WebView2 pada mesin target.

Definition of done:
- Ada build runnable.
- Dicatat perilaku yang berhasil/gagal.
- Keputusan satu jendela tetap layak atau perlu alternatif ditulis di `09-decisions.md`.

### M1 — Companion statis

Tugas:
- Asset karakter lokal.
- Mode pet dan mode panel dasar.
- Drag dan posisi aman.
- Toggle pet/chat.
- Tidak ada API/network/storage yang belum diperlukan.

Definition of done:
- Karakter tampil konsisten.
- Tidak terpotong di tepi layar.
- Animasi idle tidak menyebabkan CPU tinggi.
- Bisa ditutup secara normal.

### M2 — Chat tanpa persistence kompleks

Tugas:
- Satu provider/config development yang tidak menyimpan secret ke repo.
- Request non-streaming awal untuk membuktikan koneksi.
- Tambahkan SSE streaming dan event contract.
- Tambahkan Stop/cancel.
- Error mapping dasar.

Definition of done:
- Request berhasil ke endpoint uji.
- Streaming dan cancellation lulus.
- Tidak ada API key di UI event atau log.
- Tidak ada request duplikat ketika tombol diklik cepat.

### M3 — Provider settings aman

Tugas:
- CRUD provider profile.
- Credential store OS.
- Validasi base URL/model.
- Test connection.
- Pilih provider aktif.

Definition of done:
- Credential tidak disimpan di SQLite atau settings JSON.
- Mengganti provider tidak menghilangkan riwayat.
- Kesalahan konfigurasi dapat dipahami pengguna.

### M4 — SQLite conversation history

Tugas:
- Schema dan migration.
- Repository percakapan dan pesan.
- List/open/rename/delete conversation.
- Penyimpanan state complete/cancelled/failed.

Definition of done:
- Percakapan bertahan sesudah restart.
- Request yang dibatalkan atau gagal tidak dianggap sukses.
- Migration dan kegagalan DB diuji.

### M5 — Kualitas desktop

Tugas:
- Shortcut global.
- System tray.
- Startup opsional.
- Settings appearance.
- Installer atau portable bundle.

Definition of done:
- Shortcut tidak bentrok secara diam-diam.
- Exit menutup semua resource.
- Startup behavior jelas dan opt-in.
- Installer/portable build diuji pada Windows target yang bersih.

### M6 — RC dan dokumentasi rilis

Tugas:
- Regression test.
- Pengukuran RAM/CPU.
- Uji konfigurasi dan error provider.
- Review secret handling dan Markdown security.
- Tulis cara install/uninstall, lokasi database, dan cara menghapus data.

Definition of done:
- Seluruh P0 dari requirement lulus.
- Known issues dicatat.
- Versi dan hash artifact rilis dicatat bila sesuai.

## 3. Strategi testing

### Unit test backend

- URL normalization.
- Provider config validation.
- SSE parser (frame terfragmentasi, banyak baris data, komentar, `[DONE]`, EOF, malformed JSON).
- Error mapping HTTP.
- Conversation/message repository.
- Credential reference lifecycle (mock credential store untuk test).
- Cancellation dan request ID mapping.

### Integration test

Gunakan HTTP test server lokal atau mock provider untuk:
- Non-stream response.
- SSE dengan beberapa delta.
- Stream lambat yang dibatalkan.
- 401, 403, 404, 429, 500.
- Response JSON malformed.
- Server disconnect di tengah stream.
- Timeout.

Jangan menjadikan test suite wajib bergantung ke API berbayar/internet. Smoke test provider nyata dijalankan secara manual dengan credential lokal yang tidak dicatat.

### Frontend test

- Composer mengirim/menolak input kosong.
- Enter vs Shift+Enter.
- State Send/Stop.
- Delta dengan request ID yang salah diabaikan.
- Error notice dan retry state.
- Provider key tidak muncul dalam state visible/persisted.
- Markdown link scheme dibatasi.

### Manual test Windows

- Windows 10/11 yang didukung.
- WebView2 tersedia/tidak tersedia jika memungkinkan diuji.
- Scaling 100%, 125%, 150%.
- Satu monitor dan multi-monitor bila tersedia.
- Resolusi rendah dan tinggi.
- Window berpindah ke berbagai tepi layar.
- Sleep/resume atau display change bila relevan.
- Aplikasi keluar ketika request aktif.
- Restart komputer untuk menguji startup hanya setelah fitur dibuat.

## 4. Pengujian performa

Ukur pada build release, bukan hanya dev mode. Catat OS, hardware, resolusi, scaling, dan versi WebView2 bila dapat diperoleh.

Skenario:
1. Jalankan aplikasi tanpa provider.
2. Diamkan 60 detik pada mode pet.
3. Buka/tutup panel chat beberapa kali.
4. Buka riwayat panjang.
5. Streaming jawaban panjang.
6. Tutup aplikasi ketika request berjalan.

Metrik:
- RAM process utama + proses WebView2 yang berkaitan.
- CPU average dan peak selama idle.
- Startup time.
- Responsivitas window saat streaming.
- Pertumbuhan RAM setelah membuka/tutup beberapa conversation.

Target awal yang disarankan (bersifat **target eksperimen**, bukan jaminan):
- CPU idle rata-rata mendekati 0–1% setelah startup dan aktivitas berhenti.
- Tidak ada pertumbuhan RAM terus-menerus selama idle.
- Startup pada SSD dalam beberapa detik di mesin referensi.
- Tetapkan budget RAM final berdasarkan pengukuran baseline M0; jangan menyembunyikan konsumsi proses WebView2.

## 5. Release checklist

### Fungsional
- [ ] Build release berhasil dari clean checkout.
- [ ] Companion bisa tampil, drag, hide/show.
- [ ] Panel chat bisa dibuka dan ditutup.
- [ ] Pesan terkirim ke provider yang dipilih.
- [ ] Streaming berfungsi dan dapat dibatalkan.
- [ ] Error provider tidak membuat aplikasi crash.
- [ ] Percakapan bertahan setelah restart.
- [ ] Pengaturan provider dapat diubah dan dites.

### Keamanan
- [ ] Tidak ada API key di repository, log, SQLite, dan event frontend.
- [ ] Markdown raw HTML dimatikan/disanitasi.
- [ ] Link eksternal menggunakan allowlist scheme.
- [ ] Redirect cross-host tidak membawa Authorization.
- [ ] Tidak ada command execution atau screenshot capture.
- [ ] HTTP endpoint non-TLS mendapat warning yang sesuai.

### Performa dan stability
- [ ] Idle CPU/RAM diukur pada release build.
- [ ] Tidak ada goroutine/request yang tertinggal setelah cancel/exit.
- [ ] UI tetap merespons saat streaming.
- [ ] Window tidak hilang di luar area layar setelah monitor berubah.
- [ ] Database failure memberikan pesan yang jelas.

### Distribusi
- [ ] Versi aplikasi ditampilkan.
- [ ] Cara install/run dan uninstall didokumentasikan.
- [ ] Lokasi data lokal dijelaskan.
- [ ] WebView2 requirement dijelaskan.
- [ ] Signing/SmartScreen behavior dipertimbangkan sebelum distribusi publik.
