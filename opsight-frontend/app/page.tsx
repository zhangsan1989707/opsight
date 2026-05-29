'use client';
import { useAuth } from './context/AuthContext';
import { redirect } from 'next/navigation';

export default function Home() {
  const { isAuthenticated, loading } = useAuth();
  if (loading) return null;
  redirect(isAuthenticated ? '/dashboard' : '/login');
}
