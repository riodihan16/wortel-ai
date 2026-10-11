# 05 — Integrasi OpenAI-Compatible Provider

## 1. Tujuan

Wortel AI mendukung provider yang menyediakan API chat bergaya OpenAI. MVP fokus pada `POST /chat/completions` (atau path kompatibel yang telah dikonfigurasi) dan streaming SSE.

**Catatan kompatibilitas:** istilah “OpenAI-compatible” tidak selalu berarti kompatibel 100%. Provider mungkin berbeda dalam format base URL, header tambahan, parameter yang didukung, format error, nama model, dan perilaku streaming. Implementasikan baseline yang kecil dan tampilkan error nyata, bukan mencoba menebak semua variasi.

## 2. Konfigurasi provider

Data nonrahasia yang bisa disimpan:

```json
{
  "id": "uuid",
  "name": "Provider pribadi",
  "baseURL": "https://api.example.com/v1",
  "model": "example-model",
  "credentialRef": "wortel-provider-uuid",
  "enabled": true,
  "createdAt": "2026-01-01T00:00:00Z",
  "updatedAt": "2026-01-01T00:00:00Z"
}
```

Contoh ini hanya format ilustratif. Jangan menyimpan field `apiKey` di object yang disimpan ke SQLite atau dikembalikan ke frontend. `credentialRef` adalah nama/ID untuk mencari secret di OS credential store.

MVP minimum:
- Display name.
- Base URL.
- Daftar Model ID (minimal satu; satu ditandai aktif). Tiap ID dapat punya alias tampilan opsional; alias hanya mengubah label di UI, ID tetap dikirim ke endpoint.
- API key.
- Aktif/nonaktif dan provider aktif.

P1 bila benar-benar diperlukan:
- Custom path.
- Custom headers terbatas.
- Parameter temperature/max output yang didukung.
- Toggle streaming.

Jangan menyediakan form “custom headers” bebas tanpa membatasi nama header dan mengecualikan header sensitif seperti `Host`, `Content-Length`, serta mekanisme yang dapat mengubah tujuan koneksi tanpa kontrol.

## 3. Normalisasi URL

Aturan yang disarankan:

1. Parse URL dengan library standar; jangan manipulasi string secara naif.
2. Hanya izinkan skema `https` dan `http`. `https` adalah default yang disarankan.
3. Untuk HTTP non-localhost, tampilkan peringatan karena API key dan isi pesan dapat dikirim tanpa enkripsi transport.
4. Hindari menambahkan `/v1` secara otomatis jika base URL sudah memuat versi/path lain.
5. Gabungkan path agar `/v1/v1/chat/completions` tidak terbentuk.
6. Jangan mengikuti redirect lintas host dengan header Authorization secara otomatis. Kebijakan redirect harus diuji; bila redirect dibutuhkan, hilangkan credential pada host yang berbeda.
7. URL tidak boleh mengandung username/password.
8. Jangan mendukung skema `file:`, `javascript:`, atau skema non-HTTP lain.

Contoh base URL umum:

- `https://api.example.com/v1` → `https://api.example.com/v1/chat/completions`
- `https://gateway.example.com/openai/v1` → `https://gateway.example.com/openai/v1/chat/completions`

Jika provider memberikan endpoint lengkap yang berbeda, jangan memaksanya masuk aturan default; dokumentasikan opsi custom path atau perbedaan yang diperlukan.

## 4. Bentuk request baseline

Contoh payload konsep:

```json
{
  "model": "example-model",
  "messages": [
    { "role": "user", "content": "Jelaskan goroutine dengan sederhana." }
  ],
  "stream": true
}
```

Pesan user dengan lampiran memakai `content` array (bentuk standar OpenAI): part `text` untuk teks pengguna, `image_url` dengan data URL base64 untuk gambar, `file` (`filename` + `file_data` data URL) untuk PDF, dan berkas teks disisipkan sebagai part `text` berpagar. Hanya 3 pesan user terakhir membawa part; yang lebih lama memakai `[gambar: nama]` / `[berkas: nama]` sehingga gambar tidak ditagih ulang tanpa batas (Kenari menagih gambar di history pada tiap request). Tanpa lampiran, `content` tetap string.

Header dasar:

```http
Content-Type: application/json
Authorization: Bearer <API_KEY>
Accept: text/event-stream
```

Provider tertentu mungkin memakai header atau mekanisme autentikasi lain. Jangan membuat asumsi tersebut tanpa kebutuhan atau konfigurasi yang eksplisit.

### Batas request

- Tetapkan batas ukuran pesan yang masuk akal agar aplikasi tidak mengalokasikan body berukuran tak terbatas.
- Batasi jumlah pesan yang dikirim ke provider berdasarkan strategi context window.
- Jangan memasukkan percakapan lain ke request.
- Jangan menyertakan API key di URL/query string.
- Timeout koneksi dan idle stream harus diatur; jangan gunakan timeout sangat pendek yang mematikan respons panjang.
- Streaming dibatalkan dengan context cancellation saat pengguna menekan Stop.

## 5. Parser SSE

SSE bukan sekadar membaca setiap baris dan menganggapnya token. Implementasi harus:

