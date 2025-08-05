import React, { useState, useRef, useEffect } from 'react';
import { useTasks } from '../../hooks/api/useTasks';
import { useTaskReactions } from '../../hooks/api/useTaskReactions';
import TaskCard from './TaskCard';
import TaskForm from './TaskForm';
import Modal from './Modal';
import PlatformSelector from './PlatformSelector';
import MiniFacebookPreview from './MiniFacebookPreview';
import MiniMastodonPreview from './MiniMastodonPreview';
import CommentSection from './CommentSection';
import { FaUserPlus, FaTrash, FaEdit, FaClock, FaPlus, FaEllipsisH } from 'react-icons/fa';
import { MdOutlineModeComment, MdThumbUpOffAlt, MdFavoriteBorder } from 'react-icons/md';

const REACTION_EMOJIS = ['👍', '❤️', '😂', '🎉', '👎'];

export default function TasksSection({ workspaceId, teamMembers, currentUser }) {
  const [showModal, setShowModal] = useState(false);
  const [editTaskId, setEditTaskId] = useState(null);
  const [openCommentTaskId, setOpenCommentTaskId] = useState(null);
  
  // Backend integration
  const { tasks, loading, error, createTask, updateTask, deleteTask } = useTasks(workspaceId);

  const handleOpenModal = () => {
    setShowModal(true);
    setEditTaskId(null);
  };

  const handleEditTask = (task) => {
    setEditTaskId(task.id);
    setShowModal(true);
  };

  const handleDeleteTask = async (taskId) => {
    await deleteTask(taskId);
  };

  // Get the task being edited
  const editingTask = editTaskId ? tasks.find(t => t.id === editTaskId) : null;

  // Status badge color
  const statusColor = (status) => {
    if (status === 'Todo') return 'bg-gray-100 text-gray-700 border-gray-200';
    if (status === 'In Progress') return 'bg-yellow-50 text-yellow-800 border-yellow-200';
    return 'bg-green-50 text-green-800 border-green-200';
  };

  // Deduplicate tasks by id before rendering
  const uniqueTasks = Array.from(new Map(tasks.map(t => [t.id, t])).values());

  // Show loading state
  if (loading) {
    return (
      <section className="w-full">
        <div className="flex items-center justify-center py-8">
          <div className="text-gray-500">Loading tasks...</div>
        </div>
      </section>
    );
  }

  // Show error state
  if (error) {
    return (
      <section className="w-full">
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
          <p className="text-red-800">Error loading tasks: {error}</p>
        </div>
      </section>
    );
  }

  return (
    <section className="w-full">
      <div className="mb-6">
        <button
          className="py-2 px-6 bg-blue-600 text-white rounded hover:bg-blue-700 transition font-semibold flex items-center gap-2"
          onClick={handleOpenModal}
        >
          <FaPlus className="text-base" /> Add Task
        </button>
      </div>
      <Modal open={showModal} onClose={() => setShowModal(false)}>
        <TaskForm
          onSubmit={async (taskData) => {
            if (editTaskId) {
              // Update existing task
              const updates = {
                title: taskData.title,
                description: taskData.description,
                status: taskData.status,
                assigned_to: taskData.assigned_to,
                due_date: taskData.due_date,
              };
              const success = await updateTask(editTaskId, updates);
              if (success) {
                setEditTaskId(null);
                setShowModal(false);
              }
              return success;
            } else {
              // Create new task
              const success = await createTask(taskData);
              if (success) {
                setShowModal(false);
              }
              return success;
            }
          }}
          onCancel={() => {
            setShowModal(false);
            setEditTaskId(null);
          }}
          teamMembers={teamMembers}
          initialData={editingTask}
        />
      </Modal>
      {/* Task List Grid */}
      <div className="grid grid-cols-1 gap-8">
        {uniqueTasks.map((task) => (
          <TaskReactionWrapper 
            key={task.id} 
            task={task} 
            workspaceId={workspaceId}
            teamMembers={teamMembers}
            onEditTask={handleEditTask}
            onDeleteTask={handleDeleteTask}
            onUpdateTask={updateTask}
            openCommentTaskId={openCommentTaskId}
            setOpenCommentTaskId={setOpenCommentTaskId}
          />
        ))}
      </div>
    </section>
  );
}

