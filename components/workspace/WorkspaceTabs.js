'use client';

import React from 'react';

export default function WorkspaceTabs({ activeTab, onTabChange }) {
  const tabs = [
    { id: 'tasks', label: 'Tasks' },
    { id: 'drafts', label: 'Drafts' },
    { id: 'media', label: 'Media' }
  ];

  return (
    <div className="flex gap-2 mb-6">
      {tabs.map((tab) => (
        <button
          key={tab.id}
          className={`px-4 py-2 rounded-t font-semibold border-b-2 transition-all ${
            activeTab === tab.id 
              ? 'border-blue-600 text-blue-700 bg-white' 
              : 'border-transparent text-gray-500 bg-gray-100 hover:text-blue-600'
          }`}
          onClick={() => onTabChange(tab.id)}
        >
          {tab.label}
        </button>
      ))}
    </div>
  );
} 