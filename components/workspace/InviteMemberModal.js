import React, { useState } from 'react';

export default function InviteMemberModal({ open, onClose, onInvite, loading = false, error = null }) {
  const [email, setEmail] = useState('');
  const [role, setRole] = useState('Editor');

  const handleSubmit = (e) => {
    e.preventDefault();
    if (email.trim()) {
      onInvite(email, role);
    }
  };

  const handleClose = () => {
    setEmail('');
    setRole('Editor');
    onClose();
  };

  if (!open) return null;
  
  return (
    <div className="fixed inset-0 bg-black bg-opacity-30 flex items-center justify-center z-50">
      <div className="bg-white rounded-xl shadow-lg p-8 w-full max-w-md">
        <h2 className="text-2xl font-bold mb-4 text-gray-800">Invite Member</h2>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-gray-700 font-semibold mb-1">Email Address</label>
            <input
              type="email"
              className="w-full border rounded px-3 py-2"
              value={email}
              onChange={e => setEmail(e.target.value)}
              placeholder="Enter email address"
              required
              disabled={loading}
            />
          </div>
          <div>
            <label className="block text-gray-700 font-semibold mb-1">Role</label>
            <select
              className={`w-full border rounded px-3 py-2 transition-colors duration-200
                ${role === 'Admin' ? 'text-blue-700 bg-blue-50' : role === 'Editor' ? 'text-green-700 bg-green-50' : 'text-gray-700 bg-gray-50'}`}
              value={role}
              onChange={e => setRole(e.target.value)}
              disabled={loading}
            >
              <option value="Editor">Editor</option>
              <option value="Viewer">Viewer</option>
              <option value="Admin">Admin</option>
            </select>
            <p className="text-xs text-gray-500 mt-1">
              {role === 'Admin' && 'Can manage workspace and members'}
              {role === 'Editor' && 'Can create and edit content'}
              {role === 'Viewer' && 'Can view content only'}
            </p>
          </div>
          {error && (
            <div className="text-red-500 text-sm bg-red-50 p-3 rounded border border-red-200">
              <div className="font-semibold mb-1">Invitation Error:</div>
              {error}
            </div>
          )}
          <div className="flex justify-end gap-2 mt-6">
            <button
              type="button"
              className="px-4 py-2 rounded bg-gray-200 text-gray-700 font-semibold hover:bg-gray-300"
              onClick={handleClose}
              disabled={loading}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="px-4 py-2 rounded bg-blue-600 text-white font-semibold hover:bg-blue-700 disabled:opacity-50"
              disabled={loading || !email.trim()}
            >
              {loading ? 'Sending...' : 'Invite'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
} 