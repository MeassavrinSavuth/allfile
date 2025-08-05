'use client';

import React from 'react';
import { FaUserPlus, FaArrowLeft } from 'react-icons/fa';
import InviteMemberModal from './InviteMemberModal';

export default function WorkspaceHeader({ 
  workspace, 
  currentUser, 
  onBack, 
  onInviteMember, 
  inviteLoading, 
  inviteError,
  showInviteMemberModal,
  setShowInviteMemberModal,
  onOpenInviteMemberModal
}) {
  return (
    <>
      <div className="flex items-center justify-between gap-4 mb-8">
        <div className="flex items-center gap-4">
          <button
            className="p-2 rounded-full bg-gray-200 hover:bg-gray-300 text-gray-700"
            onClick={onBack}
            aria-label="Back to Workspaces"
          >
            <FaArrowLeft />
          </button>
          <img 
            src={workspace.avatar} 
            alt={workspace.name} 
            className="w-14 h-14 rounded-full border object-cover bg-gray-100" 
          />
          <div>
            <div className="text-2xl font-bold text-gray-800">{workspace.name}</div>
            <div className="text-gray-500 text-sm">
              Admin: <span className="font-medium">{workspace.admin_name}</span>
            </div>
          </div>
        </div>
        
        {/* Only admin can invite */}
        {workspace.admin_id === currentUser.id && (
          <button
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded font-semibold hover:bg-blue-700 transition"
            onClick={onOpenInviteMemberModal || (() => setShowInviteMemberModal(true))}
          >
            <FaUserPlus /> Invite Member
          </button>
        )}
      </div>

      <InviteMemberModal
        open={showInviteMemberModal}
        onClose={() => setShowInviteMemberModal(false)}
        onInvite={onInviteMember}
        loading={inviteLoading}
        error={inviteError}
      />
    </>
  );
} 