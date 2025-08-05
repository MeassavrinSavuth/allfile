import React from 'react';

export default function FacebookPosts({ posts, loading, error, searchQuery, setSearchQuery }) {
  // Placeholder page name and avatar (replace with real data if available)
  const pageName = 'Facebook Page';
  const pageAvatar = '/default-avatar.png';

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
      {loading && <div className="text-center text-gray-500">Loading Facebook posts...</div>}
      {error && <div className="text-center text-red-500">{error}</div>}
      {!loading && !error && posts.length === 0 && (
        <div className="text-center text-gray-400">No Facebook posts found.</div>
      )}
      <div className="flex flex-col gap-6 items-center">
        {posts.map((post) => {
          const message = post.message || '';
          const image = post.full_picture;
          const likes = post.likes?.summary?.total_count || 0;
          const comments = post.comments?.summary?.total_count || 0;
          const attachments = post.attachments || [];
          return (
            <div key={post.id} className="bg-white rounded-lg shadow border border-gray-200 max-w-xl w-full">
              {/* Header: Page/Profile */}
              <div className="flex items-center gap-3 px-4 pt-4">
                <img
                  src={pageAvatar}
                  alt={pageName}
                  className="w-10 h-10 rounded-full border"
                />
                <div>
                  <div className="font-semibold text-gray-900">{pageName}</div>
                  <div className="text-xs text-gray-500">{new Date(post.created_time).toLocaleString()}</div>
                </div>
              </div>
              {/* Content */}
              <div className="px-4 py-2 text-gray-900 whitespace-pre-line">{message}</div>
              {/* Images: grid like Facebook, not clickable */}
              {attachments.length > 0 ? (
                <div className={`grid grid-cols-${attachments.length === 1 ? 1 : 2} gap-1 px-4 py-2`}>
                  {attachments.slice(0, 4).map((img, idx) => (
                    <div key={idx} className="relative">
                      <img src={img} alt="Facebook attachment" className="w-full h-48 object-cover rounded" />
                      {/* Overlay for +N if more than 4 images */}
                      {idx === 3 && attachments.length > 4 && (
                        <div className="absolute inset-0 bg-black bg-opacity-50 flex items-center justify-center text-white text-2xl font-bold rounded">
                          +{attachments.length - 4}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              ) : image && (
                <img src={image} alt="Facebook post" className="w-full max-w-md object-cover rounded-lg border mx-auto" />
              )}
              {/* Action Bar (no counts, not functional, only once) */}
              <div className="flex justify-around py-2 text-gray-700 text-sm font-semibold border-b">
                <button className="flex items-center gap-1 hover:bg-gray-100 px-4 py-1 rounded">
                  <span className="text-blue-600">👍</span> Like {likes}
                </button>
                <button className="flex items-center gap-1 hover:bg-gray-100 px-4 py-1 rounded">
                  <span className="text-green-600">💬</span> Comment {comments}
                </button>
                <button className="flex items-center gap-1 hover:bg-gray-100 px-4 py-1 rounded">
                  <span className="text-gray-600">↗️</span> Share {/* You can show 0 or leave blank if not available */}
                </button>
              </div>
              {/* View on Facebook (only once, in footer) */}
              <div className="px-4 pb-3">
                {post.permalink_url && (
                  <a href={post.permalink_url} target="_blank" rel="noopener noreferrer" className="text-blue-600 text-xs underline">
                    View on Facebook
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