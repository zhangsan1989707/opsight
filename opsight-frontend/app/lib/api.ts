const API_BASE = typeof window !== 'undefined'
  ? `${window.location.protocol}//${window.location.hostname}:8800/api/v1`
  : 'http://localhost:8800/api/v1';

export async function fetchAPI<T = any>(path: string): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, { cache: 'no-store' });
  if (!res.ok) throw new Error(`API error: ${res.status}`);
  return res.json();
}

export async function postAPI<T = any>(path: string, body?: any): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: body ? JSON.stringify(body) : undefined,
  });
  if (!res.ok) throw new Error(`API error: ${res.status}`);
  return res.json();
}

export async function patchAPI<T = any>(path: string): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, { method: 'PATCH' });
  if (!res.ok) throw new Error(`API error: ${res.status}`);
  return res.json();
}
