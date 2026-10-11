import React from 'react'
import {createRoot} from 'react-dom/client'
import './styles/index.css'
import App from './App'

// A file dropped anywhere outside the chat drop zone (pet, Settings, panel
// header) must not make WebView2 navigate to it. The Wails runtime has its
// own drop listener and decides by the target's CSS whether to report the
// drop; cancelling the default here does not interfere with that. Only file
// drags are cancelled: text drags (into the composer, within the textarea)
// keep their default behaviour.
const isFileDrag = (e: DragEvent) => Array.from(e.dataTransfer?.types ?? []).includes("Files");
window.addEventListener("dragover", (e) => {
  if (isFileDrag(e)) e.preventDefault();
});
window.addEventListener("drop", (e) => {
  if (isFileDrag(e)) e.preventDefault();
});

const container = document.getElementById('root')

const root = createRoot(container!)

root.render(
    <React.StrictMode>
        <App/>
    </React.StrictMode>
)
