'use client';

import { createContext, useContext, useState, useCallback, ReactNode } from 'react';

interface Notification {
  id: string;
  type: 'info' | 'success' | 'warning' | 'error';
  title: string;
  message: string;
  timestamp: number;
}

interface NotificationContextType {
  notifications: Notification[];
  addNotification: (type: Notification['type'], title: string, message: string) => void;
  removeNotification: (id: string) => void;
}

const NotificationContext = createContext<NotificationContextType | undefined>(undefined);

export function useNotification() {
  const context = useContext(NotificationContext);
  if (!context) {
    throw new Error('useNotification must be used within NotificationProvider');
  }
  return context;
}

export function NotificationProvider({ children }: { children: ReactNode }) {
  const [notifications, setNotifications] = useState<Notification[]>([]);

  const addNotification = useCallback((type: Notification['type'], title: string, message: string) => {
    const id = `notif_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    setNotifications(prev => [{ id, type, title, message, timestamp: Date.now() }, ...prev].slice(0, 10));
  }, []);

  const removeNotification = useCallback((id: string) => {
    setNotifications(prev => prev.filter(n => n.id !== id));
  }, []);

  return (
    <NotificationContext.Provider value={{ notifications, addNotification, removeNotification }}>
      {children}
      <NotificationContainer />
    </NotificationContext.Provider>
  );
}

function NotificationContainer() {
  const { notifications, removeNotification } = useNotification();

  if (notifications.length === 0) return null;

  const typeStyles: Record<string, { border: string; icon: string; bg: string }> = {
    info: { border: 'border-[#0ea5e9]/30', icon: '#0ea5e9', bg: 'bg-[rgba(14,165,233,0.1)]' },
    success: { border: 'border-[#10b981]/30', icon: '#10b981', bg: 'bg-[rgba(16,185,129,0.1)]' },
    warning: { border: 'border-[#f59e0b]/30', icon: '#f59e0b', bg: 'bg-[rgba(245,158,11,0.1)]' },
    error: { border: 'border-[#ef4444]/30', icon: '#ef4444', bg: 'bg-[rgba(239,68,68,0.1)]' },
  };

  return (
    <div className="fixed top-4 right-4 z-50 space-y-2 max-w-sm w-full pointer-events-none">
      {notifications.map(notif => {
        const style = typeStyles[notif.type];
        return (
          <div
            key={notif.id}
            className={`bg-surface-50 border ${style.border} rounded-lg p-3 pointer-events-auto animate-slide-in`}
          >
            <div className="flex items-start gap-2">
              <div className={`w-6 h-6 rounded-full ${style.bg} flex items-center justify-center flex-shrink-0`}>
                <NotificationIcon type={notif.type} color={style.icon} />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-xs font-medium text-zinc-200">{notif.title}</p>
                <p className="text-[11px] text-zinc-500 mt-0.5">{notif.message}</p>
              </div>
              <button
                onClick={() => removeNotification(notif.id)}
                className="text-zinc-600 hover:text-zinc-400 transition-colors flex-shrink-0"
              >
                <svg width="12" height="12" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 12 12">
                  <path d="M2 2l8 8M10 2l-8 8" />
                </svg>
              </button>
            </div>
          </div>
        );
      })}
    </div>
  );
}

function NotificationIcon({ type, color }: { type: string; color: string }) {
  const icons: Record<string, JSX.Element> = {
    info: <svg width="12" height="12" fill="none" stroke={color} strokeWidth="2" viewBox="0 0 12 12"><circle cx="6" cy="6" r="4" /><path d="M6 5v2M6 3.5v.5" /></svg>,
    success: <svg width="12" height="12" fill="none" stroke={color} strokeWidth="2" viewBox="0 0 12 12"><path d="M2 6l3 3 5-5" /></svg>,
    warning: <svg width="12" height="12" fill="none" stroke={color} strokeWidth="2" viewBox="0 0 12 12"><path d="M6 2L1 10h10L6 2z" /><path d="M6 5v2M6 8.5v.5" /></svg>,
    error: <svg width="12" height="12" fill="none" stroke={color} strokeWidth="2" viewBox="0 0 12 12"><circle cx="6" cy="6" r="4" /><path d="M4.5 4.5l3 3M7.5 4.5l-3 3" /></svg>,
  };
  return icons[type] || icons.info;
}
