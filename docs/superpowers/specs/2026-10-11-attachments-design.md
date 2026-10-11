# Spec — Lampiran gambar dan berkas pada chat

- **Tanggal:** 2026-10-11
- **Status:** Disetujui (brainstorm 2026-10-11), menunggu plan implementasi
- **Provider acuan:** Kenari (`https://kenari.id/v1`, OpenAI Chat Completions). Format yang dipakai tetap standar OpenAI-compatible; tidak ada parameter khusus Kenari.

## 1. Tujuan

Pengguna dapat melampirkan gambar, PDF, dan berkas teks pada pesan chat sehingga model multimodal provider dapat membacanya. Lampiran tersimpan di riwayat lokal, tampil kembali saat percakapan dibuka, dan ikut dikirim ulang ke model dalam batas yang terkendali biaya.

Di luar scope: downscale gambar, engine OCR, parameter `detail`, URL gambar remote, format Anthropic Messages / OpenAI Responses, drag-drop ke mode pet, ekspor lampiran, flag vision per model, pembacaan clipboard/screenshot otomatis.

## 2. Keputusan yang sudah diambil

| Topik | Keputusan |
|---|---|
| Tipe berkas | Gambar `image/png`, `image/jpeg`, `image/webp`; PDF `application/pdf`; teks `.txt .md .csv .json` |
| Cara masuk | Tombol klip (file dialog), Ctrl+V gambar di composer, drag-drop ke panel chat |
| Riwayat ke model | Lampiran penuh hanya untuk 3 turn user terakhir; lebih lama jadi placeholder teks |
| Deteksi kemampuan model | Tanpa gating; penolakan provider dipetakan ke pesan yang jelas |
| Staging sebelum kirim | Di Go (memori), frontend hanya memegang metadata |
| Persistensi | BLOB di SQLite, tabel `attachments`, cascade dari `messages` |
| Tampil gambar | Route asset-server `/attachments/{id}`, bukan base64 lewat IPC |

## 3. Alur data

1. Pengguna melampirkan lewat salah satu jalur:
   - Tombol klip → `PickAttachments()` → `runtime.OpenMultipleFilesDialog` → Go membaca path.
   - Drop → runtime `OnFileDrop` memberi path → `StageAttachmentPaths(paths)`.
   - Ctrl+V → event `paste` pada textarea, `clipboardData.files` berisi gambar → FileReader → `StageAttachmentData(name, mime, base64)`.
2. Go memvalidasi (bagian 7) dan menyimpan bytes di **staging** memori: `map[id]item`, id UUID v4. Staging adalah isi pesan yang sedang disusun, jadi batas per pesan (≤ 4 item, ≤ 20 MB total) ditegakkan saat menambah; `SendMessage` tidak pernah ditolak karena jumlah/ukuran lampiran.
3. Frontend menerima `AttachmentMeta{id, name, mime, size, kind}` dan menampilkan chip/thumbnail. Thumbnail memakai `<img src="/attachments/{id}">`.
4. `SendMessage(conversationId, text, attachmentIds)`:
   - `text` boleh kosong bila `attachmentIds` tidak kosong. Keduanya kosong → error "empty message".
   - Semua id harus ada di staging; id tidak dikenal → error, tidak ada yang dikirim.
   - Satu transaksi: `EnsureConversation`, insert row user, insert rows `attachments` (BLOB, `position` urut), insert row assistant `streaming`. Setelah commit, item dihapus dari staging.
   - Request dibangun dari DB (bagian 5), lalu stream seperti sekarang.
5. `RetryMessage` tidak berubah secara kontrak: `PrepareRetry` menghapus reply gagal dan memakai row user yang ada; lampiran menempel pada row itu dan ikut dibangun ulang dari DB. Bila `PrepareRetry` mengembalikan `false` (row user tidak ditemukan, yaitu `SendMessage` ditolak sebelum menyentuh DB), retry berjalan sebagai kirim teks biasa tanpa lampiran, sama dengan perilaku sekarang. Kasus ini hanya terjadi pada penolakan sinkron (`generation already active`, id staging tidak dikenal, storage tidak ada), bukan pada kegagalan provider; dicatat sebagai batas yang diketahui.
6. `History` mengembalikan `ChatMessage.attachments: AttachmentMeta[]` (tanpa bytes) sehingga percakapan lama menampilkan thumbnail/chip lewat route yang sama.
7. Hapus percakapan → cascade `messages` → cascade `attachments`. Tidak ada file di disk, tidak ada sweep.
8. Tanpa DB (fallback mock): semua binding lampiran mengembalikan error "Lampiran butuh storage, restart aplikasi".

