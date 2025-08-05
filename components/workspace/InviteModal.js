import React, { useEffect } from 'react';

function getField(inv, camel, snake) {
  return inv[camel] !== undefined ? inv[camel] : inv[snake];
}

export default function InviteModal({ open, invitations, onAccept, onDecline, onClose, loading = false }) {
  useEffect(() => {
    console.log('InviteModal render - open:', open, 'invitations:', invitations);
  }, [open, invitations]);
  
  if (!open) return null;
  
  return (
    <div className="fixed inset-0 bg-black bg-opacity-30 flex items-center justify-center z-50">
      <div className="bg-white rounded-xl shadow-lg p-8 w-full max-w-md">
        <h2 className="text-2xl font-bold mb-4 text-gray-800">Invitations</h2>
        {loading ? (
          <div className="text-gray-500 text-center py-8">Loading invitations...</div>
        ) : !invitations || invitations.length === 0 ? (
          <div className="text-gray-500 text-center py-8">No pending invitations.</div>
        ) : (
          <ul className="space-y-4">
            {invitations.map(invitation => (
              <li key={getField(invitation, 'id', 'id')} className="flex items-center gap-4 bg-gray-50 rounded-lg p-4">
                <div className="w-12 h-12 rounded-full bg-blue-100 flex items-center justify-center">
                  <span className="text-blue-600 font-semibold text-lg">
                    {getField(invitation, 'workspaceName', 'workspace_name')?.charAt(0) || 'W'}
                  </span>
                </div>
                <div className="flex-1">
                  <div className="font-semibold text-gray-800">{getField(invitation, 'workspaceName', 'workspace_name')}</div>
                  <div className="text-gray-500 text-sm">
                    Invited by {getField(invitation, 'inviterName', 'inviter_name')} as {getField(invitation, 'role', 'role')}
                  </div>
                  <div className="text-xs text-gray-400 mt-1">
                    Expires {new Date(getField(invitation, 'expiresAt', 'expires_at')).toLocaleDateString()}
                  </div>
                </div>
                <div className="flex flex-col gap-2">
                  <button
                    className="px-3 py-1 rounded bg-blue-600 text-white text-sm font-semibold hover:bg-blue-700 disabled:opacity-50"
                    onClick={() => onAccept(getField(invitation, 'id', 'id'))}
                    disabled={loading}
                  >
                    Accept
                  </button>
                  <button
                    className="px-3 py-1 rounded bg-gray-200 text-gray-700 text-sm font-semibold hover:bg-gray-300 disabled:opacity-50"
                    onClick={() => onDecline(getField(invitation, 'id', 'id'))}
                    disabled={loading}
                  >
                    Decline
                  </button>
                </div>
              </li>
            ))}
          </ul>
        )}
        <div className="flex justify-end mt-6">
          <button
            className="px-4 py-2 rounded bg-gray-200 text-gray-700 font-semibold hover:bg-gray-300"
            onClick={onClose}
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
} 