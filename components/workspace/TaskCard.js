import { useState } from 'react';
import { useToggle } from '../../hooks/ui/useToggle';
import CommentSection from './CommentSection';
import { useUser } from '../../hooks/auth/useUser';

const TaskCard = ({ task, onUpdate, onDelete, workspaceId }) => {
  const [isEditing, setIsEditing] = useState(false);
  const [showComments, toggleComments] = useToggle(false);
  const [menuOpen, setMenuOpen] = useState(false);
  const { profileData: currentUser } = useUser();
  const [editForm, setEditForm] = useState({
    title: task.title,
    description: task.description || '',
    status: task.status,
    assigned_to: task.assigned_to || '',
    due_date: task.due_date ? new Date(task.due_date).toISOString().split('T')[0] : '',
  });

  const statusColors = {
    'Todo': 'bg-gray-100 text-gray-800',
    'In Progress': 'bg-blue-100 text-blue-800',
    'Review': 'bg-yellow-100 text-yellow-800',
    'Done': 'bg-green-100 text-green-800',
  };

  const handleSave = async () => {
    const updates = {};
    Object.keys(editForm).forEach(key => {
      if (editForm[key] !== task[key]) {
        updates[key] = editForm[key] || null;
      }
    });

    if (Object.keys(updates).length > 0) {
      const success = await onUpdate(task.id, updates);
      if (success) {
        setIsEditing(false);
      }
    }
  };

  const handleDelete = async () => {
    if (confirm('Are you sure you want to delete this task?')) {
      await onDelete(task.id);
    }
  };

  const formatDate = (dateString) => {
    if (!dateString) return 'No due date';
    return new Date(dateString).toLocaleDateString();
  };

  const isOverdue = task.due_date && new Date(task.due_date) < new Date() && task.status !== 'Done';

  // Show the 3-dot menu for all users in the workspace
  const isOwner = true;

  return (
    <div className="bg-white rounded-lg shadow-md border border-gray-200 p-6 mb-4">
      {isEditing ? (
        <div className="space-y-4">
          <input
            type="text"
            value={editForm.title}
            onChange={(e) => setEditForm({ ...editForm, title: e.target.value })}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="Task title"
          />
          <textarea
            value={editForm.description}
            onChange={(e) => setEditForm({ ...editForm, description: e.target.value })}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="Task description"
            rows="3"
          />
          <div className="grid grid-cols-2 gap-4">
            <select
              value={editForm.status}
              onChange={(e) => setEditForm({ ...editForm, status: e.target.value })}
              className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="Todo">Todo</option>
              <option value="In Progress">In Progress</option>
              <option value="Review">Review</option>
              <option value="Done">Done</option>
            </select>
            <input
              type="date"
              value={editForm.due_date}
              onChange={(e) => setEditForm({ ...editForm, due_date: e.target.value })}
              className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <input
            type="text"
            value={editForm.assigned_to}
            onChange={(e) => setEditForm({ ...editForm, assigned_to: e.target.value })}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="Assigned to (email)"
          />
          <div className="flex space-x-2">
            <button
              onClick={handleSave}
              className="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              Save
            </button>
            <button
              onClick={() => setIsEditing(false)}
              className="px-4 py-2 bg-gray-500 text-white rounded-md hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-gray-500"
            >
              Cancel
            </button>
          </div>
        </div>
      ) : (
        <div>
          <div className="flex items-center justify-between mb-4">
            {/* Creator info and 3-dot menu */}
            <div className="flex items-center">
              <img
                src={task.creator_avatar || '/default-avatar.png'}
                alt={task.creator_name || 'Unknown'}
                className="w-9 h-9 rounded-full border mr-3 shadow-sm"
                title={task.creator_name || 'Unknown'}
              />
              {/* 3-dot menu for all users, right next to avatar */}
              <div className="relative ml-3 flex items-center">
                <button
                  onClick={() => setMenuOpen((open) => !open)}
                  className="p-2 text-gray-500 hover:text-gray-700 focus:outline-none"
                  aria-label="Task options"
                >
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <circle cx="12" cy="6" r="1.5"/>
                    <circle cx="12" cy="12" r="1.5"/>
                    <circle cx="12" cy="18" r="1.5"/>
                  </svg>
                </button>
                {menuOpen && (
                  <div className="absolute right-0 mt-2 w-32 bg-white border rounded shadow z-10">
                    <button
                      className="block w-full text-left px-4 py-2 text-blue-700 hover:bg-blue-100 font-semibold"
                      onClick={() => { setIsEditing(true); setMenuOpen(false); }}
                    >
                      Edit
                    </button>
                    <button
                      className="block w-full text-left px-4 py-2 text-red-600 hover:bg-gray-100 font-semibold"
                      onClick={() => { setMenuOpen(false); handleDelete(); }}
                    >
                      Delete
                    </button>
                  </div>
                )}
              </div>
              <div className="flex flex-col ml-3">
                <span className="font-semibold text-base text-gray-900 leading-tight">
                  {task.creator_name || 'Unknown'}
                </span>
                <span className="text-xs text-gray-500">{task.created_at ? new Date(task.created_at).toLocaleString() : 'Just now'}</span>
              </div>
              <span className="ml-2 text-xs text-gray-400">(Posted by)</span>
            </div>
            {/* Status and Assignee with headers (unchanged) */}
            <div className="flex flex-wrap gap-2 mb-4">
              <span className={`px-2 py-1 rounded-full text-xs font-medium ${statusColors[task.status] || statusColors['Todo']}`}>
                {task.status}
              </span>
              {task.assigned_to && (
                <span className="px-2 py-1 bg-purple-100 text-purple-800 rounded-full text-xs font-medium">
                  Assigned to: {task.assigned_to}
                </span>
              )}
              <span className={`px-2 py-1 rounded-full text-xs font-medium ${
                isOverdue ? 'bg-red-100 text-red-800' : 'bg-gray-100 text-gray-800'
              }`}>
                Due: {formatDate(task.due_date)}
              </span>
            </div>
          </div>
          <div className="flex justify-between items-center">
            <button
              onClick={toggleComments}
              className="text-sm text-blue-500 hover:text-blue-700 focus:outline-none"
            >
              {showComments ? 'Hide Comments' : 'Show Comments'}
            </button>
            <div className="text-xs text-gray-500">
              Created {new Date(task.created_at).toLocaleDateString()}
            </div>
          </div>
          {showComments && (
            <div className="mt-4 pt-4 border-t border-gray-200">
              <CommentSection taskId={task.id} workspaceId={workspaceId} />
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default TaskCard;