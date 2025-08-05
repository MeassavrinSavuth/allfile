import React, { useState } from 'react';

function timeAgo(dateString) {
  const now = new Date();
  const date = new Date(dateString);
  const diff = Math.floor((now - date) / 1000);
  if (diff < 60) return `${diff}s ago`;
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
  return date.toLocaleDateString();
}

export default function MastodonPosts({ posts, loading, error, searchQuery, setSearchQuery }) {
  // For demo: track which post is "reacted" (not persistent)
  const [active, setActive] = useState({});

  const handleFakeAction = (postId, action) => {
    setActive((prev) => ({ ...prev, [postId]: { ...prev[postId], [action]: !prev[postId]?.[action] } }));
  };

  return (
    <div className="mt-8">
      {/* Search bar */}
      <div className="mb-6 flex justify-center">
        <input
          type="text"
          placeholder="Search Mastodon posts..."
          value={searchQuery}
          onChange={e => setSearchQuery(e.target.value)}
          className="border rounded px-3 py-2 w-full max-w-md text-gray-700 focus:ring-2 focus:ring-indigo-400 focus:border-indigo-400 outline-none shadow-sm"
        />
      </div>
      {loading && <div className="text-center text-gray-500">Loading Mastodon posts...</div>}
      {error && <div className="text-center text-red-500">{error}</div>}
      {!loading && !error && posts.length === 0 && (
        <div className="text-center text-gray-400">No Mastodon posts found.</div>
      )}
      <div className="grid gap-6">
        {posts.map((post) => (
          <div
            key={post.id}
            className="bg-white rounded-lg shadow border border-gray-200 flex flex-col gap-3 p-5 max-w-2xl mx-auto"
          >
            {/* Author section */}
            <div className="flex items-center gap-3 mb-1">
              {post.account && post.account.avatar && (
                <img
                  src={post.account.avatar}
                  alt={post.account.display_name || post.account.username || 'User'}
                  className="w-10 h-10 rounded-full border"
                />
              )}
              <div>
                <div className="font-semibold text-gray-900 text-base">{post.account?.display_name || post.account?.username || 'User'}</div>
                <div className="text-xs text-gray-500">@{post.account?.acct}</div>
              </div>
              <div className="ml-auto text-xs text-gray-400">{timeAgo(post.created_at)}</div>
            </div>
            {/* Content section */}
            <div className="text-gray-800 text-[15px] leading-relaxed" style={{wordBreak: 'break-word'}} dangerouslySetInnerHTML={{ __html: post.content }} />
            {/* Media section */}
            {post.media_attachments && post.media_attachments.length > 0 && (
              <div className="flex flex-wrap gap-3 mt-1">
                {post.media_attachments.map((media) => (
                  <div key={media.id} className="max-w-xs">
                    {media.type === 'image' && (
                      <img src={media.url} alt={media.description || 'Mastodon media'} className="rounded border max-h-48 object-contain" />
                    )}
                    {media.type === 'video' && (
                      <video controls className="rounded border max-h-48 w-full">
                        <source src={media.url} type={media.mime_type || 'video/mp4'} />
                        Your browser does not support the video tag.
                      </video>
                    )}
                  </div>
                ))}
              </div>
            )}
            {/* Engagement section with interactive-looking actions */}
            <div className="flex gap-6 items-center text-gray-600 text-sm mt-2 border-t pt-2">
              {/* Favorite */}
              <button
                className={`flex items-center gap-1 group focus:outline-none transition-colors ${active[post.id]?.favorite ? 'text-pink-600' : 'hover:text-pink-500'}`}
                title="Favorite"
                onClick={() => handleFakeAction(post.id, 'favorite')}
                type="button"
              >
                <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20"><path d="M3.172 5.172a4 4 0 015.656 0L10 6.343l1.172-1.171a4 4 0 115.656 5.656L10 17.657l-6.828-6.829a4 4 0 010-5.656z" /></svg>
                <span>{post.favourites_count}</span>
                <span className="ml-1 text-xs hidden sm:inline">Favorite</span>
              </button>
              {/* Boost */}
              <button
                className={`flex items-center gap-1 group focus:outline-none transition-colors ${active[post.id]?.boost ? 'text-green-600' : 'hover:text-green-500'}`}
                title="Boost"
                onClick={() => handleFakeAction(post.id, 'boost')}
                type="button"
              >
                <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20"><path d="M4 10a1 1 0 011-1h6.586l-3.293-3.293a1 1 0 111.414-1.414l5 5a1 1 0 010 1.414l-5 5a1 1 0 01-1.414-1.414L11.586 11H5a1 1 0 01-1-1z" /></svg>
                <span>{post.reblogs_count}</span>
                <span className="ml-1 text-xs hidden sm:inline">Boost</span>
              </button>
              {/* Reply */}
              <button
                className={`flex items-center gap-1 group focus:outline-none transition-colors ${active[post.id]?.reply ? 'text-blue-600' : 'hover:text-blue-500'}`}
                title="Reply"
                onClick={() => handleFakeAction(post.id, 'reply')}
                type="button"
              >
                <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20"><path d="M18 13v-2a4 4 0 00-4-4H6.414l1.293-1.293a1 1 0 00-1.414-1.414l-3 3a1 1 0 000 1.414l3 3a1 1 0 001.414-1.414L6.414 11H14a2 2 0 012 2v2a1 1 0 102 0z" /></svg>
                <span>{post.replies_count}</span>
                <span className="ml-1 text-xs hidden sm:inline">Reply</span>
              </button>
              {post.url && (
                <a
                  href={post.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="ml-auto px-3 py-1 rounded bg-indigo-500 text-white text-xs font-medium hover:bg-indigo-600 transition-colors"
                >
                  View on Mastodon
                </a>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
} 