# 04 — Struktur Project dan Tanggung Jawab Modul

## 1. Struktur yang disarankan

```text
wortel-ai/
├── main.go
├── app.go
├── go.mod
├── go.sum
├── wails.json
├── README.md
├── AGENTS.md
├── LICENSE
├── CONTRIBUTING.md
├── CODE_OF_CONDUCT.md
├── SECURITY.md
├── .gitattributes
├── .editorconfig
├── .gitignore
├── .github/
│   └── workflows/
│       ├── ci.yml
│       └── release.yml
├── docs/
│   ├── 01-product-brief.md
│   ├── 02-mvp-requirements.md
│   ├── 03-architecture.md
│   ├── 04-project-structure.md
│   ├── 05-ai-provider-integration.md
│   ├── 06-ui-ux.md
│   ├── 07-storage-security-privacy.md
│   ├── 08-roadmap-testing.md
│   └── 09-decisions.md
├── internal/
│   ├── app/
│   │   ├── service.go
│   │   ├── events.go
│   │   └── errors.go
│   ├── ai/
│   │   ├── client.go
│   │   ├── openai_compatible.go
│   │   ├── sse.go
│   │   ├── types.go
│   │   └── errors.go
│   ├── config/
│   │   ├── settings.go
│   │   └── validation.go
│   ├── credentials/
│   │   ├── store.go
│   │   └── windows.go
│   ├── storage/
│   │   ├── sqlite.go
│   │   ├── migrations/
│   │   ├── conversations.go
│   │   ├── messages.go
│   │   └── providers.go
│   └── platform/
│       └── windows/
│           ├── window.go
│           ├── hotkey.go
│           ├── tray.go
│           └── startup.go
├── frontend/
│   ├── package.json
│   ├── package-lock.json
│   ├── vite.config.ts
│   ├── tsconfig.json
│   ├── src/
│   │   ├── app/
│   │   │   ├── App.tsx
│   │   │   └── routes.ts
│   │   ├── components/
│   │   │   ├── companion/
│   │   │   │   ├── Companion.tsx
│   │   │   │   └── CompanionStatus.tsx
│   │   │   ├── chat/
│   │   │   │   ├── ChatPanel.tsx
│   │   │   │   ├── MessageList.tsx
│   │   │   │   ├── MessageBubble.tsx
│   │   │   │   ├── ChatComposer.tsx
│   │   │   │   └── StreamingIndicator.tsx
│   │   │   ├── conversations/
│   │   │   │   ├── ConversationList.tsx
│   │   │   │   └── ConversationActions.tsx
│   │   │   ├── settings/
│   │   │   │   ├── SettingsPanel.tsx
│   │   │   │   ├── ProviderForm.tsx
│   │   │   │   └── AppearanceSettings.tsx
│   │   │   └── common/
│   │   │       ├── IconButton.tsx
│   │   │       ├── Dialog.tsx
│   │   │       └── ErrorNotice.tsx
│   │   ├── hooks/
│   │   │   ├── useChat.ts
│   │   │   ├── useConversations.ts
│   │   │   └── useProviders.ts
│   │   ├── services/
│   │   │   └── backend.ts
│   │   ├── stores/
│   │   │   └── uiStore.ts
│   │   ├── types/
│   │   │   ├── chat.ts
│   │   │   ├── provider.ts
│   │   │   └── settings.ts
│   │   ├── styles/
│   │   │   └── index.css
│   │   └── main.tsx
│   └── public/
│       └── assets/
│           └── companion/
│               ├── idle.png
│               ├── thinking.png
│               └── error.png
└── build/
    └── windows/
```

Struktur di atas adalah target yang mungkin dicapai, bukan alasan untuk membuat semua file di hari pertama. Buat folder saat ada kode yang benar-benar memerlukannya. Folder `migrations/` hanya perlu menampung migration yang benar-benar dibuat.

## 2. Entry points

### `main.go`

- Menyiapkan konfigurasi awal dan dependency.
- Membuat layanan storage, credential store, AI client, app service.
- Mengonfigurasi dan menjalankan Wails.
- Tidak menyimpan business logic di sini.

### `app.go`

- Adapter tipis untuk method yang dipanggil melalui Wails.
- Menyediakan lifecycle hooks jika diperlukan.
- Meneruskan kerja ke `internal/app` atau service terkait.

## 3. Tanggung jawab paket backend

### `internal/app`

Orkestrasi operasi aplikasi dan event. Ini bukan tempat implementasi SSE atau query SQL detail.

### `internal/ai`

- Kontrak request/response internal.
- Client OpenAI-compatible.
- Normalisasi URL.
- Parser SSE.
- Error mapping dan cancel.

