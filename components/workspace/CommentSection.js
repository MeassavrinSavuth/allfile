import { useState } from 'react';
import { useComments } from '../../hooks/api/useComments';

const CommentSection = ({ taskId, workspaceId }) => {
  const [newComment, setNewComment] = useState('');
  const { comments, loading, error, addComment, deleteComment } = useComments(workspaceId, taskId);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!newComment.trim()) return;

    const success = await addComment(newComment.trim());
    if (success) {
      setNewComment('');
    }
  };

  const handleDelete = async (commentId) => {
    if (confirm('Are you sure you want to delete this comment?')) {
      await deleteComment(commentId);
    }
  };

  if (loading) {
    return <div className="text-gray-500 text-sm">Loading comments...</div>;
  }

  if (error) {
    return <div className="text-red-500 text-sm">Error loading comments: {error}</div>;
  }

  return (
    <div className="space-y-4">
      <h4 className="font-medium text-gray-900">Comments ({comments.length})</h4>
      
      {/* Add comment form */}
      <form onSubmit={handleSubmit} className="space-y-2">
        <textarea
          value={newComment}
          onChange={(e) => setNewComment(e.target.value)}
          placeholder="Add a comment..."
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none text-gray-900 placeholder-gray-400"
          rows="2"
        />
        <div className="flex justify-end">
          <button
            type="submit"
            disabled={!newComment.trim()}
            className="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Add Comment
          </button>
        </div>
      </form>

      {/* Comments list */}
      <div className="space-y-3">
        {comments.length === 0 ? (
          <p className="text-gray-500 text-sm">No comments yet. Be the first to comment!</p>
        ) : (
          comments.map((comment) => (
            <div key={comment.id} className="bg-gray-50 rounded-lg p-3">
              <div className="flex justify-between items-start">
                <div className="flex-1">
                  <p className="text-sm text-gray-900 mb-1">{comment.content}</p>
                  <div className="flex items-center space-x-2 text-xs text-gray-500">
                    <span>By: {comment.user_name || comment.user_id}</span>
                    <span>•</span>
                    <span>{new Date(comment.created_at).toLocaleString()}</span>
                  </div>
                </div>
                <button
                  onClick={() => handleDelete(comment.id)}
                  className="p-1 text-gray-400 hover:text-red-500 focus:outline-none"
                >
                  <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default CommentSection; 