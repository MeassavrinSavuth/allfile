'use client';

import { useState } from 'react';
import PostEditor from '../../components/PostEditor';
import PostPreview from '../../components/PostPreview';
import PostQueue from '../../components/PostQueue';
import { useMultiPlatformPublish } from '../../hooks/api/useMultiPlatformPublish';

// Define the list of platforms here, or import from a constants file if preferred
const platformsList = ['facebook', 'instagram', 'youtube', 'twitter', 'mastodon'];

export default function CreatePostPage() {
  const [selectedPlatforms, setSelectedPlatforms] = useState([]);
  const [message, setMessage] = useState('');
  const [youtubeConfig, setYoutubeConfig] = useState({ title: '', description: '' });
  const [mediaFiles, setMediaFiles] = useState([]); // array of Cloudinary URLs
  const [isPublishing, setIsPublishing] = useState(false);
  const [status, setStatus] = useState(null); // { success: bool, message: string }
  const [postQueue, setPostQueue] = useState([]);

  const togglePlatform = (platform) => {
    setSelectedPlatforms((prev) =>
      prev.includes(platform)
        ? prev.filter((p) => p !== platform)
        : [...prev, platform]
    );
  };

  const { publish } = useMultiPlatformPublish({
    message,
    mediaFiles,
    youtubeConfig,
  });

  const handlePublish = async () => {
    if (selectedPlatforms.length === 0) {
      setStatus({ success: false, message: 'Please select at least one platform.' });
      return;
    }
    setIsPublishing(true);
    setStatus(null);

    const newQueueItems = selectedPlatforms.map(platform => ({
      id: Date.now() + Math.random(),
      platform: platform,
      status: 'pending',
      timestamp: new Date().toISOString(),
      messageSnippet: message.substring(0, 50) + '...',
      mediaCount: mediaFiles.length,
    }));
    setPostQueue(prev => [...prev, ...newQueueItems]);

    try {
      const results = await publish(selectedPlatforms);
      

      const allSuccess = results.every((r) => r.success);
      if (allSuccess) {
        setStatus({ success: true, message: 'All posts published successfully!' });
      } else {
        const errors = results
          .filter((r) => !r.success)
          .map((r) => `${r.platform}: ${r.error || 'Unknown error'}`)
          .join('; ');
        setStatus({ success: false, message: `Some posts failed: ${errors}` });
      }

      setPostQueue(prev => prev.map(item => {
        const result = results.find(r => r.platform === item.platform);
        if (result) {
          return {
            ...item,
            status: result.success ? 'completed' : 'failed',
            resultId: result.postId,
            error: result.error,
          };
        }
        return item;
      }));

    } catch (error) {
      setStatus({ success: false, message: `Publish failed: ${error.message}` });
      setPostQueue(prev => prev.map(item => ({
        ...item,
        status: 'failed',
        error: error.message,
      })));
    } finally {
      setIsPublishing(false);
    }
  };

  return (
    // Main page container with improved padding
    <div className="min-h-screen bg-gray-50 py-10 px-4 sm:px-6 lg:px-8 font-sans">

      {/* Grid container for the three columns */}
      <div className="grid grid-cols-1 lg:grid-cols-[1fr_2fr_1fr] gap-6 max-w-8xl mx-auto items-start">
        {/* Left Column: Post Preview */}
        <div className="bg-white rounded-lg shadow-md p-6 h-full">
          <h2 className="text-xl font-semibold text-gray-800 mb-4">Post Preview</h2>
          <PostPreview
            selectedPlatforms={selectedPlatforms}
            message={message}
            mediaFiles={mediaFiles}
            youtubeConfig={youtubeConfig}
            platformsList={platformsList} // Pass the platformsList to PostPreview
          />
        </div>

        {/* Middle Column: Post Editor (now includes Platform Selection) */}
        <div className="bg-white rounded-lg shadow-md p-6 h-full">
          <PostEditor
            message={message}
            setMessage={setMessage}
            mediaFiles={mediaFiles}
            setMediaFiles={setMediaFiles}
            youtubeConfig={youtubeConfig}
            setYoutubeConfig={setYoutubeConfig}
            selectedPlatforms={selectedPlatforms}
            togglePlatform={togglePlatform}
            handlePublish={handlePublish}
            isPublishing={isPublishing}
            status={status}
          />
        </div>

        {/* Right Column: Post Queue */}
        <div className="bg-white rounded-lg shadow-md p-6 h-full">
          <h2 className="text-xl font-semibold text-gray-800 mb-4">Posting Queue</h2>
          <PostQueue postQueue={postQueue} />
        </div>
      </div>
    </div>
  );
}