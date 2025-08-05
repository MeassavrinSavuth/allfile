import { useState, useRef, useEffect } from 'react';
import { FaThumbsUp, FaCommentAlt, FaShare, FaGlobeAmericas, FaEllipsisH } from 'react-icons/fa';

const REACTION_EMOJIS = ['👍', '❤️', '😂', '🎉', '👎'];

export default function MiniFacebookPreview({ task, onReact, showReactions = true, showTitle = false, fullWidth = false, onEdit, onPost, onDelete }) {
  const [menuOpen, setMenuOpen] = useState(false);
  const menuRef = useRef();
  const [comments, setComments] = useState(task.comments || []);
  const [newComment, setNewComment] = useState("");
  const [showComments, setShowComments] = useState(false);
  const [openReactionPicker, setOpenReactionPicker] = useState(null);

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

  const handleAddComment = (e) => {
    e.preventDefault();
    if (!newComment.trim()) return;
    setComments((prev) => [
      ...prev,
      {
        author: task.author,
        content: newComment,
        timestamp: new Date().toLocaleString(),
        reactions: {},
      },
    ]);
    setNewComment("");
  };

  return (
    <div className={`bg-white rounded-xl shadow border border-gray-200 ${fullWidth ? 'w-full' : 'w-full max-w-lg mx-auto'}`}>
      {/* Header */}
      <div className="flex items-center px-4 pt-4 pb-2 justify-between relative">
        <div className="flex items-center">
          <img src={task.author.avatar} alt={task.author.name} className="w-9 h-9 rounded-full border mr-3" />
          <div className="flex flex-col">
            <span className="font-semibold text-[15px] text-gray-900 leading-tight">{task.author.name}</span>
            <span className="flex items-center text-xs text-gray-500 gap-1">
              Just now <FaGlobeAmericas className="inline-block ml-1 text-xs" />
            </span>
          </div>
        </div>
        <div className="relative">
          <button
            className="text-gray-400 text-lg cursor-pointer hover:text-gray-600"
            onClick={() => setMenuOpen((open) => !open)}
          >
            <FaEllipsisH />
          </button>
          {menuOpen && (
            <div ref={menuRef} className="absolute right-0 mt-2 w-32 bg-white border rounded shadow z-10">
              <button className="block w-full text-left px-4 py-2 text-blue-700 hover:bg-blue-100 font-semibold" onClick={() => { setMenuOpen(false); onEdit && onEdit(); }}>Edit</button>
              <button className="block w-full text-left px-4 py-2 text-blue-700 hover:bg-blue-100 font-semibold" onClick={() => { setMenuOpen(false); onPost && onPost(); }}>Post</button>
              <button className="block w-full text-left px-4 py-2 text-red-600 hover:bg-gray-100 font-semibold" onClick={() => { setMenuOpen(false); onDelete && onDelete(); }}>Delete</button>
            </div>
          )}
        </div>
      </div>
      {/* Content */}
      <div className="px-4 pb-2 min-h-[120px]">
        <div className="text-[15px] text-gray-900 mb-2 whitespace-pre-line">{showTitle ? task.title : task.description}</div>
      </div>
      {/* Media */}
      {task.photo && (
        <div className="px-4 pb-4">
          <img src={task.photo} alt="Preview" className="mx-auto my-2 max-w-xs h-60 object-contain rounded-xl border" />
        </div>
      )}
      {/* Actions Bar */}
      {showReactions && (
        <div className="flex justify-around items-center py-2 border-t border-gray-200 text-gray-700 text-sm">
          <button onClick={() => onReact(task.id, 'thumbsUp')} className="flex items-center space-x-2 p-2 rounded-lg hover:bg-gray-100 transition">
            <FaThumbsUp />
            <span>Like</span>
            <span className="ml-1 text-blue-600 font-bold">{task.reactions.thumbsUp}</span>
          </button>
          <button
            className="flex items-center space-x-2 p-2 rounded-lg hover:bg-gray-100 transition"
            onClick={() => setShowComments((v) => !v)}
            aria-expanded={showComments}
          >
            <FaCommentAlt />
            <span>Comment</span>
          </button>
          <button className="flex items-center space-x-2 p-2 rounded-lg hover:bg-gray-100 transition" disabled>
            <FaShare />
            <span>Share</span>
          </button>
        </div>
      )}
      {/* Comments Section (toggle) */}
      {showComments && (
        <div className="px-4 pt-2 pb-4">
          <div className="space-y-2 mb-2">
            {comments.length === 0 && <div className="text-xs text-gray-400">No comments yet.</div>}
            {comments.map((c, i) => (
              <div key={i} className="flex items-start gap-2">
                <img src={c.author.avatar} alt={c.author.name} className="w-7 h-7 rounded-full border" />
                <div>
                  <div className="text-xs text-gray-800 font-semibold">{c.author.name} <span className="ml-2 text-gray-500">{c.timestamp}</span></div>
                  <div className="text-sm text-gray-700">{c.content}</div>
                  <div className="flex gap-2 mt-1 items-center">
                    {/* Shown reactions (count > 0) */}
                    {REACTION_EMOJIS.filter(emoji => c.reactions?.[emoji] > 0).map((emoji) => (
                      <span key={emoji} className="inline-flex items-center px-2 py-0.5 rounded-full bg-gray-100 text-gray-800 text-base font-medium shadow-sm border border-gray-200 mr-1">
                        {emoji} <span className="ml-1 text-xs font-semibold">{c.reactions[emoji]}</span>
                      </span>
                    ))}
                    {/* Add reaction button */}
                    <button
                      type="button"
                      className="ml-1 px-2 py-0.5 rounded-full bg-gray-200 text-gray-600 text-base font-semibold hover:bg-gray-300 focus:outline-none"
                      onClick={() => setOpenReactionPicker(openReactionPicker === i ? null : i)}
                      aria-label="Add reaction"
                    >
                      +
                    </button>
                    {/* Emoji picker row */}
                    {openReactionPicker === i && (
                      <div className="flex gap-1 ml-2 bg-white border border-gray-200 rounded-full px-2 py-1 shadow z-10">
                        {REACTION_EMOJIS.map((emoji) => (
                          <button
                            key={emoji}
                            type="button"
                            className="text-lg px-1 rounded-full hover:bg-gray-100 focus:outline-none"
                            onClick={() => {
                              setComments(prev => prev.map((com, idx) => idx === i ? {
                                ...com,
                                reactions: {
                                  ...com.reactions,
                                  [emoji]: (com.reactions?.[emoji] || 0) + 1
                                }
                              } : com));
                              setOpenReactionPicker(null);
                            }}
                            aria-label={`React with ${emoji}`}
                          >
                            {emoji}
                          </button>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
          <form onSubmit={handleAddComment} className="flex gap-2 mt-2">
            <input
              className="flex-1 border rounded px-2 py-1 text-xs text-gray-900 placeholder-gray-500"
              value={newComment}
              onChange={e => setNewComment(e.target.value)}
              placeholder="Add a comment..."
            />
            <button type="submit" className="px-3 py-1 bg-blue-600 text-white rounded text-xs font-semibold hover:bg-blue-700">Post</button>
          </form>
        </div>
      )}
    </div>
  );
} 