## 4. Kontrak backend

### Tipe (`internal/app/attachments.go`)

```go
type AttachmentMeta struct {
    ID   string `json:"id"`
    Name string `json:"name"`
    Mime string `json:"mime"`
    Size int64  `json:"size"`
    Kind string `json:"kind"` // image | pdf | text
}
```

`ChatMessage` (history) mendapat `Attachments []AttachmentMeta` dengan tag `json:"attachments"` (selalu array, tidak pernah null).

### Binding (`app.go`)

| Binding | Perilaku |
|---|---|
| `PickAttachments() ([]AttachmentMeta, error)` | Dialog multi-file dengan filter "Gambar, PDF, teks"; batal → `[]`, bukan error |
| `StageAttachmentPaths(paths []string) ([]AttachmentMeta, error)` | Dari drop; path yang gagal validasi menggagalkan seluruh batch dengan pesan menyebut nama berkas (bukan path penuh) |
| `StageAttachmentData(name, mime string, data string) (AttachmentMeta, error)` | Dari paste; `data` base64 standar tanpa prefix `data:` |
| `UnstageAttachment(id string)` | Idempoten |
| `SendMessage(conversationId, text string, attachmentIds []string) (string, error)` | Signature berubah; semua pemanggil lama diperbarui |
| `RetryMessage(conversationId, text string)` | Tetap |
| `History(conversationId, n)` | Item membawa `attachments` |

Semua binding yang menyentuh staging/DB dibungkus `Service`; `app.go` tetap adapter tipis.

### Route asset

`main.go` menambahkan `assetserver.Options{Assets, Handler: attachmentHandler}`. Wails memanggil `Handler` hanya untuk path yang tidak ada di embed.

- Path `GET /attachments/{id}`; `{id}` wajib cocok `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`. Lainnya → 404.
- Urutan cari: staging, lalu DB (`SELECT mime, data FROM attachments WHERE id=?`). Tidak ada → 404.
- Header: `Content-Type` dari mime tersimpan, `X-Content-Type-Options: nosniff`, `Cache-Control: private, max-age=3600`, `Content-Disposition: inline`.
- Hanya metode GET; metode lain → 405.

Drag-drop: `options.App.DragAndDrop = &options.DragAndDrop{EnableFileDrop: true}`; elemen target memakai `--wails-drop-target: drop`.

## 5. Request ke provider

### Tipe (`internal/ai`)

```go
type Part struct {
    Type     string    `json:"type"`                // text | image_url | file
    Text     string    `json:"text,omitempty"`
    ImageURL *ImageURL `json:"image_url,omitempty"`
    File     *FilePart `json:"file,omitempty"`
}
type ImageURL struct{ URL string `json:"url"` }
type FilePart struct {
    Filename string `json:"filename"`
    FileData string `json:"file_data"`
}
type Message struct {
    Role    string
    Content string
    Parts   []Part
}
```

`Message` punya `MarshalJSON` kustom: `Parts` kosong → `{"role","content":"<string>"}` (kompatibel mock, tes lama, dan provider yang tidak memahami array). `Parts` terisi → `content` array; `Content` yang tidak kosong menjadi part `text` pertama.

### Pemetaan lampiran ke part

| Kind | Part |
|---|---|
| image | `{"type":"image_url","image_url":{"url":"data:<mime>;base64,<data>"}}` |
| pdf | `{"type":"file","file":{"filename":"<name>","file_data":"data:application/pdf;base64,<data>"}}` |
| text | `{"type":"text","text":"Isi berkas `<name>`:\n```\n<isi>\n```"}` |

Urutan part dalam satu pesan: teks pengguna (bila ada), lalu lampiran sesuai `position`.

### Kebijakan riwayat

