import {useCallback, useEffect, useState} from 'react';
import {
    EventsOff,
    EventsOn,
    Quit,
} from '../wailsjs/runtime/runtime';
import {AutostartEnabled, GetAppearance, GetShortcut, HideWindow, HotkeyError, SetAppearance, SetAutostart, SetShortcut, SetWindowMode, Version} from '../wailsjs/go/main/App';
import {app} from '../wailsjs/go/models';
import {ChatHeader} from './components/chat/ChatHeader';
import {ChatPanel} from './components/chat/ChatPanel';
import {ModelPicker} from './components/chat/ModelPicker';
import {PetSurface} from './components/pet/PetSurface';
import {SettingsPanel} from './components/settings/SettingsPanel';
import {useChat} from './hooks/useChat';
import {usePetDrag} from './hooks/usePetDrag';
import {usePetSignal} from './hooks/usePetSignal';
import {useProviders} from './hooks/useProviders';

type Mode = 'pet' | 'chat' | 'settings';

// Backend owns the sizes and keeps the window inside the monitor work area,
// so a pet parked near an edge opens a panel that is not cut off (ADR-042).
// Awaited before the view switches so content never renders clipped in the
// old size.
async function fitWindow(mode: Mode): Promise<void> {
    try {
        await SetWindowMode(mode === 'pet' ? 'pet' : 'panel');
    } catch {
        // Only an IPC failure lands here: the backend logs placement errors
        // and falls back to a plain resize itself. Switching the view anyway
        // is the best the UI can do; there is no surface to report it on.
    }
}

async function switchTo(mode: Mode, setMode: (m: Mode) => void): Promise<void> {
    await fitWindow(mode);
    setMode(mode);
}

function App() {
    const [mode, setMode] = useState<Mode>('pet');
    const chat = useChat();
    const provs = useProviders();

    // M5: tray menu intents. Show -> chat mode, Settings -> settings mode.
    useEffect(() => {
        EventsOn('app:show', () => {
            void switchTo('chat', setMode);
        });
        EventsOn('app:settings', () => {
            void switchTo('settings', setMode);
        });
        EventsOn('app:toggle', () => {
            void switchTo('chat', setMode);
        });
        // Global shortcut while the panel is open: back to the pet (ADR-050).
        EventsOn('app:pet', () => {
            void switchTo('pet', setMode);
        });
        return () => {
            EventsOff('app:show');
            EventsOff('app:settings');
            EventsOff('app:toggle');
            EventsOff('app:pet');
        };
    }, []);

    // Both surfaces are the same window, so every mode change goes through
    // switchTo — otherwise the pet renders inside a 400x600 window for a frame.
    const showChat = useCallback(() => {
        void switchTo('chat', setMode);
    }, []);

    const showPet = useCallback(() => {
        void switchTo('pet', setMode);
    }, []);

    // Pet click opens chat, pet drag moves the window. The 4 px threshold lives
    // in usePetDrag (ADR-041): a plain --wails-draggable surface would start the
    // move on the first mousemove and swallow the click (ADR-029).
    const toggle = useCallback(() => {
        void switchTo(mode === 'pet' ? 'chat' : 'pet', setMode);
    }, [mode]);
    const petDrag = usePetDrag(toggle);
    const pet = usePetSignal(mode, chat.status);

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
    const [version, setVersion] = useState<string>("");

    useEffect(() => {
        AutostartEnabled().then(setAutostartOn).catch(() => setAutostartOn(false));
        HotkeyError().then(setHotkeyErr).catch(() => setHotkeyErr(""));
        GetAppearance().then(setAppearance).catch(() => {});
        GetShortcut().then(setShortcut).catch(() => {});
        Version().then(setVersion).catch(() => setVersion("dev"));
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
                setShortcutErr('Pakai minimal satu tombol Ctrl, Alt, Shift, atau Win.');
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
            <PetSurface
                status={pet.petStatus}
                replyReady={pet.replyReady}
                drag={petDrag}
                onHide={() => HideWindow()}
                onQuit={Quit}
                shortcut={shortcut.enabled && !hotkeyErr ? shortcut.combo : undefined}
            />
        );
    }

    if (mode === 'settings') {
        return (
            <div id="chat">
                <SettingsPanel providers={provs.providers} activeId={provs.activeId} listError={provs.error} test={provs.test} testing={provs.testing} onSave={provs.save} onDelete={provs.remove} onActivate={provs.activate} onTest={provs.runTest} onClearTest={provs.clearTest} onClose={showChat} autostart={autostart} autostartErr={autostartErr} onToggleAutostart={toggleAutostart} hotkeyErr={hotkeyErr} appearance={appearance} appearanceErr={appearanceErr} onAppearanceChange={changeAppearance} shortcut={shortcut} shortcutErr={shortcutErr} onShortcutChange={changeShortcut} capturing={capturing} onCaptureToggle={() => setCapturing((v) => !v)} version={version} />
            </div>
        );
    }

    return (
        <div id="chat">
            <ChatHeader
                onSettings={() => setMode('settings')}
                onPet={showPet}
                onHide={() => HideWindow()}
                onQuit={Quit}
            />
            <ChatPanel messages={chat.messages} error={chat.error} busy={chat.busy} conversations={chat.conversations} conversationId={chat.conversationId} onSend={chat.send} onCancel={chat.cancel} onRetry={chat.retry} onDismissError={chat.clearError} onNew={chat.newConversation} onSwitch={chat.switchConversation} onRename={chat.renameConversation} onDelete={chat.deleteConversation} draft={draft} onDraftChange={setDraft} modelPicker={<ModelPicker provider={provs.activeProvider} busy={chat.busy} onModelChange={provs.setModel} />} status={chat.status} onOpenSettings={() => setMode('settings')} hasProvider={provs.activeProvider !== null} />
        </div>
    );
}

export default App;
