// components/PostQueue.js
'use client';

import React from 'react';

export default function PostQueue({ postQueue = [] }) {
  const getStatusColor = (status) => {
    switch (status) {
      case 'pending':
        return 'text-yellow-600 bg-yellow-100';
      case 'completed':
        return 'text-green-600 bg-green-100';
      case 'failed':
        return 'text-red-600 bg-red-100';
      default:
        return 'text-gray-600 bg-gray-100';
    }
  };

  return (
    <div className="space-y-4">
      {postQueue.length === 0 ? (
        <p className="text-gray-500 italic">No posts in the queue yet.</p>
      ) : (
        <ul className="space-y-3">
          {postQueue.map((item) => (
            <li key={item.id} className="bg-gray-50 p-3 rounded-lg shadow-sm border border-gray-200">
              <div className="flex justify-between items-start mb-1">
                <span className="font-medium text-gray-800 capitalize">{item.platform}</span>
                <span className={`text-xs font-semibold px-2 py-0.5 rounded-full ${getStatusColor(item.status)} capitalize`}>
                  {item.status}
                </span>
              </div>
              <p className="text-sm text-gray-700 truncate mb-1">{item.messageSnippet}</p>
              {item.mediaCount > 0 && (
                <p className="text-xs text-gray-500">{item.mediaCount} media file(s)</p>
              )}
              {item.status === 'failed' && item.error && (
                <p className="text-xs text-red-500 mt-1">Error: {item.error}</p>
              )}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}