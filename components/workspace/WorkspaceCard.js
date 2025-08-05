import React from 'react';
import { FaTrash } from 'react-icons/fa';

export default function WorkspaceCard({ avatar, name, admin, onClick, isAdmin, onDelete }) {
  return (
    <div
      className="bg-white rounded-2xl shadow-lg p-6 flex flex-col items-center border hover:shadow-xl transition cursor-pointer relative"
      onClick={onClick}
    >
      {isAdmin && (
        <button
          className="absolute top-2 right-2 p-2 bg-red-100 hover:bg-red-200 rounded-full text-red-600 hover:text-red-800 transition z-10"
          onClick={e => { e.stopPropagation(); onDelete && onDelete(); }}
          title="Delete Workspace"
        >
          <FaTrash />
        </button>
      )}
      <img
        src={avatar}
        alt={name}
        className="w-20 h-20 rounded-full border mb-4 object-cover bg-gray-100"
      />
      <div className="text-xl font-semibold text-gray-800 mb-1">{name}</div>
      <div className="text-gray-500 text-sm mb-2">Admin: <span className="font-medium">{admin}</span></div>
    </div>
  );
} 