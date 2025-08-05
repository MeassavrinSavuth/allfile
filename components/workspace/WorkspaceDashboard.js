'use client';

import React from 'react';
import { FaBell } from 'react-icons/fa';
import WorkspaceCard from './WorkspaceCard';
import ConfirmModal from './ConfirmModal';
import InviteModal from './InviteModal';

export default function WorkspaceDashboard({
  workspaces,
  loading,
  error,
  currentUser,
  invitations,
  showCreateModal,
  setShowCreateModal,
  showInvitesModal,
  setShowInvitesModal,
  onEnterWorkspace,
  onDeleteWorkspace,
  onCreateWorkspace,
  deleteWorkspaceId,
  setDeleteWorkspaceId,
  confirmDeleteWorkspace,
  newWorkspaceName,
  setNewWorkspaceName,
  newWorkspaceAvatar,
  setNewWorkspaceAvatar,
  onAcceptInvitation,
  onDeclineInvitation,
  invitationsLoading
}) {
  return (
    <div className="min-h-screen bg-gray-50 p-8">
      <div className="flex items-center justify-between mb-8">
        <h1 className="text-3xl font-bold text-gray-800">Your Workspaces</h1>
        <div className="flex items-center gap-4">
          <button
            className="relative focus:outline-none"
            onClick={() => {
              console.log('Bell clicked!');
              setShowInvitesModal(true);
            }}
            aria-label="View Invitations"
          >
            <FaBell className="text-2xl text-gray-600 hover:text-blue-600 transition" />
            {Array.isArray(invitations) && invitations.length > 0 && (
              <span className="absolute -top-1 -right-1 bg-red-500 text-white text-xs rounded-full px-1.5 py-0.5 font-bold animate-pulse">
                {invitations.length}
              </span>
            )}
          </button>
          <button
            className="py-2 px-6 bg-blue-600 text-white rounded-lg font-semibold hover:bg-blue-700 transition"
            onClick={() => setShowCreateModal(true)}
          >
            + Create Workspace
          </button>
        </div>
      </div>

      {loading ? (
        <div className="text-center py-8">
          <div className="text-gray-500">Loading workspaces...</div>
        </div>
      ) : error ? (
        <div className="text-center py-8">
          <div className="text-red-500">Error loading workspaces: {error}</div>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-8">
          {Array.isArray(workspaces) && workspaces.map((ws) => (
            <WorkspaceCard
              key={ws.id}
              avatar={ws.avatar || 'https://randomuser.me/api/portraits/lego/1.jpg'}
              name={ws.name}
              admin={ws.admin_name}
              onClick={() => onEnterWorkspace(ws)}
              isAdmin={ws.admin_id === currentUser?.id}
              onDelete={() => onDeleteWorkspace(ws.id)}
            />
          ))}
        </div>
      )}

      {/* Delete Workspace Confirmation Modal */}
      <ConfirmModal
        isOpen={!!deleteWorkspaceId}
        title="Delete Workspace"
        message="Are you sure you want to delete this workspace? This action cannot be undone."
        onConfirm={confirmDeleteWorkspace}
        onCancel={() => setDeleteWorkspaceId(null)}
      />

      {/* Create Workspace Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black bg-opacity-30 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl shadow-lg p-8 w-full max-w-md">
            <h2 className="text-2xl font-bold mb-4 text-gray-800">Create Workspace</h2>
            <form onSubmit={onCreateWorkspace} className="space-y-4">
              <div>
                <label className="block text-gray-700 font-semibold mb-1">Workspace Name</label>
                <input
                  type="text"
                  className="w-full border rounded px-3 py-2 text-black"
                  value={newWorkspaceName}
                  onChange={e => setNewWorkspaceName(e.target.value)}
                  required
                />
              </div>
              <div>
                <label className="block text-blue-600 font-semibold mb-1">Avatar Image (optional)</label>
                <input
                  type="file"
                  accept="image/*"
                  className="w-full border rounded px-3 py-2 text-black"
                  onChange={async e => {
                    if (e.target.files && e.target.files[0]) {
                      const file = e.target.files[0];
                      // Show a loading state if desired
                      const { uploadToCloudinary } = await import('../../hooks/api/uploadToCloudinary');
                      const url = await uploadToCloudinary(file);
                      setNewWorkspaceAvatar(url);
                    }
                  }}
                />
                {newWorkspaceAvatar && (
                  <img src={newWorkspaceAvatar} alt="Avatar Preview" className="mt-2 w-16 h-16 rounded-full object-cover border" />
                )}
              </div>
              <div className="flex gap-3 pt-4">
                <button
                  type="submit"
                  className="flex-1 py-2 px-4 bg-blue-600 text-white rounded font-semibold hover:bg-blue-700 transition"
                >
                  Create Workspace
                </button>
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  className="flex-1 py-2 px-4 bg-gray-300 text-gray-700 rounded font-semibold hover:bg-gray-400 transition"
                >
                  Cancel
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Invitations Modal */}
      <InviteModal
        open={showInvitesModal}
        onClose={() => {
          console.log('InviteModal closed');
          setShowInvitesModal(false);
        }}
        invitations={invitations}
        onAccept={onAcceptInvitation}
        onDecline={onDeclineInvitation}
        loading={invitationsLoading}
      />
    </div>
  );
} 