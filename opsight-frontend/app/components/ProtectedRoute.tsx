'use client';
import { ReactNode } from 'react';
import { redirect } from 'next/navigation';
import { useAuth } from '../context/AuthContext';
import { LoadingSpinner } from './UI';

export default function ProtectedRoute({ children }: { children: ReactNode }) {
  const { isAuthenticated, loading } = useAuth();

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <LoadingSpinner size="lg" />
      </div>
    );
  }

  if (!isAuthenticated) {
    redirect('/login');
  }

  return <>{children}</>;
}
