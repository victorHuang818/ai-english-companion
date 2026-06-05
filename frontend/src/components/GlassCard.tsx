import React from 'react';
import './GlassCard.css';

interface GlassCardProps {
  children: React.ReactNode;
  className?: string;
  glow?: boolean;
}

export const GlassCard: React.FC<GlassCardProps> = ({ children, className = '', glow = false }) => {
  return (
    <div className={`glass-card ${glow ? 'glass-card-glow' : ''} ${className}`}>
      {children}
    </div>
  );
};
