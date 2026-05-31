'use client';
import { createContext, useContext, useState, useEffect, useRef, ReactNode } from 'react';
import { useAuth } from './AuthContext';

interface WSContextType {
  connected: boolean;
  lastEvent: { type: string; data: any; time: string } | null;
}

const WSContext = createContext<WSContextType>({ connected: false, lastEvent: null });

export function WSProvider({ children }: { children: ReactNode }) {
  const { token, isAuthenticated } = useAuth();
  const [connected, setConnected] = useState(false);
  const [lastEvent, setLastEvent] = useState<WSContextType['lastEvent']>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const retryRef = useRef<ReturnType<typeof setTimeout>>();

  useEffect(() => {
    if (!isAuthenticated || !token) {
      wsRef.current?.close();
      setConnected(false);
      return;
    }

    function connect() {
      const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const ws = new WebSocket(`${proto}//${window.location.host}/api/v1/ws?token=${token}`);
      wsRef.current = ws;

      ws.onopen = () => setConnected(true);
      ws.onclose = () => {
        setConnected(false);
        // Auto-reconnect after 5s
        retryRef.current = setTimeout(connect, 5000);
      };
      ws.onerror = () => {
        ws.close();
      };

      ws.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data);
          setLastEvent(msg);
        } catch {}
      };
    }

    connect();

    return () => {
      clearTimeout(retryRef.current);
      wsRef.current?.close();
    };
  }, [isAuthenticated, token]);

  return (
    <WSContext.Provider value={{ connected, lastEvent }}>
      {children}
    </WSContext.Provider>
  );
}

export function useWS() {
  return useContext(WSContext);
}