// Separate component to handle reactions for each task
function TaskReactionWrapper({ 
  task, 
  workspaceId, 
  teamMembers, 
  onEditTask, 
  onDeleteTask, 
  onUpdateTask,
  openCommentTaskId,
  setOpenCommentTaskId
}) {
  const { reactions, userReactions, toggleReaction } = useTaskReactions(workspaceId, task.id);
  const [menuOpen, setMenuOpen] = useState(false);
  const menuRef = useRef();
  useEffect(() => {
    function handleClickOutside(event) {
      if (menuRef.current && !menuRef.current.contains(event.target)) {
        setMenuOpen(false);
      }
    }
    if (menuOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [menuOpen]);

  const handleReact = async (reactionType) => {
    await toggleReaction(reactionType);
  };

  const statusColor = (status) => {
    if (status === 'Todo') return 'bg-gray-100 text-gray-700 border-gray-200';
    if (status === 'In Progress') return 'bg-yellow-50 text-yellow-800 border-yellow-200';
    return 'bg-green-50 text-green-800 border-green-200';
  };

  return (
    <div
      className={
        `bg-white rounded-2xl shadow-md border border-gray-100 p-6 mb-4 transition-all duration-200 flex flex-col relative group ` +
        `hover:shadow-xl hover:-translate-y-1 `
      }
      style={{ borderLeft: `6px solid ${task.status === 'Todo' ? '#a0aec0' : task.status === 'In Progress' ? '#ecc94b' : task.status === 'Review' ? '#fbbf24' : '#38a169'}` }}
    >
      {/* Top row: Creator info, [Status & Assignee + 3-dot menu] */}
      <div className="flex items-center justify-between mb-4 w-full">
        {/* Left: Creator info */}
        <div className="flex items-center">
          <img
            src={task.creator_avatar || '/default-avatar.png'}
            alt={task.creator_name || 'Unknown'}
            className="w-9 h-9 rounded-full border mr-3 shadow-sm"
            title={task.creator_name || 'Unknown'}
          />
          <div className="flex flex-col">
            <span className="font-semibold text-base text-gray-900 leading-tight">
              {task.creator_name || 'Unknown'}
            </span>
            <span className="text-xs text-gray-500">{task.created_at ? new Date(task.created_at).toLocaleString() : 'Just now'}</span>
          </div>
          <span className="ml-2 text-xs text-gray-400">(Posted by)</span>
        </div>
        {/* Right: Status & Assignee group + 3-dot menu */}
        <div className="flex items-center gap-4">
          {/* Status & Assignee group */}
          <div className="flex items-end gap-6">
            {/* Status column */}
            <div className="flex flex-col items-center">
              <span className="text-xs text-gray-400 font-medium mb-1">Status</span>
              <span className={
                `px-3 py-1 rounded-full text-xs font-semibold ` +
                (task.status === 'Done' ? 'bg-green-100 text-green-800' :
                 task.status === 'In Progress' ? 'bg-blue-100 text-blue-800' :
                 task.status === 'Review' ? 'bg-yellow-100 text-yellow-800' :
                 'bg-gray-100 text-gray-800')
              }>
                {task.status}
              </span>
            </div>
            {/* Vertical divider */}
            <div className="h-9 border-l border-gray-300" />
            {/* Assignee column */}
            <div className="flex flex-col items-center">
              <span className="text-xs text-gray-400 font-medium mb-1">Assignee</span>
              {task.assigned_to ? (() => {
                const member = teamMembers.find(m => m.id === task.assigned_to);
                return member ? (
                  <div className="flex items-center">
                    <img
                      src={member.avatar}
                      alt={member.name}
                      className="w-7 h-7 rounded-full border mr-1 shadow-sm"
                      title={member.name}
                    />
                    <span className="text-xs text-gray-700 font-medium">{member.name}</span>
                  </div>
                ) : (
                  <span className="text-xs text-gray-400">Unknown user</span>
                );
              })() : <span className="text-xs text-gray-400">None</span>}
            </div>
          </div>
          {/* 3-dot menu */}
          <div className="relative flex items-center">
            <button
              onClick={() => setMenuOpen((open) => !open)}
              className="p-2 text-gray-500 hover:text-gray-700 focus:outline-none"
              aria-label="Task options"
            >
              <FaEllipsisH className="w-5 h-5" />
            </button>
            {menuOpen && (
              <div ref={menuRef} className="absolute right-0 mt-8 w-32 bg-white border rounded shadow z-20">
                <button
                  className="block w-full text-left px-4 py-2 text-blue-700 hover:bg-blue-100 font-semibold"
                  onClick={() => { setMenuOpen(false); onEditTask(task); }}
                >
                  Edit
                </button>
                <button
                  className="block w-full text-left px-4 py-2 text-red-600 hover:bg-gray-100 font-semibold"
                  onClick={() => { setMenuOpen(false); onDeleteTask(task.id); }}
                >
                  Delete
                </button>
              </div>
            )}
          </div>
        </div>
      </div>
      {/* Title and content */}
      <div className="mb-2 text-xl font-bold text-gray-900 leading-tight tracking-tight">{task.title}</div>
      <div className="mb-4 text-gray-600 text-base bg-gray-50 rounded-lg px-4 py-3 whitespace-pre-line border border-gray-100 shadow-sm">{task.description}</div>
      {/* Media preview */}
      {task.media && (
        <div className="my-2">
          {task.media.includes('video') || task.media.match(/\.(mp4|mov|avi|mkv|wmv|flv|webm)$/i) ? (
            <video src={task.media} controls className="mx-auto max-w-xs h-32 object-contain rounded-xl border" />
          ) : (
            <img src={task.media} alt="Task Media" className="mx-auto max-w-xs h-32 object-contain rounded-xl border" />
          )}
        </div>
      )}
      {/* Reactions Bar */}
      <div className="flex items-center gap-4 mt-4">
        <button
          className={`flex items-center text-xs rounded px-2 py-1 transition ${
            userReactions.includes('thumbsUp') 
              ? 'text-blue-700 bg-blue-100' 
              : 'text-blue-700 hover:underline'
          }`}
          onClick={() => handleReact('thumbsUp')}
        >
          <MdThumbUpOffAlt className="mr-1" /> Like {reactions.thumbsUp > 0 && <span className="ml-1 font-bold">{reactions.thumbsUp}</span>}
        </button>
        <button
          className={`flex items-center text-xs rounded px-2 py-1 transition ${
            userReactions.includes('heart') 
              ? 'text-red-700 bg-red-100' 
              : 'text-red-700 hover:underline'
          }`}
          onClick={() => handleReact('heart')}
        >
          <MdFavoriteBorder className="mr-1" /> Heart {reactions.heart > 0 && <span className="ml-1 font-bold">{reactions.heart}</span>}
        </button>
        <button
          className="flex items-center text-xs text-blue-700 hover:underline rounded px-2 py-1"
          onClick={() => setOpenCommentTaskId(openCommentTaskId === task.id ? null : task.id)}
        >
          <MdOutlineModeComment className="mr-1" /> Comment
        </button>
        {/* Status dropdown */}
        <select
          className="ml-auto text-xs border rounded px-2 py-1 text-gray-800 bg-white"
          value={task.status}
          onChange={async (e) => {
            await onUpdateTask(task.id, { status: e.target.value });
          }}
        >
          <option value="Todo">Todo</option>
          <option value="In Progress">In Progress</option>
          <option value="Done">Done</option>
        </select>
      </div>
      {/* Comments section */}
      {openCommentTaskId === task.id && (
        <div className="mt-4 pt-4 border-t border-gray-200">
          <CommentSection taskId={task.id} workspaceId={workspaceId} />
        </div>
      )}
    </div>
  );
} 