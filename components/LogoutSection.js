import { FaSignOutAlt } from 'react-icons/fa';
import LogoutConfirm from './LogoutConfirm';

export default function LogoutSection({ showLogoutConfirm, setShowLogoutConfirm }) {
  return (
    <div className="pt-6 border-t border-gray-200 mt-8">
      <button
        onClick={() => setShowLogoutConfirm(true)}
        className="w-full sm:w-auto px-5 py-2.5 bg-red-500 text-white font-medium rounded-lg hover:bg-red-700 transition-all duration-200 flex items-center justify-center space-x-2 shadow-sm hover:shadow transform hover:-translate-y-0.5 text-base"
      >
        <FaSignOutAlt className="text-sm" />
        <span>Logout</span>
      </button>

      {showLogoutConfirm && <LogoutConfirm onCancel={() => setShowLogoutConfirm(false)} />}
    </div>
  );
}