Paket ini tidak boleh bergantung pada React maupun database SQLite.

### `internal/config`

- Pengaturan aplikasi yang bukan rahasia.
- Validasi base URL, model ID, batas output, dan prefensi appearance.
- Tidak memegang API key permanen.

### `internal/credentials`

- Interface untuk simpan/ambil/hapus secret.
- Implementasi Windows Credential Manager atau penyimpanan yang dipilih setelah spike.
- Memastikan secret tidak dimasukkan ke log/error string.

### `internal/storage`

- Inisialisasi SQLite dan migrations.
- Repository untuk conversations/messages.
- Transaksi dan validasi konsistensi data.
- Tidak menjalankan request jaringan.

### `internal/platform/windows`

- Adapter integrasi native Windows yang benar-benar dibutuhkan.
- Gunakan interface/fallback jika fungsi opsional gagal.
- Kode platform tidak boleh menyebar ke semua paket.

## 4. Frontend: aturan pembagian

- `components/`: komponen tampilan, jangan menyimpan semua request/state di satu file.
- `hooks/`: alur stateful, misalnya `useChat` untuk streaming event dan send/cancel.
- `services/backend.ts`: wrapper typed untuk fungsi Go dan event Wails. Hindari memanggil runtime API di banyak komponen.
- `types/`: definisi kontrak yang mencerminkan payload backend.
- `stores/`: hanya state global yang memang dibutuhkan. State input lokal tetap berada di komponen.

Jangan menggabungkan Redux/Zustand/query framework hanya karena tersedia. Mulai dengan React state/hooks; tambah library bila kompleksitas nyata membutuhkannya.

## 5. Konvensi penamaan

### Go
- Package name pendek, lowercase, satu konsep per package.
- Error penting menggunakan sentinel atau typed error yang jelas.
- Context diteruskan untuk request jaringan dan operasi yang dapat dibatalkan.
- Hindari global mutable state jika dependency injection sederhana bisa digunakan.

### TypeScript
- Komponen React: `PascalCase.tsx`.
- Hooks: `useSomething.ts`.
- Fungsi/variabel: `camelCase`.
- API payload gunakan tipe eksplisit; hindari `any` tanpa alasan.
- Semua state request harus memiliki lifecycle jelas.

## 6. Batas dependensi

Arah dependensi yang diinginkan:

```text
Wails adapter -> app service -> (AI service, repositories, platform interfaces)
AI service -> standard library HTTP + internal types
Repositories -> SQLite
Frontend -> Wails bindings/events
```

Hindari siklus import. `internal/ai` tidak perlu mengetahui `internal/storage`; app service yang mengatur penyimpanan hasil.

## 7. Pengelolaan aset karakter

Untuk MVP, gunakan satu set asset lokal statis dengan ukuran wajar. Pilihan teknis:

1. PNG/WebP terpisah untuk state idle/thinking/error — paling sederhana.
2. Sprite sheet — cocok jika ada beberapa frame pixel-art.
3. Animasi berbasis CSS — cocok bila karakter dirancang dari bentuk vector sederhana.

Lottie bukan keharusan. Tambahkan hanya bila ada manfaat visual dan hasil pengukuran tetap memuaskan.

## 8. File yang tidak boleh masuk Git

- API key asli dan file `.env` berisi secret.
- Database pengguna dan percakapan nyata.
- Credential dump, log berisi data pribadi, crash dump sensitif.
- File build besar yang dihasilkan otomatis, kecuali ada kebutuhan distribusi yang jelas.

Tambahkan `.gitignore` sesuai Go, Node/Vite, Wails, SQLite, editor, dan file runtime lokal.

## 9. Aksesibilitas rilis: file OSS dan CI

- `LICENSE` (MIT) — syarat publikasi open source dan kualifikasi SignPath Foundation.
- `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, `SECURITY.md` — alur kontribusi dan pelaporan kerentanan.
- `.gitattributes` — normalisasi EOL: repo menyimpan LF, kecuali `*.ps1`/`*.bat` tetap CRLF.
- `.editorconfig` — indentasi dan EOL per jenis file.
- `.github/workflows/ci.yml` — gate wajib: `gofmt`, `go vet`, `go test`, `go test -race` (runner Windows), typecheck + `npm test` + build frontend, lalu `wails build`.
- `.github/workflows/release.yml` — dipicu tag `v*`: build, penandatanganan SignPath opsional (environment `release` + `vars.SIGNPATH_ENABLED`), lalu GitHub release draft + `SHA256SUMS.txt`.

CI adalah sumber kebenaran gate; perintah di dokumen ini harus tetap cocok dengan workflow. Lihat ADR-040 untuk alasan pilihan.
