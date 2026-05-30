'use client';
import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from './context/AuthContext';

export default function Home() {
  const { isAuthenticated, loading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!loading) {
      router.replace(isAuthenticated ? '/dashboard' : '/login');
    }
  }, [loading, isAuthenticated, router]);

  return null;
}
