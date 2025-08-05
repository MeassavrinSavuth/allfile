"use client";
import React from 'react';
import { useWorkspaceState } from '../../hooks/workspace/useWorkspaceState';
import WorkspaceDashboard from '../../components/workspace/WorkspaceDashboard';
import WorkspaceHeader from '../../components/workspace/WorkspaceHeader';
import WorkspaceContent from '../../components/workspace/WorkspaceContent';
import MemberList from '../../components/workspace/MemberList';
import ConfirmModal from '../../components/workspace/ConfirmModal';
import InviteModal from '../../components/workspace/InviteModal';

export default function WorkspacePage() {
  const {
    // State
    workspaces,
    loading,
    error,
    currentUser,
    userLoading,
    invitations,
    invitationsLoading,
    selectedWorkspace,
    activeTab,
    members,
    membersLoading,
    membersError,
    
    // Modal states
    showCreateModal,
    setShowCreateModal,
    showInvitesModal,
    setShowInvitesModal,
    showInviteMemberModal,
    setShowInviteMemberModal,
    showLeaveModal,
    setShowLeaveModal,
    showMemberList,
    setShowMemberList,
    
    // Form states
    newWorkspaceName,
    setNewWorkspaceName,
    newWorkspaceAvatar,
    setNewWorkspaceAvatar,
    inviteError,
    setInviteError,
    
    // Loading states
    roleChangeLoading,
    deleteWorkspaceId,
    setDeleteWorkspaceId,
    deleteLoading,
    leaveLoading,
    
    // Handlers
    handleEnterWorkspace,
    handleBackToDashboard,
    handleCreateWorkspace,
    handleDeleteWorkspace,
    confirmDeleteWorkspace,
    handleInviteMember,
    handleOpenInviteMemberModal,
    handleRoleChange,
    handleLeaveWorkspace,
    confirmLeaveWorkspace,
    handleRemoveMember,
    setActiveTab,
    
    // API functions
    acceptInvitation,
    declineInvitation,
    fetchInvitations,
    showKickModal,
    kickMemberName,
    confirmKickMember,
    cancelKickMember,
  } = useWorkspaceState();

  // Show loading state
  if (userLoading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-gray-500">Loading...</div>
      </div>
    );
  }

  // Show dashboard view (no workspace selected)
  if (!selectedWorkspace) {
    return (
      <WorkspaceDashboard
        workspaces={workspaces}
        loading={loading}
        error={error}
        currentUser={currentUser}
        invitations={invitations}
        showCreateModal={showCreateModal}
        setShowCreateModal={setShowCreateModal}
        showInvitesModal={showInvitesModal}
        setShowInvitesModal={setShowInvitesModal}
        onEnterWorkspace={handleEnterWorkspace}
        onDeleteWorkspace={handleDeleteWorkspace}
        onCreateWorkspace={handleCreateWorkspace}
        deleteWorkspaceId={deleteWorkspaceId}
        setDeleteWorkspaceId={setDeleteWorkspaceId}
        confirmDeleteWorkspace={confirmDeleteWorkspace}
        newWorkspaceName={newWorkspaceName}
        setNewWorkspaceName={setNewWorkspaceName}
        newWorkspaceAvatar={newWorkspaceAvatar}
        setNewWorkspaceAvatar={setNewWorkspaceAvatar}
        onAcceptInvitation={acceptInvitation}
        onDeclineInvitation={declineInvitation}
        invitationsLoading={invitationsLoading}
        onRefreshInvitations={fetchInvitations}
      />
    );
  }

  // Show workspace view
  return (
    <div className="min-h-screen bg-gray-50 p-8">
      <WorkspaceHeader
        workspace={selectedWorkspace}
        currentUser={currentUser}
        onBack={handleBackToDashboard}
        onInviteMember={handleInviteMember}
        inviteLoading={invitationsLoading}
        inviteError={inviteError}
        showInviteMemberModal={showInviteMemberModal}
        setShowInviteMemberModal={setShowInviteMemberModal}
        onOpenInviteMemberModal={handleOpenInviteMemberModal}
      />

      <MemberList
        showMemberList={showMemberList}
        setShowMemberList={setShowMemberList}
        members={members}
        membersLoading={membersLoading}
        membersError={membersError}
        currentUser={currentUser}
        selectedWorkspace={selectedWorkspace}
        roleChangeLoading={roleChangeLoading}
        onRoleChange={handleRoleChange}
        onLeaveWorkspace={handleLeaveWorkspace}
        onRemoveMember={handleRemoveMember}
      />

      <WorkspaceContent
        activeTab={activeTab}
        onTabChange={setActiveTab}
        selectedWorkspace={selectedWorkspace}
        members={members}
        currentUser={currentUser}
      />

      {/* Leave Workspace Confirmation Modal */}
      <ConfirmModal
        isOpen={showLeaveModal}
        title="Leave Workspace"
        message="Are you sure you want to leave this workspace?"
        onConfirm={confirmLeaveWorkspace}
        onCancel={() => setShowLeaveModal(false)}
      />

      {/* Kick Member Confirmation Modal */}
      <ConfirmModal
        isOpen={showKickModal}
        title="Remove Member"
        message={`Are you sure you want to remove ${kickMemberName} from this workspace?`}
        onConfirm={confirmKickMember}
        onCancel={cancelKickMember}
      />

      {/* Invitations Modal */}
      <InviteModal
        isOpen={showInvitesModal}
        onClose={() => setShowInvitesModal(false)}
        invitations={invitations}
        onAccept={acceptInvitation}
        onDecline={declineInvitation}
        loading={invitationsLoading}
      />
    </div>
  );
}