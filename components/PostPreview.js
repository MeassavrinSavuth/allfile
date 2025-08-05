'use client';

import {
  FaEllipsisH,
  FaThumbsUp,
  FaCommentAlt,
  FaShare,
  FaInstagram,
  FaYoutube,
  FaTwitter,
  FaMastodon,
} from 'react-icons/fa'; // Import icons

const mockProfile = { // Mock data for previews
  name: "Your Page Name",
  avatar: "/default-avatar.png",
  timestamp: "Just now",
};

const isVideoFile = (url) => /\.(mp4|mov|avi|mkv|wmv|flv|webm)$/i.test(url) || url.includes('video');

const FacebookPreview = ({ message, mediaFiles }) => (
  <div className="bg-gray-900 text-white rounded-lg shadow-lg overflow-hidden">
    {/* Header */}
    <div className="flex items-center p-4">
      <img
        src={mockProfile.avatar}
        alt="Profile"
        className="w-10 h-10 rounded-full object-cover mr-3 flex-shrink-0"
        onError={(e) => (e.target.src = '/default-avatar.png')}
      />
      <div className="flex-grow">
        <p className="font-semibold text-sm">{mockProfile.name}</p>
        <p className="text-xs text-gray-400">{mockProfile.timestamp}</p>
      </div>
      <FaEllipsisH className="text-gray-400 text-lg cursor-pointer" />
    </div>

    {/* Post Message */}
    {message && <p className="px-4 pb-3 text-sm whitespace-pre-line">{message}</p>}

    {/* Media */}
    {mediaFiles.length > 0 && (
      <div className="relative w-full overflow-hidden">
        {mediaFiles.map((url, i) => i === 0 && (
          isVideoFile(url) ? (
            <video key={i} src={url} controls className="w-full h-auto object-cover max-h-[400px]" />
          ) : (
            <img key={i} src={url} alt={`Post media ${i + 1}`} className="w-full h-auto object-cover max-h-[400px]" />
          )
        ))}
      </div>
    )}

    {/* Footer Actions */}
    <div className="flex justify-around items-center py-3 border-t border-gray-700 text-gray-400 text-sm">
      <button className="flex items-center space-x-2 p-2 rounded-lg hover:bg-gray-800 transition"><FaThumbsUp /><span>Like</span></button>
      <button className="flex items-center space-x-2 p-2 rounded-lg hover:bg-gray-800 transition"><FaCommentAlt /><span>Comment</span></button>
      <button className="flex items-center space-x-2 p-2 rounded-lg hover:bg-gray-800 transition"><FaShare /><span>Share</span></button>
    </div>
  </div>
);

const InstagramPreview = ({ message, mediaFiles }) => (
  <div className="bg-white border border-gray-300 rounded-lg shadow-sm overflow-hidden">
    {/* Header */}
    <div className="flex items-center p-3 border-b border-gray-200">
      <FaInstagram className="text-2xl mr-3 text-instagram" />
      <p className="text-sm font-semibold">Your Instagram</p>
    </div>

    {/* Media */}
    {mediaFiles.length > 0 && (
      <div className="relative w-full overflow-hidden">
        {mediaFiles.map((url, i) => i === 0 && (
          isVideoFile(url) ? (
            <video key={i} src={url} className="w-full h-auto object-cover max-h-[400px]" controls />
          ) : (
            <img key={i} src={url} alt={`Post media ${i + 1}`} className="w-full h-auto object-cover max-h-[400px]" />
          )
        ))}
      </div>
    )}

    {/* Message */}
    {message && <p className="px-3 py-2 text-sm">{message}</p>}
  </div>
);