- Memproses stream secara incremental dan mempertahankan buffer untuk baris/frame yang terbelah di antara network reads.
- Mengenali frame yang dipisahkan baris kosong.
- Mendukung satu atau lebih baris `data:` per event sesuai aturan SSE.
- Mengabaikan komentar/keepalive yang diawali `:`.
- Menangani `data: [DONE]` bila digunakan provider.
- Mem-parsing JSON per event secara aman.
- Mengenali event error provider bila ada.
- Menangani EOF tanpa `[DONE]` dengan status yang tidak ambigu.
- Membatasi panjang frame/buffer dan menangani UTF-8 yang terpotong antar-chunk.
- Menutup body response dan membebaskan resource pada semua jalur selesai/error/cancel.

Jangan mengasumsikan setiap payload selalu memiliki `choices[0].delta.content`. Ada event metadata, tool call, role, finish reason, dan variasi provider. MVP cukup mendukung teks; event nonteks yang belum didukung harus diabaikan secara aman atau diberi error yang jelas, bukan membuat aplikasi crash.

## 6. Kontrak event frontend

Backend bisa meneruskan event seperti:

```ts
type ChatEvent =
  | { type: "started"; requestId: string; conversationId: string; messageId: string }
  | { type: "delta"; requestId: string; conversationId: string; messageId: string; delta: string }
  | { type: "completed"; requestId: string; conversationId: string; messageId: string }
  | { type: "cancelled"; requestId: string; conversationId: string; messageId: string }
  | { type: "failed"; requestId: string; conversationId: string; messageId: string; error: PublicError };
```

Ini kontrak ilustratif. Pilih pola final—event terpisah atau satu event dengan `type`—dan gunakan konsisten.

Frontend tidak boleh menganggap `delta` dari request lama milik request aktif yang baru. Selalu cocokkan `requestId` dan `conversationId`.

## 7. Test Connection

Test connection harus menjawab pertanyaan praktis: “Apakah credential, URL, dan model ini dapat dipakai untuk mengirim request sederhana?”

Rencana implementasi:
1. Validasi URL/model dan dapatkan secret dari credential store.
2. Gunakan request seminimal mungkin yang didukung provider. Bila endpoint model listing tidak dijamin ada, jangan mensyaratkannya.
3. Hindari membuat request yang mahal atau mengonsumsi banyak output.
4. Batasi waktu test.
5. Kembalikan kategori hasil: success, unauthorized, forbidden, not found/model, rate limit, server error, network/timeout, unsupported response.
6. Jangan menampilkan atau mencatat header Authorization.

Jika provider hanya mendukung Chat Completions, test boleh mengirim prompt sangat singkat dengan batas output kecil bila parameter tersebut diketahui didukung. Label UI sebaiknya memberi tahu bahwa tes mungkin menghasilkan biaya kecil tergantung provider.

## 8. Error mapping

| Kondisi | Penanganan yang diharapkan |
|---|---|
| 400 | Tampilkan bahwa request/parameter ditolak; sertakan detail aman dan ringkas. |
| 401 | Minta pengguna mengecek API key dan endpoint. |
| 403 | Jelaskan bahwa akses tidak diizinkan; mungkin masalah permission/quota. |
| 404 | Periksa path endpoint dan model ID. |
| 408 / timeout | Tampilkan timeout dan tombol coba lagi. |
| 429 | Tampilkan rate limit/quota; hormati `Retry-After` bila ada. |
| 5xx | Gangguan provider; izinkan retry manual. |
| 413 | Lampiran terlalu besar untuk provider; minta pengguna mengurangi ukuran/jumlah. |
| 400 yang menyebut image/file/vision/modal/document | Model tidak menerima gambar/berkas; sarankan model lain atau kirim teks saja. |
| TLS/DNS/network | Periksa koneksi, URL, sertifikat, proxy, atau firewall. |
| SSE malformed | Tandai protocol error, simpan teks parsial, jangan crash. |
| Cancel | Tampilkan status dibatalkan, bukan error generik. |

Jangan mengulang otomatis request setelah stream mulai menghasilkan token. Retry dapat membuat biaya ganda atau menghasilkan jawaban duplikat.

## 9. Batas fitur MVP

Dukung:
- Text messages dengan role `system` opsional, `user`, dan `assistant` sesuai implementasi.
- Streaming teks.
- Pilihan model sebagai string yang dimasukkan pengguna.
- Satu request chat aktif per conversation.
- Pembatalan request.
- Multimodal/image input dan upload file (gambar, PDF, teks; ADR-054).

Tidak wajib pada MVP:
- Tool/function calling.
- JSON schema/structured outputs.
- Audio/realtime.
- Provider-specific reasoning controls.

## 10. Pengujian provider

Minimal uji:
- Provider kompatibel dengan OpenAI secara standar.
- Base URL dengan prefix path khusus.
- API key salah.
- Model ID yang tidak ada.
- Rate limit/429 bila bisa diuji aman.
- Respons streaming lambat dan stream yang terputus.
- Pengguna menekan Stop.
- Provider mengirim error dalam body non-2xx.
- Pengguna mengganti/menghapus provider saat tidak ada request aktif.

Jangan menaruh credential nyata di test fixture, screenshot, issue tracker, atau CI log.
