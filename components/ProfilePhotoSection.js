import Image from 'next/image';
import { FaCamera } from 'react-icons/fa';

export default function ProfilePhotoSection({ profileImage, imagePreview, onImageChange, error }) {
  return (
    <section className="space-y-3">
      <h2 className="text-xl font-semibold text-gray-800 border-b pb-2 mb-2">Profile Photo</h2>
      <div className="flex flex-col items-center space-y-4">
        <div className="relative group">
          <div className="w-28 h-28 rounded-full overflow-hidden bg-gray-200 shadow ring-2 ring-indigo-100 transition-all duration-200 group-hover:ring-indigo-300">
            <Image
              src={imagePreview || profileImage || '/default-avatar.png'}
              alt="Profile"
              width={112}
              height={112}
              className="object-cover w-full h-full"
            />
          </div>
          <label className="absolute bottom-1 right-1 bg-indigo-600 p-2 rounded-full cursor-pointer hover:bg-indigo-700 transition-colors duration-200 transform group-hover:scale-105 shadow">
            <FaCamera className="text-white text-base" />
            <input type="file" className="hidden" accept="image/*" onChange={onImageChange} />
          </label>
        </div>
        {error && (
          <p className="text-red-600 text-sm mt-2">{error}</p>
        )}
      </div>
    </section>
  );
}
