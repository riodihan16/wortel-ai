import {useCallback, useRef, useState} from 'react';
import type {CSSProperties} from 'react';
import './App.css';
import {
    Quit,
    WindowCenter,
    WindowGetPosition,
    WindowGetSize,
    WindowSetAlwaysOnTop,
    WindowSetSize,
} from '../wailsjs/runtime/runtime';
import {ChatPanel} from './components/chat/ChatPanel';
import {useChat} from './hooks/useChat';

type Mode = 'pet' | 'chat';

const PET = {w: 180, h: 180};
const CHAT = {w: 400, h: 600};

function App() {
    const [mode, setMode] = useState<Mode>('pet');
    const [info, setInfo] = useState<string>('probe belum jalan');
    const [top, setTop] = useState<boolean>(true);
    const chat = useChat();
    const lastUser = useRef<string>('');

    const send = useCallback((text: string) => {
        lastUser.current = text.trim();
        chat.send(text);
    }, [chat]);

    const retry = useCallback(() => {
        if (lastUser.current) chat.send(lastUser.current);
    }, [chat]);

    const toggle = useCallback(() => {
        const next: Mode = mode === 'pet' ? 'chat' : 'pet';
        const s = next === 'pet' ? PET : CHAT;
        WindowSetSize(s.w, s.h);
        setMode(next);
    }, [mode]);

    const probe = useCallback(async () => {
        const size = await WindowGetSize();
        const pos = await WindowGetPosition();
        setInfo(`${size.w}x${size.h} @ ${pos.x},${pos.y}`);
    }, []);

    const flipTop = useCallback(() => {
        WindowSetAlwaysOnTop(!top);
        setTop(!top);
    }, [top]);

    if (mode === 'pet') {
        return (
            <div id="pet" style={{'--wails-draggable': 'drag'} as CSSProperties}>
                <button id="pet-dot" onClick={toggle} title="Buka chat">W</button>
                <div id="pet-hint">drag area, klik W buka chat</div>
                <div id="pet-actions">
                    <button onClick={probe}>Probe</button>
                    <button onClick={() => WindowCenter()}>Center</button>
                    <button onClick={Quit}>Quit</button>
                </div>
                <div id="pet-info">{info}</div>
            </div>
        );
    }

    return (
        <div id="chat">
            <div id="chat-header" style={{'--wails-draggable': 'drag'} as CSSProperties}>
                <span>Wortel AI</span>
                <span>
                    <button onClick={toggle}>Pet</button>
                    <button onClick={Quit}>Quit</button>
                </span>
            </div>
            <ChatPanel messages={chat.messages} status={chat.status} error={chat.error} busy={chat.busy} onSend={send} onCancel={chat.cancel} onRetry={retry} />
            <div id="chat-body">
                <div id="chat-actions">
                    <button onClick={probe}>Probe</button>
                    <button onClick={flipTop}>Top:{top ? 'on' : 'off'}</button>
                    <button onClick={() => WindowCenter()}>Center</button>
                </div>
                <div id="chat-info">{info}</div>
            </div>
        </div>
    );
}

export default App;
