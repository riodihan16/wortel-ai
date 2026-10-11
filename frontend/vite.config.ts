import {defineConfig} from 'vite'
import type {Plugin} from 'vite'
import react from '@vitejs/plugin-react'

// Under `wails dev`, Wails proxies GETs to Vite and falls back to the Go
// asset handler only on 404/405. Vite's SPA fallback answers an <img>
// request (Accept: */*) with index.html and 200, so attachment thumbnails
// never reached Go. Answer /attachments/ with 404 so Wails asks Go instead.
// Dev server only; the production build has no Vite in the path.
function attachmentsToGo(): Plugin {
  return {
    name: 'wortel-attachments-to-go',
    apply: 'serve',
    configureServer(server) {
      // Registered before Vite's own middlewares, so the SPA fallback never
      // sees these requests. Connect matches the route as a path prefix.
      server.middlewares.use('/attachments/', (_req, res) => {
        res.statusCode = 404
        res.end()
      })
    }
  }
}

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react(), attachmentsToGo()]
})