const YoutubePreview = ({ message, mediaFiles, youtubeConfig }) => (
  <div className="bg-white border border-gray-300 rounded-lg shadow-sm overflow-hidden">
    {/* Header */}
    <div className="flex items-center p-3 border-b border-gray-200">
      <FaYoutube className="text-2xl mr-3 text-youtube" />
      <p className="text-sm font-semibold">Your Channel</p>
    </div>

    {/* Media */}
    {mediaFiles.length > 0 && (
      <div className="relative w-full overflow-hidden">
        {mediaFiles.map((url, i) => i === 0 && (
          isVideoFile(url) ? (
            <video key={i} src={url} className="w-full h-auto object-cover max-h-[400px]" controls />
          ) : null // YouTube is primarily video
        ))}
      </div>
    )}

    {/* Title and Description */}
    {youtubeConfig.title && <p className="px-3 pt-2 text-sm font-semibold">{youtubeConfig.title}</p>}
    {youtubeConfig.description && <p className="px-3 pb-2 text-sm">{youtubeConfig.description}</p>}
    {message && <p className="px-3 pb-2 text-sm">{message}</p>}
  </div>
);

const TwitterPreview = ({ message, mediaFiles }) => (
  <div className="bg-white border border-gray-300 rounded-lg shadow-sm overflow-hidden">
    {/* Header */}
    <div className="flex items-center p-3 border-b border-gray-200">
      <FaTwitter className="text-2xl mr-3 text-twitter" />
      <p className="text-sm font-semibold">Your Twitter</p>
    </div>

    {/* Message */}
    {message && <p className="px-3 py-2 text-sm whitespace-pre-line">{message}</p>}

    {/* Media */}
    {mediaFiles.length > 0 && (
      <div className="flex flex-wrap gap-2 p-3">
        {mediaFiles.map((url, i) => (
          isVideoFile(url) ? (
            <video key={i} src={url} className="w-24 h-24 object-cover rounded" />
          ) : (
            <img key={i} src={url} alt={`media-${i}`} className="w-24 h-24 object-cover rounded" />
          )
        ))}
      </div>
    )}
  </div>
);

const MastodonPreview = ({ message, mediaFiles }) => (
  <div className="bg-white border border-gray-300 rounded-lg shadow-sm overflow-hidden">
    {/* Header */}
    <div className="flex items-center p-3 border-b border-gray-200">
      <FaMastodon className="text-2xl mr-3 text-mastodon" />
      <p className="text-sm font-semibold">Your Mastodon</p>
    </div>

    {/* Message */}
    {message && <p className="px-3 py-2 text-sm whitespace-pre-line">{message}</p>}

    {/* Media */}
    {mediaFiles.length > 0 && (
      <div className="flex flex-wrap gap-2 p-3">
        {mediaFiles.map((url, i) => (
          isVideoFile(url) ? (
            <video key={i} src={url} className="w-24 h-24 object-cover rounded" />
          ) : (
            <img key={i} src={url} alt={`media-${i}`} className="w-24 h-24 object-cover rounded" />
          )
        ))}
      </div>
    )}
  </div>
);

export default function PostPreview({ selectedPlatforms, message, mediaFiles, youtubeConfig }) {
  return (
    <div className="space-y-6">
      {selectedPlatforms.includes('facebook') && <FacebookPreview message={message} mediaFiles={mediaFiles} />}
      {selectedPlatforms.includes('instagram') && <InstagramPreview message={message} mediaFiles={mediaFiles} />}
      {selectedPlatforms.includes('youtube') && <YoutubePreview message={message} mediaFiles={mediaFiles} youtubeConfig={youtubeConfig} />}
      {selectedPlatforms.includes('twitter') && <TwitterPreview message={message} mediaFiles={mediaFiles} />}
      {selectedPlatforms.includes('mastodon') && <MastodonPreview message={message} mediaFiles={mediaFiles} />}

      {selectedPlatforms.length > 0 && (
        <div className="pt-4 border-t border-gray-300">
          <h3 className="font-medium text-gray-800 mb-2">Platforms Selected</h3>
          <ul className="flex flex-wrap gap-2">
            {selectedPlatforms.map(platform => (
              <li key={platform} className="bg-blue-100 text-blue-800 text-xs font-medium px-2.5 py-0.5 rounded-full capitalize">
                {platform}
              </li>
            ))}
          </ul>
        </div>
      )}

      {selectedPlatforms.length === 0 && mediaFiles.length === 0 && !message && (
        <p className="text-gray-500 italic">Start composing your post to see a preview here.</p>
      )}
    </div>
  );
}