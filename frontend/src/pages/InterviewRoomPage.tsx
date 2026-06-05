import React, { useEffect, useState, useRef, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { GlassCard } from '../components/GlassCard';
import { GlowingButton } from '../components/GlowingButton';
import { AudioVisualizer } from '../components/AudioVisualizer';
import { useMicrophone } from '../hooks/useMicrophone';
import { Mic, MicOff, PhoneOff } from 'lucide-react';
import { useAuthStore } from '../store/authStore';
import './InterviewRoomPage.css';

interface ChatMessage {
  id: string;
  sender: 'ai' | 'user';
  text: string;
}

export const InterviewRoomPage: React.FC = () => {
  const { sessionId } = useParams<{ sessionId: string }>();
  const navigate = useNavigate();

  const [connected, setConnected] = useState(false);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [suggestion, setSuggestion] = useState('');
  const [loadingSuggestion, setLoadingSuggestion] = useState(false);
  const [error, setError] = useState('');

  const wsRef = useRef<WebSocket | null>(null);
  const audioContextRef = useRef<AudioContext | null>(null);
  const audioQueueRef = useRef<AudioBuffer[]>([]);
  const isPlayingRef = useRef(false);

  const playNextAudio = useCallback(() => {
    if (audioQueueRef.current.length === 0 || isPlayingRef.current || !audioContextRef.current) return;
    isPlayingRef.current = true;
    const audioBuffer = audioQueueRef.current.shift()!;
    const source = audioContextRef.current.createBufferSource();
    source.buffer = audioBuffer;
    source.connect(audioContextRef.current.destination);
    source.onended = () => { isPlayingRef.current = false; playNextAudio(); };
    source.start();
  }, []);

  const decodeAndQueueAudio = useCallback((base64PCM: string, mimeType?: string) => {
    if (!audioContextRef.current) return;
    const binaryStr = atob(base64PCM);
    const bytes = new Uint8Array(binaryStr.length);
    for (let i = 0; i < binaryStr.length; i++) bytes[i] = binaryStr.charCodeAt(i);
    const int16Array = new Int16Array(bytes.buffer);
    const float32Array = new Float32Array(int16Array.length);
    for (let i = 0; i < int16Array.length; i++) float32Array[i] = int16Array[i] / (int16Array[i] < 0 ? 0x8000 : 0x7FFF);

    let sampleRate = 16000;
    if (mimeType) {
      const match = mimeType.match(/rate=(\d+)/);
      if (match) {
        sampleRate = parseInt(match[1], 10);
      }
    }
    const audioBuffer = audioContextRef.current.createBuffer(1, float32Array.length, sampleRate);
    audioBuffer.getChannelData(0).set(float32Array);
    audioQueueRef.current.push(audioBuffer);
    playNextAudio();
  }, [playNextAudio]);

  const handleAudioData = useCallback((base64Data: string) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      const binaryStr = atob(base64Data);
      const bytes = new Uint8Array(binaryStr.length);
      for (let i = 0; i < binaryStr.length; i++) bytes[i] = binaryStr.charCodeAt(i);
      wsRef.current.send(bytes.buffer);
    }
  }, []);

  const { isRecording, startRecording, stopRecording, stream } = useMicrophone(handleAudioData);

  useEffect(() => {
    if (!sessionId) { navigate('/'); return; }
    audioContextRef.current = new (window.AudioContext || (window as any).webkitAudioContext)({ sampleRate: 16000 });
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const token = useAuthStore.getState().token;
    const ws = new WebSocket(`${protocol}//${window.location.host}/ws/interview?token=${encodeURIComponent(token || '')}`);
    wsRef.current = ws;

    ws.onopen = () => { setConnected(true); ws.send(JSON.stringify({ session_id: sessionId })); };
    ws.onmessage = async (event) => {
      try {
        const data = JSON.parse(event.data instanceof Blob ? await event.data.text() : event.data);
        if (data.serverContent?.inputTranscription?.text) {
          setMessages(prev => [...prev, { id: 'user-' + Date.now(), sender: 'user', text: data.serverContent.inputTranscription.text }]);
        }
        if (data.serverContent?.modelTurn?.parts) {
          data.serverContent.modelTurn.parts.forEach((part: any) => {
            if (part.inlineData?.data) {
              decodeAndQueueAudio(part.inlineData.data, part.inlineData.mimeType);
            }
          });
        }
        if (data.serverContent?.outputTranscription?.text) {
          setMessages(prev => [...prev, { id: 'ai-' + Date.now(), sender: 'ai', text: data.serverContent.outputTranscription.text }]);
          setLoadingSuggestion(true);
        }
        if (data.aiSuggestion) { setSuggestion(data.aiSuggestion.suggestion); setLoadingSuggestion(false); }
        


        if (data.error === 'balance_insufficient') { setError('余额不足，面试已终止。'); ws.close(); }
      } catch (err) { console.error("WS error", err); }
    };
    ws.onclose = () => { setConnected(false); stopRecording(); };
    return () => ws.close();
  }, [sessionId, navigate, decodeAndQueueAudio, stopRecording]);

  return (
    <div className="interview-container">
      {error && <div className="error-toast">{error}</div>}
      <div className="interview-grid">
        <div className="interview-left-col">
          <GlassCard className="visualizer-card">
            <div className="status-indicator">
              <div className={`led ${connected ? 'led-green' : 'led-red'}`}></div>
              <span>{connected ? 'Active' : 'Connecting...'}</span>
            </div>
            <div className="avatar-area"><div className={`ai-avatar ${isPlayingRef.current ? 'speaking' : ''}`}>AI</div></div>
            <AudioVisualizer stream={stream} isActive={isRecording} color="primary" />
            <div className="controls">
              <GlowingButton variant={isRecording ? 'danger' : 'primary'} onClick={() => isRecording ? stopRecording() : startRecording()} disabled={!connected}>
                {isRecording ? <MicOff size={24} /> : <Mic size={24} />}
              </GlowingButton>
              <GlowingButton variant="danger" onClick={() => navigate('/')}><PhoneOff size={24} /></GlowingButton>
            </div>
          </GlassCard>


        </div>

        <div className="interview-right-col">
          <GlassCard className="chat-card">
            <div className="chat-header"><h3>Live Transcription</h3></div>
            <div className="chat-messages">
              {messages.length === 0 ? <div className="empty-chat">Awaiting...</div> : messages.map(msg => (
                <div key={msg.id} className={`message ${msg.sender}`}><div className="message-content">{msg.text}</div></div>
              ))}
            </div>
          </GlassCard>
          <GlassCard className="suggestion-card-horizontal">
            <div className="card-header"><span className="accent-label">AI HINT</span></div>
            <div className="suggestion-content">
              {loadingSuggestion ? <div className="typing-loader">Thinking...</div> : <p>{suggestion || 'Hints will appear here...'}</p>}
            </div>
          </GlassCard>
        </div>
      </div>
    </div>
  );
};
