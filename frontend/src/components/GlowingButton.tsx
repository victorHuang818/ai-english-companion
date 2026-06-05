import React from 'react';
import './GlowingButton.css';

interface GlowingButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'danger';
  fullWidth?: boolean;
}

export const GlowingButton: React.FC<GlowingButtonProps> = ({ 
  children, 
  variant = 'primary', 
  fullWidth = false,
  className = '',
  ...props 
}) => {
  return (
    <button 
      className={`glowing-button btn-${variant} ${fullWidth ? 'btn-full' : ''} ${className}`}
      {...props}
    >
      {children}
    </button>
  );
};
