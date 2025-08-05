'use client';
import React, { useState, useEffect } from 'react';
import { FaFacebook, FaInstagram, FaYoutube, FaTwitter } from 'react-icons/fa';
import { SiMastodon } from 'react-icons/si';
import { useProtectedFetch } from '../../hooks/auth/useProtectedFetch';
import MastodonPosts from '../../components/postfolder/MastodonPosts';

import YouTubePosts from '../../components/postfolder/YouTubePosts';
import FacebookPosts from '../../components/postfolder/FacebookPosts';
import InstagramPosts from '../../components/postfolder/InstagramPosts';

function AppIconCard({ icon, label, color, onClick }) {
  return (
    <button
      onClick={onClick}
      className="flex flex-col items-center justify-center w-24 h-24 rounded-2xl shadow-lg bg-white hover:bg-gray-50 transition cursor-pointer border border-gray-100"
      style={{ boxShadow: '0 4px 16px rgba(0,0,0,0.08)' }}
    >
      <div className={`text-4xl mb-2 ${color}`}>{icon}</div>
      <span className="text-xs font-semibold text-gray-700">{label}</span>
    </button>
  );
}

export default function PostsFolderPage() {
  const [selectedPlatform, setSelectedPlatform] = useState(null);
  const [mastodonPosts, setMastodonPosts] = useState([]);
  const [twitterPosts, setTwitterPosts] = useState([]);
  const [youtubePosts, setYouTubePosts] = useState([]);
  const [facebookPosts, setFacebookPosts] = useState([]);
  const [instagramPosts, setInstagramPosts] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [searchQuery, setSearchQuery] = useState('');
  const protectedFetch = useProtectedFetch();

  useEffect(() => {
    if (selectedPlatform === 'mastodon') {
      setLoading(true);
      setError(null);
      protectedFetch('/api/mastodon/posts')
        .then(async (res) => {
          if (!res) return;
          const data = await res.json();
          setMastodonPosts(data);
        })
        .catch(() => {
          setError('Failed to fetch Mastodon posts');
        })
        .finally(() => setLoading(false));
    } else if (selectedPlatform === 'twitter') {
      setLoading(true);
      setError(null);
      protectedFetch('/api/twitter/posts')
        .then(async (res) => {
          if (!res) return;
          if (res.status === 429) {
            setError('Twitter API rate limit exceeded. Please wait a few minutes and try again.');
            return;
          }
          const data = await res.json();
          // Twitter API returns { data: [...], includes: { media: [...] }, ... }
          setTwitterPosts(data);
        })
        .catch(() => {
          setError('Failed to fetch Twitter posts');
        })
        .finally(() => setLoading(false));
    } else if (selectedPlatform === 'youtube') {
      setLoading(true);
      setError(null);
      protectedFetch('/api/youtube/posts')
        .then(async (res) => {
          if (!res) return;
          const data = await res.json();
          // YouTube API returns { items: [...] }
          setYouTubePosts(data.items || []);
        })
        .catch(() => {
          setError('Failed to fetch YouTube videos');
        })
        .finally(() => setLoading(false));
    } else if (selectedPlatform === 'facebook') {
      setLoading(true);
      setError(null);
      protectedFetch('/api/facebook/posts')
        .then(async (res) => {
          if (!res) return;
          const data = await res.json();
          // Facebook API returns { data: [...] }
          setFacebookPosts(data.data || []);
        })
        .catch(() => {
          setError('Failed to fetch Facebook posts');
        })
        .finally(() => setLoading(false));
    } else if (selectedPlatform === 'instagram') {
      setLoading(true);
      setError(null);
      protectedFetch('/api/instagram/posts')
        .then(async (res) => {
          if (!res) return;
          const data = await res.json();
          // Instagram API returns { data: [...] }
          setInstagramPosts(data.data || []);
        })
        .catch(() => {
          setError('Failed to fetch Instagram posts');
        })
        .finally(() => setLoading(false));
    }
  }, [selectedPlatform]);

  const handlePlatformClick = (platform) => {
    setSelectedPlatform(platform);
    setMastodonPosts([]);
    setTwitterPosts([]);
    setYouTubePosts([]);
    setFacebookPosts([]);
    setInstagramPosts([]);
    setError(null);
    setSearchQuery('');
  };

  // Filter posts by search query (case-insensitive, search in HTML content or tweet text)
  const filteredMastodonPosts = mastodonPosts.filter(post =>
    post.content.toLowerCase().includes(searchQuery.toLowerCase())
  );
  const filteredTwitterPosts = (twitterPosts.data || []).filter(tweet =>
    tweet.text && tweet.text.toLowerCase().includes(searchQuery.toLowerCase())
  );
  const filteredYouTubePosts = youtubePosts.filter(video => {
    const snippet = video.snippet || {};
    return (
      (snippet.title && snippet.title.toLowerCase().includes(searchQuery.toLowerCase())) ||
      (snippet.description && snippet.description.toLowerCase().includes(searchQuery.toLowerCase()))
    );
  });
  const filteredFacebookPosts = facebookPosts.filter(post =>
    (post.message && post.message.toLowerCase().includes(searchQuery.toLowerCase()))
  );
  const filteredInstagramPosts = instagramPosts.filter(post =>
    (post.caption && post.caption.toLowerCase().includes(searchQuery.toLowerCase()))
  );
  // Helper: get media for a tweet
  const getTweetMedia = (tweet, includes) => {
    if (!tweet.attachments || !tweet.attachments.media_keys || !includes || !includes.media) return [];
    return tweet.attachments.media_keys.map(key =>
      includes.media.find(m => m.media_key === key)
    ).filter(Boolean);
  };
  // Helper: get tweet author info (for now, use the logged-in user info if available)
  const getTweetAuthor = (twitterPosts) => {
    // Twitter API v2 /users/:id/tweets does not return author info per tweet, but you can use the user info from /users/me
    if (twitterPosts && twitterPosts.includes && twitterPosts.includes.users && twitterPosts.includes.users.length > 0) {
      return twitterPosts.includes.users[0];
    }
    if (twitterPosts && twitterPosts.data && twitterPosts.data.length > 0 && twitterPosts.data[0].author) {
      return twitterPosts.data[0].author;
    }
    return null;
  };
  const tweetAuthor = getTweetAuthor(twitterPosts);

  return (
    <div className="w-full max-w-4xl mx-auto py-10 px-4">
      <h1 className="text-2xl md:text-3xl font-bold text-gray-900 mb-6 text-center">
        Select platform to see the post
      </h1>
      <div className="flex gap-6 justify-center my-6 flex-wrap">
        <AppIconCard icon={<FaFacebook />} label="Facebook" color="text-blue-600" onClick={() => handlePlatformClick('facebook')} />
        <AppIconCard icon={<FaInstagram />} label="Instagram" color="text-pink-500" onClick={() => handlePlatformClick('instagram')} />
        <AppIconCard icon={<FaYoutube />} label="YouTube" color="text-red-600" onClick={() => handlePlatformClick('youtube')} />
        <AppIconCard icon={<SiMastodon />} label="Mastodon" color="text-purple-600" onClick={() => handlePlatformClick('mastodon')} />
        <AppIconCard icon={<FaTwitter />} label="Twitter" color="text-sky-500" onClick={() => handlePlatformClick('twitter')} />
      </div>
      {selectedPlatform && (
        <div className="text-center text-lg mt-4">
          Selected platform: <span className="font-semibold capitalize">{selectedPlatform}</span>
        </div>
      )}
      {/* Mastodon posts display */}
      {selectedPlatform === 'mastodon' && (
        <MastodonPosts
          posts={filteredMastodonPosts}
          loading={loading}
          error={error}
          searchQuery={searchQuery}
          setSearchQuery={setSearchQuery}
        />
      )}
      {/* Twitter posts display */}
      {selectedPlatform === 'twitter' && (
        <TwitterPosts
          posts={filteredTwitterPosts}
          includes={twitterPosts.includes}
          tweetAuthor={tweetAuthor}
          loading={loading}
          error={error}
          searchQuery={searchQuery}
          setSearchQuery={setSearchQuery}
        />
      )}
      {/* YouTube posts display */}
      {selectedPlatform === 'youtube' && (
        <YouTubePosts
          posts={filteredYouTubePosts}
          loading={loading}
          error={error}
          searchQuery={searchQuery}
          setSearchQuery={setSearchQuery}
        />
      )}
      {/* Facebook posts display */}
      {selectedPlatform === 'facebook' && (
        <FacebookPosts
          posts={filteredFacebookPosts}
          loading={loading}
          error={error}
          searchQuery={searchQuery}
          setSearchQuery={setSearchQuery}
        />
      )}
      {/* Instagram posts display */}
      {selectedPlatform === 'instagram' && (
        <InstagramPosts
          posts={filteredInstagramPosts}
          loading={loading}
          error={error}
          searchQuery={searchQuery}
          setSearchQuery={setSearchQuery}
        />
      )}
      {/* Platform selector and posts grid/table will go here */}
    </div>
  );
}


