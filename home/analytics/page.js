"use client";

import React, { useEffect, useState } from 'react';

function timeAgo(dateString) {
  const now = new Date();
  const date = new Date(dateString);
  const diff = Math.floor((now - date) / 1000);
  if (diff < 60) return `${diff}s ago`;
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
  return date.toLocaleDateString();
}

function stripHtml(html) {
  if (!html) return '';
  return html.replace(/<[^>]+>/g, '').replace(/&nbsp;/g, ' ').trim();
}

export default function AnalyticsPage() {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    async function fetchAnalytics() {
      setLoading(true);
      setError(null);
      try {
        const res = await fetch('/api/analytics/mastodon', {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('accessToken')}`,
          },
        });
        if (!res.ok) {
          const msg = await res.text();
          throw new Error(msg || 'Failed to fetch analytics');
        }
        const json = await res.json();
        setData(json);
      } catch (err) {
        setError(err.message || 'Error fetching analytics');
      } finally {
        setLoading(false);
      }
    }
    fetchAnalytics();
  }, []);

  return (
    <div className="max-w-4xl mx-auto py-8 px-4">
      {/* Dashboard welcome section */}
      <div className="bg-gradient-to-r from-indigo-100 to-blue-100 rounded-lg shadow p-6 mb-10 flex flex-col items-center text-center">
        <h1 className="text-3xl font-bold mb-2 text-gray-900">Welcome!</h1>
        <p className="text-gray-900 text-lg mb-2 font-medium">This is your Social Media Analytics Dashboard.</p>
        <p className="text-gray-800 font-medium">Click on a social media platform below to see your analytics.</p>
        <div className="flex gap-6 mt-4">
          <span className="flex flex-col items-center cursor-pointer hover:scale-105 transition-transform">
            <img src="/public/logo-ss.png" alt="Mastodon" className="w-10 h-10 rounded" />
            <span className="text-xs mt-1 text-gray-900 font-semibold">Mastodon</span>
          </span>
          <span className="flex flex-col items-center cursor-pointer opacity-50">
            <img src="/public/youtube.svg" alt="YouTube" className="w-10 h-10 rounded" />
            <span className="text-xs mt-1 text-gray-900 font-semibold">YouTube</span>
          </span>
          <span className="flex flex-col items-center cursor-pointer opacity-50">
            <img src="/public/facebook.svg" alt="Facebook" className="w-10 h-10 rounded" />
            <span className="text-xs mt-1 text-gray-900 font-semibold">Facebook</span>
          </span>
          <span className="flex flex-col items-center cursor-pointer opacity-50">
            <img src="/public/instagram.svg" alt="Instagram" className="w-10 h-10 rounded" />
            <span className="text-xs mt-1 text-gray-900 font-semibold">Instagram</span>
          </span>
        </div>
      </div>
      {/* Mastodon analytics section */}
      <h2 className="text-2xl font-bold mb-6 text-gray-900">Mastodon Analytics</h2>
      {loading && <div className="text-center text-gray-500">Loading analytics...</div>}
      {error && <div className="text-center text-red-500">{error}</div>}
      {data && (
        <>
          {/* Summary cards */}
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-10">
            <div className="bg-white rounded-lg shadow p-4 flex flex-col items-center">
              <span className="text-2xl font-bold text-indigo-700">{data.totalPosts}</span>
              <span className="text-xs text-gray-900 mt-1 font-medium">Total Posts</span>
            </div>
            <div className="bg-white rounded-lg shadow p-4 flex flex-col items-center">
              <span className="text-2xl font-bold text-pink-700">{data.totalFavourites}</span>
              <span className="text-xs text-gray-900 mt-1 font-medium">Favourites</span>
            </div>
            <div className="bg-white rounded-lg shadow p-4 flex flex-col items-center">
              <span className="text-2xl font-bold text-green-700">{data.totalBoosts}</span>
              <span className="text-xs text-gray-900 mt-1 font-medium">Boosts</span>
            </div>
            <div className="bg-white rounded-lg shadow p-4 flex flex-col items-center">
              <span className="text-2xl font-bold text-blue-700">{data.totalReplies}</span>
              <span className="text-xs text-gray-900 mt-1 font-medium">Replies</span>
            </div>
          </div>
          {/* Top posts table */}
          <div className="bg-white rounded-lg shadow p-6">
            <h3 className="text-xl font-semibold mb-4 text-gray-900">Top Posts</h3>
            <div className="overflow-x-auto">
              <table className="min-w-full text-sm border-separate border-spacing-y-2">
                <thead>
                  <tr className="text-left text-gray-700 border-b">
                    <th className="py-2 pr-4">Content</th>
                    <th className="py-2 px-2 text-center">Favourites</th>
                    <th className="py-2 px-2 text-center">Boosts</th>
                    <th className="py-2 px-2 text-center">Replies</th>
                    <th className="py-2 pl-4">Date</th>
                  </tr>
                </thead>
                <tbody>
                  {data.topPosts.map((post) => (
                    <tr key={post.id} className="bg-gray-50 hover:bg-indigo-50 rounded-lg shadow-sm">
                      <td className="py-2 pr-4 max-w-xs truncate text-gray-900 font-medium" title={stripHtml(post.content)}>
                        {stripHtml(post.content)}
                      </td>
                      <td className="py-2 px-2 text-center text-pink-700 font-semibold">{post.favourites_count}</td>
                      <td className="py-2 px-2 text-center text-green-700 font-semibold">{post.reblogs_count}</td>
                      <td className="py-2 px-2 text-center text-blue-700 font-semibold">{post.replies_count}</td>
                      <td className="py-2 pl-4 text-xs text-gray-700 whitespace-nowrap">{timeAgo(post.created_at)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