- Konteks tetap 20 pesan terakhir (`RecentMessages`).
- Konstanta `AttachmentTurns = 3`: tiga pesan user terakhir (dihitung dari yang terbaru) membawa part lampiran penuh.
- Pesan user yang lebih lama: tiap lampiran menjadi satu baris teks placeholder yang ditambahkan setelah teks pengguna: `[gambar: nama.png]`, `[berkas: nama.pdf]`, `[berkas: catatan.txt]`. Pesan itu tetap dikirim sebagai `content` string.
- Bytes dibaca dari DB hanya untuk pesan yang masuk jendela 3 turn; query metadata tidak memuat kolom `data`.

### Pemetaan error tambahan (`internal/ai/errors.go`)

- `413` → `{code:"payload_too_large", message:"Lampiran terlalu besar untuk provider. Kurangi ukuran atau jumlah berkas.", retryable:false}`.
- `400`: body dibaca maksimal 4 KB (tidak di-log, tidak di-event). Bila body (lowercase) memuat salah satu dari `image`, `file`, `vision`, `modal`, `multimodal`, `document` → `{code:"unsupported_input", message:"Model ini tidak menerima gambar/berkas. Pilih model lain atau kirim teks saja."}`. Selain itu tetap `bad_request`.

## 6. Storage

Migrasi **v5** (`currentVersion = 5`):

```sql
CREATE TABLE attachments (
  id TEXT PRIMARY KEY,
  message_id TEXT NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
  kind TEXT NOT NULL,
  mime TEXT NOT NULL,
  name TEXT NOT NULL,
  size INTEGER NOT NULL,
  data BLOB NOT NULL,
  position INTEGER NOT NULL,
  created_at TEXT NOT NULL
);
CREATE INDEX idx_attachments_msg ON attachments(message_id, position);
```

Fungsi storage baru:

- `AppendUserTurn(user MessageRow, atts []AttachmentRow, assistant MessageRow) error` — satu transaksi untuk tiga insert di atas. Jalur kirim (`SendMessage`) selalu memakai fungsi ini, juga saat `atts` kosong. Jalur retry yang memakai ulang row user tetap memakai `AppendMessage` untuk row assistant saja.
- `AttachmentsFor(messageIDs []string) (map[string][]AttachmentRow, error)` — metadata saja.
- `AttachmentData(id string) (mime string, data []byte, err error)`.
- `RecentContext(conversationID string, n int) ([]ContextRow, error)` — pengganti `RecentMessages`: id pesan + role + content + metadata lampiran, agar service dapat menerapkan kebijakan 3 turn. `RecentMessages` dan tesnya dihapus; pemanggil satu-satunya adalah `Service.generate`. Perbaikan ikutan: row assistant dengan `content=''` (row `streaming` yang baru dibuat untuk balasan ini) tidak lagi ikut dikirim ke provider; sebelumnya `RecentMessages` menyertakannya sebagai `{"role":"assistant","content":""}` di akhir request.
- `PrepareRetry` tidak berubah (hanya menghapus row assistant).

`docs/07-storage-security-privacy.md` ditambah tabel `attachments`; `data` adalah BLOB konten pengguna, bukan secret.

## 7. Batas dan validasi (backend, berlaku untuk ketiga jalur masuk)

| Aturan | Nilai |
|---|---|
| Jumlah per pesan | ≤ 4 |
| Gambar | `http.DetectContentType` ∈ {png, jpeg, webp}; ≤ 8 MB |
| PDF | sniff `%PDF-`; ≤ 15 MB |
| Teks | ekstensi `.txt .md .csv .json` → mime `text/plain`, `text/markdown`, `text/csv`, `application/json`; ≤ 256 KB; wajib `utf8.Valid`; byte NUL ditolak |
| Total mentah per pesan | ≤ 20 MB (base64 ≈ 27 MB, di bawah 32 MiB body Kenari) |
| Nama | basename saja, ≤ 120 rune, karakter kontrol dibuang |
| Staging | sama dengan batas per pesan (≤ 4 item, ≤ 20 MB); penambahan ke-5 atau yang melewati total ditolak saat menambah |

