'use client';

import { useEffect, useState } from 'react';
import { fetchAPI, postAPI, putAPI, deleteAPI } from '../lib/api';
import { LoadingState, EmptyState } from '../components/UI';
import { useNotification } from '../components/Notification';

export default function Team() {
  const [members, setMembers] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editMember, setEditMember] = useState<any>(null);
  const { addNotification } = useNotification();

  const loadMembers = () => {
    fetchAPI('/team')
      .then(d => setMembers(d.members || []))
      .catch(() => {})
      .finally(() => setLoading(false));
  };

  useEffect(() => { loadMembers(); }, []);

  const handleDelete = async (id: string) => {
    if (!confirm('确定删除此成员？')) return;
    try {
      await deleteAPI(`/team/${id}`);
      setMembers(members.filter(m => m.id !== id));
      addNotification('success', '删除成功', '成员已删除');
    } catch (e: any) {
      addNotification('error', '删除失败', e.message);
    }
  };

  const handleSave = async (form: any) => {
    try {
      if (editMember) {
        await putAPI(`/team/${editMember.id}`, form);
        addNotification('success', '更新成功', '成员信息已更新');
      } else {
        await postAPI('/team', form);
        addNotification('success', '添加成功', '新成员已添加');
      }
      setShowForm(false);
      setEditMember(null);
      loadMembers();
    } catch (e: any) {
      addNotification('error', '保存失败', e.message);
    }
  };

  if (loading) return <LoadingState text="加载团队成员…" />;

  return (
    <>
      <div className="flex items-center justify-between">
        <div><h1 className="text-xl font-semibold text-white">团队管理</h1><p className="text-sm text-zinc-500 mt-0.5">管理团队成员与权限</p></div>
        <button onClick={() => { setEditMember(null); setShowForm(true); }}
          className="text-xs px-3 py-1.5 rounded-lg bg-accent/10 text-accent hover:bg-accent/20 transition-colors">+ 添加成员</button>
      </div>

      {members.length === 0 ? <EmptyState title="暂无成员" description="点击上方按钮添加团队成员" /> : (
        <div className="bg-surface-50 border border-white/5 rounded-xl overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead><tr className="text-left border-b border-white/5">
                {['成员', '邮箱', '角色', '团队', '操作'].map(h => (
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
                      <span className={`font-mono text-[11px] px-2 py-0.5 rounded-full ${m.role === 'admin' ? 'bg-accent/10 text-accent' : m.role === 'editor' ? 'bg-[rgba(245,158,11,0.1)] text-[#fbbf24]' : 'bg-surface-300 text-zinc-400'}`}>
                        {m.role === 'admin' ? '管理员' : m.role === 'editor' ? '编辑' : '查看者'}
                      </span>
                    </td>
                    <td className="px-4 py-3 font-mono text-xs text-zinc-400">{m.team}</td>
                    <td className="px-4 py-3">
                      <div className="flex gap-2">
                        <button onClick={() => { setEditMember(m); setShowForm(true); }} className="text-xs text-zinc-500 hover:text-accent">编辑</button>
                        <button onClick={() => handleDelete(m.id)} className="text-xs text-zinc-500 hover:text-danger">删除</button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {showForm && <MemberForm member={editMember} onSave={handleSave} onClose={() => { setShowForm(false); setEditMember(null); }} />}
    </>
  );
}

function MemberForm({ member, onSave, onClose }: { member: any; onSave: (f: any) => void; onClose: () => void }) {
  const [form, setForm] = useState({
    name: member?.name || '',
    email: member?.email || '',
    role: member?.role || 'viewer',
    team: member?.team || '',
  });

  return (
    <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center" onClick={onClose}>
      <div className="bg-surface-50 border border-white/10 rounded-xl p-6 w-full max-w-md" onClick={e => e.stopPropagation()}>
        <h3 className="text-sm font-medium text-zinc-200 mb-4">{member ? '编辑成员' : '添加成员'}</h3>
        <div className="space-y-3">
          <div>
              <label htmlFor="member-name" className="font-mono text-[10px] text-zinc-500 mb-1 block">姓名</label>
              <input id="member-name" value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} placeholder="张三…"
                className="w-full bg-surface-100 border border-white/5 text-sm text-zinc-300 placeholder-zinc-600 rounded-lg px-3 py-2 outline-none focus:border-accent/30" />
            </div>
            <div>
              <label htmlFor="member-email" className="font-mono text-[10px] text-zinc-500 mb-1 block">邮箱</label>
              <input id="member-email" type="email" autoComplete="email" value={form.email} onChange={e => setForm({ ...form, email: e.target.value })} placeholder="zhangsan@example.com" spellCheck={false}
                className="w-full bg-surface-100 border border-white/5 text-sm text-zinc-300 placeholder-zinc-600 rounded-lg px-3 py-2 outline-none focus:border-accent/30" />
            </div>
            <div>
              <label htmlFor="member-role" className="font-mono text-[10px] text-zinc-500 mb-1 block">角色</label>
              <select id="member-role" value={form.role} onChange={e => setForm({ ...form, role: e.target.value })}
                className="w-full bg-surface-100 border border-white/5 text-sm text-zinc-300 rounded-lg px-3 py-2 outline-none focus:border-accent/30">
                <option value="admin">管理员</option>
                <option value="editor">编辑</option>
                <option value="viewer">查看者</option>
              </select>
            </div>
            <div>
              <label htmlFor="member-team" className="font-mono text-[10px] text-zinc-500 mb-1 block">团队</label>
              <input id="member-team" value={form.team} onChange={e => setForm({ ...form, team: e.target.value })} placeholder="平台运维…"
                className="w-full bg-surface-100 border border-white/5 text-sm text-zinc-300 placeholder-zinc-600 rounded-lg px-3 py-2 outline-none focus:border-accent/30" />
            </div>
        </div>
        <div className="flex justify-end gap-2 mt-5">
          <button onClick={onClose} className="text-xs px-3 py-1.5 rounded-lg text-zinc-400 hover:text-zinc-200">取消</button>
          <button onClick={() => onSave(form)} className="text-xs px-3 py-1.5 rounded-lg bg-accent/10 text-accent hover:bg-accent/20">保存</button>
        </div>
      </div>
    </div>
  );
}
