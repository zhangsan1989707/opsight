'use client';
import { createContext, useContext, useState, useEffect, useCallback, ReactNode } from 'react';

interface User {
  id: number;
  name: string;
  email: string;
  role: string;
}

interface AuthContextType {
  user: User | null;
  token: string | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  isAuthenticated: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const clearAuth = useCallback(() => {
    localStorage.removeItem('opsight_token');
    localStorage.removeItem('opsight_user');
    setToken(null);
    setUser(null);
  }, []);

  const logout = useCallback(() => {
    clearAuth();
  }, [clearAuth]);

  // On mount, check localStorage for existing token and validate it
  useEffect(() => {
    const storedToken = localStorage.getItem('opsight_token');
    const storedUser = localStorage.getItem('opsight_user');
    if (storedToken && storedUser) {
      setToken(storedToken);
      setUser(JSON.parse(storedUser));
      // Verify token is still valid
      fetch('/api/v1/auth/me', {
        headers: { 'Authorization': `Bearer ${storedToken}` }
      }).then(res => {
        if (!res.ok) { clearAuth(); }
      }).catch(() => clearAuth());
    }
    setLoading(false);
  }, [clearAuth]);

  const login = useCallback(async (email: string, password: string) => {
    const res = await fetch('/api/v1/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password })
    });
    const json = await res.json();
    if (!res.ok || json.code !== 200) {
      throw new Error(json.message || 'Login failed');
    }
    const { token: newToken, user: newUser } = json.data;
    localStorage.setItem('opsight_token', newToken);
    localStorage.setItem('opsight_user', JSON.stringify(newUser));
    setToken(newToken);
    setUser(newUser);
  }, []);

  return (
    <AuthContext.Provider value={{ user, token, loading, login, logout, isAuthenticated: !!token }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
