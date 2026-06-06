import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { GlassCard } from '../components/GlassCard';
import { GlowingButton } from '../components/GlowingButton';
import { api } from '../services/api';
import { useAuthStore } from '../store/authStore';
import './AuthPage.css';

export const AuthPage: React.FC = () => {
  const [isLogin, setIsLogin] = useState(true);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [username, setUsername] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  
  const navigate = useNavigate();
  const { setAuth } = useAuthStore();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      if (isLogin) {
        const res = await api.login({ email, password });
        if (res.token) {
          useAuthStore.setState({ token: res.token });
          const userInfo = await api.getUserInfo();
          setAuth(res.token, userInfo);
          navigate('/');
        }
      } else {
        await api.register({ username, email, password });
        // Automatically login after register
        const res = await api.login({ email, password });
        if (res.token) {
          // We need to set the token manually in store before fetching user info
          // Actually api.ts uses getState().token, so we should set it first
          useAuthStore.setState({ token: res.token });
          const userInfo = await api.getUserInfo();
          setAuth(res.token, userInfo);
          navigate('/');
        }
      }
    } catch (err: any) {
      setError(err.message || 'Authentication failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="auth-container animate-fade-in">
      <div className="auth-background-decoration"></div>
      <GlassCard className="auth-card" glow>
        <div className="auth-header">
          <h1>AI English Companion</h1>
          <p>{isLogin ? 'Welcome back to the future' : 'Start your journey'}</p>
        </div>

        {error && <div className="auth-error">{error}</div>}

        <form onSubmit={handleSubmit} className="auth-form">
          {!isLogin && (
            <div className="form-group">
              <label>Username</label>
              <input 
                type="text" 
                value={username} 
                onChange={e => setUsername(e.target.value)}
                required={!isLogin}
                placeholder="CyberNinja"
              />
            </div>
          )}
          
          <div className="form-group">
            <label>Email</label>
            <input 
              type="email" 
              value={email} 
              onChange={e => setEmail(e.target.value)}
              required
              placeholder="neo@matrix.com"
            />
          </div>

          <div className="form-group">
            <label>Password</label>
            <input 
              type="password" 
              value={password} 
              onChange={e => setPassword(e.target.value)}
              required
              placeholder="••••••••"
            />
          </div>

          <GlowingButton type="submit" fullWidth disabled={loading}>
            {loading ? 'Processing...' : (isLogin ? 'Initialize Sequence' : 'Create Identity')}
          </GlowingButton>
        </form>

        <div className="auth-toggle">
          <span>{isLogin ? "Don't have an identity?" : "Already initialized?"}</span>
          <button type="button" onClick={() => setIsLogin(!isLogin)} className="toggle-btn">
            {isLogin ? 'Register' : 'Login'}
          </button>
        </div>
      </GlassCard>
    </div>
  );
};
