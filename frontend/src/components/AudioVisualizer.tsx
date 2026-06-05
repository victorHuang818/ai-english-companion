import React, { useEffect, useRef } from 'react';
import './AudioVisualizer.css';

interface AudioVisualizerProps {
  stream?: MediaStream | null;
  isActive: boolean;
  color?: 'primary' | 'accent';
}

export const AudioVisualizer: React.FC<AudioVisualizerProps> = ({ 
  stream, 
  isActive,
  color = 'primary'
}) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const animationRef = useRef<number | null>(null);
  const analyzerRef = useRef<AnalyserNode | null>(null);


  useEffect(() => {
    if (!stream || !isActive || !canvasRef.current) return;

    const audioCtx = new (window.AudioContext || (window as any).webkitAudioContext)();
    const analyzer = audioCtx.createAnalyser();
    analyzer.fftSize = 256;
    
    const source = audioCtx.createMediaStreamSource(stream);
    source.connect(analyzer);
    analyzerRef.current = analyzer;

    const canvas = canvasRef.current;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    const bufferLength = analyzer.frequencyBinCount;
    const dataArray = new Uint8Array(bufferLength);

    const draw = () => {
      if (!ctx || !canvas) return;
      
      const width = canvas.width;
      const height = canvas.height;
      
      animationRef.current = requestAnimationFrame(draw);
      analyzer.getByteFrequencyData(dataArray);

      ctx.clearRect(0, 0, width, height);

      const barWidth = (width / bufferLength) * 2.5;
      let barHeight;
      let x = 0;

      for (let i = 0; i < bufferLength; i++) {
        barHeight = dataArray[i] / 2;

        const gradient = ctx.createLinearGradient(0, height, 0, 0);
        if (color === 'primary') {
          gradient.addColorStop(0, 'rgba(102, 252, 241, 0.2)');
          gradient.addColorStop(1, 'rgba(102, 252, 241, 0.8)');
        } else {
          gradient.addColorStop(0, 'rgba(197, 179, 255, 0.2)');
          gradient.addColorStop(1, 'rgba(197, 179, 255, 0.8)');
        }

        ctx.fillStyle = gradient;
        ctx.fillRect(x, height - barHeight, barWidth, barHeight);

        x += barWidth + 1;
      }
    };

    draw();

    return () => {
      if (animationRef.current) {
        cancelAnimationFrame(animationRef.current);
      }
      source.disconnect();
      audioCtx.close();
    };
  }, [stream, isActive, color]);

  return (
    <div className={`audio-visualizer-container ${isActive ? 'active' : ''}`}>
      <canvas 
        ref={canvasRef} 
        width={300} 
        height={100} 
        className="audio-canvas"
      />
      {!isActive && (
        <div className="audio-placeholder">
          Waiting for audio...
        </div>
      )}
    </div>
  );
};