Pesan error validasi menyebut nama berkas dan aturan yang dilanggar, tanpa path penuh. Frontend menampilkan error ini di banner `#chat-error` yang sudah ada.

## 8. UI

### Composer (`ChatPanel.tsx`)

- Tombol klip (`IconPaperclip`, baru) di kiri textarea, `aria-label="Lampirkan berkas"`, nonaktif saat `busy` atau `!hasProvider`.
- Baris `#chat-pending` di atas textarea, tampil hanya bila ada item: thumbnail 56×56 (`object-fit: cover`) untuk gambar, chip ikon+nama+ukuran untuk pdf/teks; tiap item tombol × `aria-label="Hapus lampiran <nama>"`. Ukuran ditulis `KB`/`MB` satu desimal.
- Kirim aktif bila `draft.trim() !== "" || pending.length > 0`. Enter mengirim meski teks kosong bila ada lampiran.
- `onPaste`: bila `clipboardData.files` berisi gambar, `preventDefault`, baca tiap gambar (maks sisa kuota 4), panggil `StageAttachmentData`. Teks biasa tidak diintervensi.
- Placeholder tetap; hint empty state ditambah satu kalimat: "Tempel atau seret gambar/PDF ke sini."
- `#chat-main` mendapat `--wails-drop-target: drop`. Saat `dragover` → overlay `#chat-drop` "Lepas untuk melampirkan"; `dragleave`/`drop` menyembunyikan. `OnFileDrop` didaftarkan di effect ChatPanel dan dilepas (`OnFileDropOff`) saat unmount, sehingga mode pet/settings tidak menerima drop.

### State (`App.tsx` + hook `useAttachments`)

- `pending: AttachmentMeta[]` hidup di `App.tsx` bersama `draft`, sehingga bertahan saat panel diciutkan ke pet (FR-003) dan saat pindah percakapan (sama seperti draft).
- `useAttachments` menyediakan `pending`, `error`, `pick()`, `stagePaths(paths)`, `stageFiles(files)`, `remove(id)`, `reset()`, `clearError()`. Setiap binding yang gagal mengisi `error` dan tidak mengubah `pending`. `reset()` hanya mengosongkan daftar lokal (backend sudah memindahkan item ke DB saat kirim).
- `useChat.send(text, attachments)`: bubble user lokal membawa `attachments` metadata dan `SendMessage` menerima id-nya. Composer mengosongkan draft dan daftar pending saat submit (aturan draft yang sudah ada); penolakan sinkron ditampilkan di banner seperti sekarang.

### Transkrip (`MessageItem.tsx`)

- Bubble user: di atas teks, grid thumbnail (maks lebar 160 px, `alt` = nama) dan chip berkas. Klik thumbnail → toggle ukuran penuh inline (maks lebar bubble). Tidak ada lightbox.
- `aria-label` grid: "Lampiran: n berkas".
- Bubble assistant tidak berubah.

### Styling

Token yang ada (`frontend/src/styles/chat.css`); tidak ada nilai mentah. Thumbnail memakai `--radius-sm`, chip memakai pola `.chip` yang ada.

## 9. Keamanan dan privasi

- Berkas dibaca hanya dari aksi eksplisit pengguna (dialog, drop, paste). Tidak ada polling clipboard, tidak ada screenshot (AGENTS.md).
- Path penuh tidak masuk log, event, DB (kolom `name` = basename), maupun pesan error.
- Bytes lampiran tidak pernah di-log atau dikirim lewat event; hanya lewat request provider dan route asset lokal.
- Route asset: UUID ketat, hanya id yang ada di staging/DB, `nosniff`. Tidak melayani path lain.
- PDF tidak dirender di WebView, hanya chip. Isi berkas teks dikirim ke provider sebagai data pengguna; tidak dirender sebagai HTML.
- Tipe ditentukan dari isi (magic bytes), bukan ekstensi; teks diverifikasi UTF-8.
- Staging dibersihkan saat `shutdown` dan saat item dikirim/dihapus; tidak ada TTL.

## 10. Pengujian

### Go (test-first)

