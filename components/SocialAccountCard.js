import Image from 'next/image';
import { FaCheckCircle, FaTimesCircle } from 'react-icons/fa';

const DEFAULT_AVATAR = '/default-avatar.png'; // Put a default image in your public folder

export default function SocialAccountCard({
  platform,
  IconComponent,
  connected,
  userProfilePic,
  accountName,
  onConnect,
}) {
  const getPlatformColorClass = (platformName) => {
    switch (platformName) {
      case 'Facebook':
        return 'bg-[#3b5998]';
      case 'Instagram':
        return 'bg-gradient-to-r from-pink-500 to-purple-600';
      case 'YouTube':
        return 'bg-red-600';
      case 'TikTok':
        return 'bg-black';
      case 'Twitter (X)':
        return 'bg-black';
      case 'Mastodon':
        return 'bg-[#6364FF]'
      case 'Threads': // Changed to black background
        return 'bg-black';
      case 'Telegram':
        return 'bg-[#0088CC]';
      default:
        return 'bg-gray-800';
    }
  };

  const iconBgClass = getPlatformColorClass(platform);

  // Use fallback if null or empty string
  const validProfilePic = userProfilePic && userProfilePic !== 'null' ? userProfilePic : DEFAULT_AVATAR;

  return (
    <div className="relative w-full max-w-sm p-8 rounded-3xl shadow-2xl bg-white transition-all transform hover:scale-105 duration-300">
      {/* Status Icon */}
      <div className="absolute top-4 right-4 text-gray-400">
        {connected ? (
          <FaCheckCircle className="text-purple-500" size={22} />
        ) : (
          <FaTimesCircle className="text-gray-400" size={22} />
        )}
      </div>

      {connected ? (
        <div className="flex flex-col items-center justify-center mb-6">
          <div className="flex items-center justify-center gap-4 mb-3">
            {/* Platform Icon */}
            <div className={`w-20 h-20 rounded-2xl flex items-center justify-center shadow-md text-white text-4xl ${iconBgClass}`}>
              {IconComponent && <IconComponent size={28} />}
            </div>

            {/* Emoji Link Icon */}
            <div className="text-2xl text-gray-400">🔗</div>

            {/* Profile Picture */}
            <div className="w-20 h-20 rounded-full overflow-hidden shadow-lg">
              <Image
                src={validProfilePic}
                alt={`${platform} Profile`}
                width={80}
                height={80}
                className="object-cover"
                priority={true}
                quality={100}
              />
            </div>
          </div>

          {/* Account Name (already includes @username from backend) */}
          {accountName && (
            <p className="text-sm text-gray-500 font-medium mt-1">{accountName}</p>
          )}
        </div>
      ) : (
        <div className="w-20 h-20 rounded-2xl flex items-center justify-center mx-auto mb-6 shadow-md text-white text-4xl">
          <div className={`w-full h-full flex items-center justify-center rounded-2xl ${iconBgClass}`}>
            {IconComponent && <IconComponent size={32} />}
          </div>
        </div>
      )}

      <h3 className="text-center text-gray-800 font-semibold text-2xl mb-6">{platform}</h3>

      <button
        onClick={onConnect}
        className={`w-full py-3 rounded-xl text-white font-semibold text-lg transition-all ${
          connected
            ? 'bg-red-400 text-red-500 border border-red-200 cursor-pointer hover:bg-red-500'
            : 'bg-gradient-to-r from-purple-500 to-pink-500 hover:from-purple-600 hover:to-pink-600'
        }`}
      >
        {connected ? 'Disconnect' : 'Connect'}
      </button>
    </div>
  );
}