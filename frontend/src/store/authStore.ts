import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface UserInfo {
  id: string;
  username: string;
  email: string;
  daily_free_tokens: number;
  recharge_tokens: number;
}

interface AuthState {
  token: string | null;
  user: UserInfo | null;
  setAuth: (token: string, user: UserInfo) => void;
  updateUser: (user: UserInfo) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      user: null,
      setAuth: (token, user) => set({ token, user }),
      updateUser: (user) => set({ user }),
      logout: () => set({ token: null, user: null }),
    }),
    {
      name: 'auth-storage',
    }
  )
)
