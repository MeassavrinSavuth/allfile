import React from 'react';

export default function InstagramPosts({ posts, loading, error, searchQuery, setSearchQuery }) {
  return (
    <div className="mt-8">
      {/* Search bar */}
      <div className="mb-6 flex justify-center">
        <input
          type="text"
          placeholder="Search posts..."
          value={searchQuery}
          onChange={e => setSearchQuery(e.target.value)}
          className="border rounded px-3 py-2 w-full max-w-md text-gray-700"
        />
      </div>
      {loading && <div className="text-center text-gray-500">Loading Instagram posts...</div>}
      {error && <div className="text-center text-red-500">{error}</div>}
      {!loading && !error && posts.length === 0 && (
        <div className="text-center text-gray-400">No Instagram posts found.</div>
      )}
      <div className="flex flex-col gap-6 items-center">
        {posts.map((post) => {
          const caption = post.caption || '';
          const mediaType = post.media_type;
          const mediaUrl = post.media_url || post.thumbnail_url;
          const likes = post.like_count || 0;
          const comments = post.comments_count || 0;
          return (
            <div key={post.id} className="bg-white rounded-lg shadow border border-gray-200 max-w-md w-full">
              {/* Media */}
              {mediaType === 'IMAGE' || mediaType === 'CAROUSEL_ALBUM' ? (
                <img src={mediaUrl} alt="Instagram post" className="w-full h-80 object-cover rounded-t-lg" />
              ) : mediaType === 'VIDEO' ? (
                <video controls className="w-full h-80 object-cover rounded-t-lg">
                  <source src={mediaUrl} type="video/mp4" />
                  Your browser does not support the video tag.
                </video>
              ) : null}
              {/* Caption */}
              <div className="px-4 py-2 text-gray-900 whitespace-pre-line">{caption}</div>
              {/* Like/Comment counts */}
              <div className="flex items-center gap-6 px-4 py-2 text-gray-600 text-sm border-b">
                <span className="flex items-center gap-1"><span className="text-pink-500">❤️</span> {likes}</span>
                <span className="flex items-center gap-1"><span className="text-blue-500">💬</span> {comments}</span>
              </div>
              {/* Date and link */}
              <div className="flex items-center justify-between px-4 py-2">
                <div className="text-xs text-gray-500">{new Date(post.timestamp).toLocaleString()}</div>
                {post.permalink && (
                  <a href={post.permalink} target="_blank" rel="noopener noreferrer" className="text-blue-600 text-xs underline">
                    View on Instagram
                  </a>
                )}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
} 