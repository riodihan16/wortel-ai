import {useCallback, useEffect, useRef, useState} from 'react';
import type {CSSProperties} from 'react';
import './App.css';
import {
    EventsOff,
    EventsOn,
    Quit,
    WindowSetSize,
} from '../wailsjs/runtime/runtime';
import {AutostartEnabled, GetAppearance, GetShortcut, HideWindow, HotkeyError, SetAppearance, SetAutostart, SetShortcut} from '../wailsjs/go/main/App';
import {app} from '../wailsjs/go/models';
import {ChatPanel} from './components/chat/ChatPanel';
import {SettingsPanel} from './components/settings/SettingsPanel';
import {useChat} from './hooks/useChat';
import {useProviders} from './hooks/useProviders';

type Mode = 'pet' | 'chat' | 'settings';

const PET = {w: 180, h: 180};
const CHAT = {w: 400, h: 600};

function App() {
    const [mode, setMode] = useState<Mode>('pet');
    const chat = useChat();
    const provs = useProviders();
    const lastUser = useRef<string>('');

    const send = useCallback((text: string) => {
        lastUser.current = text.trim();
        chat.send(text);
    }, [chat]);

    // M5: tray menu intents. Show -> chat mode, Settings -> settings mode.
    useEffect(() => {
        EventsOn('app:show', () => {
            WindowSetSize(CHAT.w, CHAT.h);
            setMode('chat');
        });
        EventsOn('app:settings', () => {
            WindowSetSize(CHAT.w, CHAT.h);
            setMode('settings');
        });
        EventsOn('app:toggle', () => {
            WindowSetSize(CHAT.w, CHAT.h);
            setMode('chat');
        });
        return () => {
            EventsOff('app:show');
            EventsOff('app:settings');
            EventsOff('app:toggle');
        };
    }, []);

    const retry = useCallback(() => {
        if (lastUser.current) chat.send(lastUser.current);
    }, [chat]);

    const toggle = useCallback(() => {
        const next: Mode = mode === 'pet' ? 'chat' : 'pet';
        const s = next === 'pet' ? PET : CHAT;
        WindowSetSize(s.w, s.h);
        setMode(next);
    }, [mode]);

    const [autostart, setAutostartOn] = useState<boolean>(false);
    const [autostartErr, setAutostartErr] = useState<string | null>(null);
    const [hotkeyErr, setHotkeyErr] = useState<string>("");
    const [appearance, setAppearance] = useState<app.Appearance>({theme: 'dark', fontScale: 1});
    const [appearanceErr, setAppearanceErr] = useState<string>("");
    const [shortcut, setShortcut] = useState<app.Shortcut>({enabled: true, combo: 'Ctrl+Shift+W'});
    const [shortcutErr, setShortcutErr] = useState<string>("");
    const [capturing, setCapturing] = useState<boolean>(false);
    // Draft lives here, not in ChatPanel: switching to pet mode unmounts the
    // panel, and the draft must survive that (FR-003).
    const [draft, setDraft] = useState<string>("");

    useEffect(() => {
        AutostartEnabled().then(setAutostartOn).catch(() => setAutostartOn(false));
        HotkeyError().then(setHotkeyErr).catch(() => setHotkeyErr(""));
        GetAppearance().then(setAppearance).catch(() => {});
        GetShortcut().then(setShortcut).catch(() => {});
    }, []);

    const changeShortcut = useCallback(async (next: app.Shortcut) => {
        setShortcutErr("");
        try {
            await SetShortcut(next);
            setShortcut(next);
        } catch (e) {
            setShortcutErr(e instanceof Error ? e.message : String(e));
        }
        HotkeyError().then(setHotkeyErr).catch(() => {});
    }, []);

    // Capture the next combo while the user is rebinding. Runs only in the
    // capture state, so it cannot swallow normal typing elsewhere.
    useEffect(() => {
        if (!capturing) return;
        const onKey = (e: KeyboardEvent) => {
            e.preventDefault();
            if (e.key === 'Escape') { setCapturing(false); return; }
            if (['Control', 'Alt', 'Shift', 'Meta'].includes(e.key)) return;
            const mods: string[] = [];
            if (e.ctrlKey) mods.push('Ctrl');
            if (e.altKey) mods.push('Alt');
            if (e.shiftKey) mods.push('Shift');
            if (e.metaKey) mods.push('Win');
            if (mods.length === 0) {
                setShortcutErr('Shortcut butuh minimal satu modifier (Ctrl/Alt/Shift/Win)');
                return;
            }
            setCapturing(false);
            changeShortcut({enabled: true, combo: [...mods, e.key.length === 1 ? e.key.toUpperCase() : e.key].join('+')});
        };
        window.addEventListener('keydown', onKey, true);
        return () => window.removeEventListener('keydown', onKey, true);
    }, [capturing, changeShortcut]);

    // Apply appearance to the document root. Theme drives the CSS token
    // block; fontScale scales the base font size (docs 06 §6).
    useEffect(() => {
        const root = document.documentElement;
        root.dataset.theme = appearance.theme;
        root.style.setProperty('--font-scale', String(appearance.fontScale));
    }, [appearance]);

    const changeAppearance = useCallback(async (next: app.Appearance) => {
        setAppearanceErr("");
        try {
            await SetAppearance(next);
            setAppearance(next);
        } catch (e) {
            setAppearanceErr(e instanceof Error ? e.message : String(e));
        }
    }, []);

    const toggleAutostart = useCallback(async () => {
        const next = !autostart;
        setAutostartErr(null);
        try {
            await SetAutostart(next);
            setAutostartOn(next);
        } catch (e) {
            setAutostartErr(e instanceof Error ? e.message : String(e));
        }
    }, [autostart]);

    if (mode === 'pet') {
        return (
            <div id="pet" style={{'--wails-draggable': 'drag'} as CSSProperties}>
                <button id="pet-dot" onClick={toggle} title="Buka chat">W</button>
                <div id="pet-hint">drag area, klik W buka chat</div>
                <div id="pet-actions">
                    <button onClick={() => HideWindow()}>Hide</button>
                    <button onClick={Quit}>Quit</button>
                </div>
            </div>
        );
    }

    if (mode === 'settings') {
        return (
            <div id="chat">
                <SettingsPanel providers={provs.providers} activeId={provs.activeId} listError={provs.error} test={provs.test} testing={provs.testing} onSave={provs.save} onDelete={provs.remove} onActivate={provs.activate} onTest={provs.runTest} onClose={() => setMode('chat')} autostart={autostart} autostartErr={autostartErr} onToggleAutostart={toggleAutostart} hotkeyErr={hotkeyErr} appearance={appearance} appearanceErr={appearanceErr} onAppearanceChange={changeAppearance} shortcut={shortcut} shortcutErr={shortcutErr} onShortcutChange={changeShortcut} capturing={capturing} onCaptureToggle={() => setCapturing((v) => !v)} />
            </div>
        );
    }

    return (
        <div id="chat">
            <div id="chat-header" style={{'--wails-draggable': 'drag'} as CSSProperties}>
                <span>Wortel AI</span>
                <span>
                    <button onClick={() => setMode('settings')}>Settings</button>
                    <button onClick={toggle}>Pet</button>
                    <button onClick={() => HideWindow()}>Hide</button>
                    <button onClick={Quit}>Quit</button>
                </span>
            </div>
            <ChatPanel messages={chat.messages} status={chat.status} error={chat.error} busy={chat.busy} conversations={chat.conversations} conversationId={chat.conversationId} onSend={send} onCancel={chat.cancel} onRetry={retry} onNew={chat.newConversation} onSwitch={chat.switchConversation} onRename={chat.renameConversation} onDelete={chat.deleteConversation} draft={draft} onDraftChange={setDraft} />
        </div>
    );
}

export default App;
