import { useState, useRef, useCallback } from 'react';

export const useMicrophone = (onAudioData: (base64Data: string) => void) => {
  const [isRecording, setIsRecording] = useState(false);
  const [stream, setStream] = useState<MediaStream | null>(null);
  
  const audioContextRef = useRef<AudioContext | null>(null);
  const processorRef = useRef<ScriptProcessorNode | null>(null);
  const streamRef = useRef<MediaStream | null>(null);

  const startRecording = useCallback(async () => {
    try {
      const mediaStream = await navigator.mediaDevices.getUserMedia({ audio: true });
      setStream(mediaStream);
      streamRef.current = mediaStream;

      const audioCtx = new (window.AudioContext || (window as any).webkitAudioContext)({
        sampleRate: 16000 // Force 16kHz for Gemini
      });
      audioContextRef.current = audioCtx;

      const source = audioCtx.createMediaStreamSource(mediaStream);
      
      // Using ScriptProcessor for ease of PCM extraction in single file
      // 4096 buffer size is a good balance for latency
      const processor = audioCtx.createScriptProcessor(4096, 1, 1);
      processorRef.current = processor;

      processor.onaudioprocess = (e) => {
        const inputData = e.inputBuffer.getChannelData(0);
        // Convert Float32 to Int16
        const pcmData = new Int16Array(inputData.length);
        for (let i = 0; i < inputData.length; i++) {
          const s = Math.max(-1, Math.min(1, inputData[i]));
          pcmData[i] = s < 0 ? s * 0x8000 : s * 0x7FFF;
        }

        // Convert Int16Array to Base64
        const uint8Array = new Uint8Array(pcmData.buffer);
        let binary = '';
        for (let i = 0; i < uint8Array.byteLength; i++) {
          binary += String.fromCharCode(uint8Array[i]);
        }
        const base64Data = btoa(binary);
        
        onAudioData(base64Data);
      };

      source.connect(processor);
      processor.connect(audioCtx.destination);
      
      setIsRecording(true);
    } catch (err) {
      console.error("Microphone access denied or error:", err);
    }
  }, [onAudioData, isRecording]);

  const stopRecording = useCallback(() => {
    if (processorRef.current) {
      processorRef.current.disconnect();
      processorRef.current = null;
    }
    if (audioContextRef.current) {
      audioContextRef.current.close();
      audioContextRef.current = null;
    }
    if (streamRef.current) {
      streamRef.current.getTracks().forEach(track => track.stop());
      streamRef.current = null;
    }
    setStream(null);
    setIsRecording(false);
  }, []);

  return { isRecording, startRecording, stopRecording, stream };
};
