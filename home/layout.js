'use client';

import { useState, useEffect } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { useUser } from '../hooks/auth/useUser';
import {
  FaHome, FaChartBar, FaCommentDots, FaFolder, FaUser, FaPlus, FaBriefcase,
  FaAngleLeft, FaAngleRight
} from 'react-icons/fa';

import { Pacifico } from 'next/font/google';

const pacifico = Pacifico({
  weight: '400',
  subsets: ['latin'],
});

export default function DashboardLayout({ children }) {
  const router = useRouter();
  const pathname = usePathname();
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [activeTab, setActiveTab] = useState(null);
  const { profileData, isLoading } = useUser();

  const defaultProfilePic = '/default-avatar.png';
  const imageUrl =
  profileData?.profileImage && profileData.profileImage.startsWith('http')
    ? profileData.profileImage
    : defaultProfilePic;

  // Redirect to login if no tokens
  useEffect(() => {
    if (typeof window == 'undefined') return;
    const accessToken = localStorage.getItem('accessToken');
    const refreshToken = localStorage.getItem('refreshToken');
    if (!accessToken || !refreshToken) {
      router.push('/login');
    }
  }, [router]);

  // Set active tab based on path (reactive on path change)
  useEffect(() => {
    if (pathname === '/home/profile') {
      setActiveTab(null);
    } else if (pathname.startsWith('/home/dashboard')) {
      setActiveTab('Home');
    } else if (pathname.startsWith('/home/create-post')) {
      setActiveTab('Create Post');
    } else if (pathname.startsWith('/home/analytics')) {
      setActiveTab('Analytics');
    } else if (pathname.startsWith('/home/manage-comments')) {
      setActiveTab('Manage Comments');
    } else if (pathname.startsWith('/home/posts-folder')) {
      setActiveTab('Posts Folder');
    } else if (pathname.startsWith('/home/manage-accounts')) {
      setActiveTab('Manage Account');
    } else if (pathname.startsWith('/home/workspace')) {
      setActiveTab('Workspace');
    }
    else {
      setActiveTab(null); // fallback
    }
  }, [pathname]);

  const handleNavClick = (label, path) => {
    setActiveTab(label);
    router.push(path);
  };

  return (
    <div className="flex h-screen bg-gray-50 font-sans">
      <aside
        className={`transition-all duration-300 ease-in-out bg-white shadow-md p-4 flex flex-col justify-between border-r border-gray-200 ${
          sidebarOpen ? 'w-64' : 'w-20'
        }`}
      >
        {/* Logo & Toggle */}
        <div>
          <div className="flex items-center justify-between mb-6">
            {sidebarOpen && (
              <h1 className={`${pacifico.className} text-2xl ml-4 font-bold text-indigo-500`}>
                SocialSync
              </h1>
            )}
            <button
              onClick={() => setSidebarOpen(!sidebarOpen)}
              className="flex items-center justify-center w-10 h-10 bg-indigo-50 hover:bg-indigo-100 rounded-lg transition"
              aria-label="Toggle Sidebar"
            >
              {sidebarOpen ? (
                <FaAngleLeft className="text-xl text-indigo-600" />
              ) : (
                <FaAngleRight className="text-xl text-indigo-600" />
              )}
            </button>
          </div>
          
          {/* Sidebar Items */}
          <nav className="space-y-1">
            <NavItem icon={<FaHome />} label="Home" open={sidebarOpen} active={activeTab === 'Home'} onClick={() => handleNavClick('Home', '/home/dashboard')} />
            <NavItem icon={<FaPlus />} label="Create Post" open={sidebarOpen} active={activeTab === 'Create Post'} onClick={() => handleNavClick('Create Post', '/home/create-post')} />
            <NavItem icon={<FaChartBar />} label="Analytics" open={sidebarOpen} active={activeTab === 'Analytics'} onClick={() => handleNavClick('Analytics', '/home/analytics')} />
            <NavItem icon={<FaCommentDots />} label="Manage Comments" open={sidebarOpen} active={activeTab === 'Manage Comments'} onClick={() => handleNavClick('Manage Comments', '/home/manage-comments')} />
            <NavItem icon={<FaFolder />} label="Posts Folder" open={sidebarOpen} active={activeTab === 'Posts Folder'} onClick={() => handleNavClick('Posts Folder', '/home/posts-folder')} />
            <NavItem icon={<FaUser />} label="Manage Account" open={sidebarOpen} active={activeTab === 'Manage Account'} onClick={() => handleNavClick('Manage Account', '/home/manage-accounts')} />
            
            {/* Horizontal Rule for Separation */}
            {sidebarOpen && <hr className="my-4 border-gray-200" />} {/* Show only when sidebar is open */}
            {!sidebarOpen && <div className="py-2"></div>} {/* Add vertical space when sidebar is collapsed */}

            <NavItem icon={<FaBriefcase />} label="Workspace" open={sidebarOpen} active={activeTab === 'Workspace'} onClick={() => handleNavClick('Workspace', '/home/workspace')} />
          </nav>
        </div>

        {/* Profile Footer */}
        <div
          className="flex items-center gap-3 border-t pt-4 cursor-pointer hover:bg-gray-100 rounded-xl px-2 transition overflow-hidden"
          onClick={() => router.push('/home/profile')}
        >
          <img
            src={imageUrl}
            alt="Profile"
            className="w-10 h-10 rounded-full object-cover border border-gray-300 shadow-sm flex-shrink-0"
            onError={(e) => (e.target.src = defaultProfilePic)}
          />

          {sidebarOpen && !isLoading && (
            <div className="min-w-0">
              <p className="text-sm font-medium text-gray-900 truncate">{profileData?.name || 'Unknown'}</p>
              <p className="text-xs text-gray-500 truncate max-w-[12rem]">{profileData?.email || ''}</p>
            </div>
          )}
        </div>

      </aside>

      <main className="flex-1 p-6 overflow-y-auto bg-white rounded-l-2xl shadow-inner">{children}</main>
    </div>
  );
}

// Tooltip-enhanced sidebar item
function NavItem({ icon, label, open, active, onClick }) {
  return (
    <div
      className={`relative flex items-center gap-4 px-3 py-3 rounded-xl cursor-pointer transition-all group
        ${active ? 'bg-indigo-100 text-indigo-700 font-semibold' : 'text-gray-800 hover:bg-indigo-50 hover:text-indigo-600'}
      `}
      onClick={onClick}
    >
      <div className="relative flex items-center justify-center w-8 h-8 text-[18px]">
        {icon}
        {!open && (
          <div className="absolute left-full ml-2 hidden group-hover:flex px-2 py-1 bg-black text-white text-xs rounded shadow-md z-10 whitespace-nowrap">
            {label}
          </div>
        )}
      </div>
      {open && <span className="text-base select-none">{label}</span>}
    </div>
  );
}