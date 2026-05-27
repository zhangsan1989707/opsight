'use client';

import { useEffect, useState } from 'react';
import { fetchAPI } from '../lib/api';

export default function Team() {
  const [members, setMembers] = useState<any[]>([]);

  useEffect(() => { fetchAPI('/team').then(d => setMembers(d.members || [])).catch(console.error); }, []);

  return (
    <>
      <div><h1 className="text-xl font-semibold text-white">Team</h1><p className="text-sm text-zinc-500 mt-0.5">Manage team members and permissions</p></div>

      <div className="bg-surface-50 border border-white/5 rounded-xl overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead><tr className="text-left border-b border-white/5">
              {['Member', 'Email', 'Role', 'Team'].map(h => (
                <th key={h} className="font-mono text-[10px] uppercase tracking-wider text-zinc-600 px-4 py-2.5">{h}</th>
              ))}
            </tr></thead>
            <tbody>
              {members.map(m => (
                <tr key={m.id} className="border-b border-white/[0.03] hover:bg-white/[0.02] transition-colors">
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-3">
                      <div className="w-8 h-8 rounded-full bg-accent/15 flex items-center justify-center text-accent text-xs font-semibold">
                        {m.name.split(' ').map((n: string) => n[0]).join('').slice(0, 2)}
                      </div>
                      <span className="text-zinc-300">{m.name}</span>
                    </div>
                  </td>
                  <td className="px-4 py-3 font-mono text-xs text-zinc-400">{m.email}</td>
                  <td className="px-4 py-3">
                    <span className={`font-mono text-[11px] px-2 py-0.5 rounded-full ${m.role === 'Admin' ? 'bg-accent/10 text-accent' : m.role === 'Editor' ? 'bg-[rgba(245,158,11,0.1)] text-[#fbbf24]' : 'bg-surface-300 text-zinc-400'}`}>{m.role}</span>
                  </td>
                  <td className="px-4 py-3 font-mono text-xs text-zinc-400">{m.team}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
}
