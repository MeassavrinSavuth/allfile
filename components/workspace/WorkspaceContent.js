'use client';

import React from 'react';
import TasksSection from './TasksSection';
import DraftsSection from './DraftsSection';
import MediaLibraryGrid from './MediaLibraryGrid';
import WorkspaceTabs from './WorkspaceTabs';

export default function WorkspaceContent({
  activeTab,
  onTabChange,
  selectedWorkspace,
  members,
  currentUser
}) {
  const renderContent = () => {
    switch (activeTab) {
      case 'tasks':
        return (
          <TasksSection 
            workspaceId={selectedWorkspace?.id} 
            teamMembers={members} 
            currentUser={currentUser} 
          />
        );
      case 'drafts':
        return selectedWorkspace ? (
          <DraftsSection
            teamMembers={members}
            currentUser={currentUser}
            workspaceId={selectedWorkspace.id}
          />
        ) : null;
      case 'media':
        return <MediaLibraryGrid workspaceId={selectedWorkspace.id} />;
      default:
        return null;
    }
  };

  return (
    <div>
      <WorkspaceTabs activeTab={activeTab} onTabChange={onTabChange} />
      {renderContent()}
    </div>
  );
} 