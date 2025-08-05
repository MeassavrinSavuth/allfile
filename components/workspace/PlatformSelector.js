'use client';

import { FaFacebook, FaInstagram, FaYoutube, FaTwitter } from 'react-icons/fa';
import { SiMastodon } from 'react-icons/si';

const platformIcons = {
  facebook: FaFacebook,
  instagram: FaInstagram,
  youtube: FaYoutube,
  twitter: FaTwitter,
  mastodon: SiMastodon,
};

const platformColors = {
  facebook: 'bg-[#1877F3] text-white',
  instagram: '',
  youtube: 'bg-[#FF0000] text-white',
  twitter: 'bg-[#1DA1F2] text-white',
  mastodon: 'bg-[#6364FF] text-white',
};

const platformsList = ['facebook', 'instagram', 'youtube', 'twitter', 'mastodon'];

export default function PlatformSelector({ selectedPlatforms, togglePlatform }) {
  const isSelected = (platform) => selectedPlatforms.includes(platform);

  return (
    <div className="flex items-center justify-center space-x-4 mb-4">
      {platformsList.map((platform) => {
        const Icon = platformIcons[platform];
        const selected = isSelected(platform);
        const isInstagram = platform === 'instagram';
        const buttonClass = selected
          ? isInstagram
            ? 'text-white'
            : platformColors[platform]
          : 'bg-gray-100 text-gray-400 hover:bg-gray-200';
        const style = isInstagram && selected
          ? { background: 'radial-gradient(circle at 30% 107%, #fdf497 0%, #fdf497 5%, #fd5949 45%, #d6249f 60%, #285AEB 90%)', color: 'white' }
          : {};
        return (
          <button
            key={platform}
            type="button"
            onClick={() => togglePlatform(platform)}
            className={`rounded-full flex items-center justify-center w-16 h-16 text-3xl transition-all duration-200 ${buttonClass}`}
            title={platform.charAt(0).toUpperCase() + platform.slice(1)}
            style={style}
          >
            <Icon size={32} />
          </button>
        );
      })}
    </div>
  );
}