- Validasi: tiap aturan bagian 7 (magic bytes salah, melebihi ukuran, teks non-UTF-8, nama panjang/path).
- Staging: item ke-5 ditolak, total > 20 MB ditolak, unstage idempoten, unknown id saat send menolak seluruh kirim.
- `Message.MarshalJSON`: tanpa parts → string; dengan parts → array dengan urutan teks lalu lampiran; tiga jenis part.
- Kebijakan riwayat: 5 turn user berlampiran → 3 terakhir part penuh, 2 pertama placeholder; pesan assistant tidak tersentuh.
- Storage: migrasi v4→v5 pada DB lama berisi data; `AppendUserTurn` atomik (gagal di tengah → tidak ada row); cascade hapus percakapan menghapus `attachments`; `RecentContext` tidak memuat `data`.
- Route: id bukan UUID → 404; id tidak ada → 404; staging → 200 dengan mime; metode POST → 405.
- Error map: 413 → `payload_too_large`; 400 dengan body "image_url not supported" → `unsupported_input`; 400 lain → `bad_request`.
- `SendMessage` teks kosong + lampiran → diterima; keduanya kosong → ditolak.

### Vitest (test-first)

- `useAttachments`: pick/stagePaths/stageFiles/remove/reset, error binding tidak mengubah `pending`, item ke-5 ditolak lokal tanpa memanggil backend.
- `ChatPanel`: tombol klip ada dan nonaktif saat busy; baris pending tampil dengan ×; kirim aktif tanpa teks bila ada lampiran; Enter mengirim `attachmentIds`; paste gambar memanggil stage dan tidak mengisi draft; `OnFileDropOff` dipanggil saat unmount.
- `MessageItem`: user dengan gambar → `<img alt=nama src="/attachments/{id}">`; pdf → chip nama; tanpa lampiran → DOM tidak berubah dari sekarang.
- `useChat`: `send` meneruskan ids; history dengan `attachments` masuk ke state.

### Manual (Kenari)

1. Model vision: kirim gambar + pertanyaan → jawaban menyebut isi gambar; follow-up tanpa gambar masih nyambung (turn ke-2 dan ke-3), turn ke-5 memakai placeholder (cek payload lewat log debug sementara, dihapus setelah uji).
2. PDF ber-text-layer → ringkasan benar. PDF > 15 MB → ditolak lokal sebelum request.
3. Model non-vision → pesan "Model ini tidak menerima gambar/berkas…".
4. Drop ke panel chat saat mode chat → terlampir; drop saat mode pet → tidak ada reaksi.
5. Ctrl+V screenshot → thumbnail muncul; Ctrl+V teks → masuk draft seperti biasa.
6. Restart aplikasi → thumbnail riwayat tampil dari DB; hapus percakapan → `SELECT COUNT(*) FROM attachments` turun.
7. Build release `wails build`, ukuran exe dan RAM idle tidak berubah berarti.

## 11. Dokumen yang ikut diperbarui

- `docs/02-mvp-requirements.md`: FR baru "Lampiran gambar dan berkas (P1)" dengan kriteria dari bagian 7–8.
- `docs/05-ai-provider-integration.md`: bagian 4 (bentuk request multimodal), bagian 8 (413/400 unsupported), bagian 9 (multimodal masuk scope).
- `docs/07-storage-security-privacy.md`: tabel `attachments`.
- `docs/09-decisions.md`: ADR-054 "Lampiran: staging Go, BLOB SQLite, 3 turn, tanpa gating".
- `docs/10-frontend-ui-inventory.md`: composer + MessageItem.
- `README.md`: fitur dan batas ukuran.

## 12. Risiko

- Screenshot PNG besar tanpa downscale → token input mahal; dicatat di ADR, downscale jadi kandidat lanjutan.
- `OnFileDrop` Wails membutuhkan `--wails-drop-target` pada elemen dan dapat berinteraksi dengan `--wails-draggable` di header panel; wajib diuji di build nyata sebelum dinyatakan selesai.
- Provider OpenAI-compatible lain mungkin tidak memahami part `file`; error dipetakan ke `unsupported_input`, pengguna diarahkan memilih model/provider lain.
- Ukuran DB bertambah sebesar lampiran; hapus percakapan membebaskan halaman untuk dipakai ulang, bukan mengecilkan file. Dicatat di docs 07.
