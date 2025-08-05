import React, { useState, useRef, useEffect } from 'react';
import Modal from './Modal';
import PlatformSelector from './PlatformSelector';
import MiniFacebookPreview from './MiniFacebookPreview';
import MiniMastodonPreview from './MiniMastodonPreview';
import { FaCommentAlt, FaEllipsisH } from 'react-icons/fa';
import { useDraftPosts } from '../../hooks/api/useDraftPosts';
import { uploadToCloudinary } from '../../hooks/api/uploadToCloudinary';

export default function DraftsSection({ teamMembers, currentUser, workspaceId }) {
  const [showModal, setShowModal] = useState(false);
  const [selectedPlatforms, setSelectedPlatforms] = useState([]);
  const [content, setContent] = useState('');
  const [title, setTitle] = useState('');
  const [media, setMedia] = useState(null);
  const [mediaPreview, setMediaPreview] = useState(null);
  const [openCommentDraftId, setOpenCommentDraftId] = useState(null);
  const [newComment, setNewComment] = useState('');
  const [menuOpenDraftId, setMenuOpenDraftId] = useState(null);
  const [editDraft, setEditDraft] = useState(null);
  const [filterMember, setFilterMember] = useState('all');
  const [cloudMediaUrl, setCloudMediaUrl] = useState(null);
  const [mediaUploading, setMediaUploading] = useState(false);
  const [mediaUploadError, setMediaUploadError] = useState(null);
  const menuRef = useRef();

  // Use backend-powered drafts
  const { drafts, loading, error, createDraft, updateDraft, deleteDraft, publishDraft } = useDraftPosts(workspaceId);

  const handleOpenModal = () => setShowModal(true);
  const handleCloseModal = () => {
    setShowModal(false);
    setSelectedPlatforms([]);
    setContent('');
    setTitle('');
    setMedia(null);
    setMediaPreview(null);
    setEditDraft(null);
  };

  const handlePlatformToggle = (platformKey) => {
    setSelectedPlatforms((prev) =>
      prev.includes(platformKey)
        ? prev.filter((key) => key !== platformKey)
        : [...prev, platformKey]
    );
  };

  const handleMediaChange = async (e) => {
    const file = e.target.files[0];
    if (file) {
      setMedia(file);
      setMediaPreview(URL.createObjectURL(file));
      setMediaUploading(true);
      setMediaUploadError(null);
      try {
        const url = await uploadToCloudinary(file);
        setCloudMediaUrl(url);
      } catch (err) {
        setMediaUploadError('Failed to upload media.');
        setCloudMediaUrl(null);
      } finally {
        setMediaUploading(false);
      }
    } else {
      setMedia(null);
      setMediaPreview(null);
      setCloudMediaUrl(null);
    }
  };

  const handleEditDraft = (draft) => {
    setEditDraft(draft);
    setShowModal(true);
    setTitle(draft.title || '');
    setContent(draft.content || '');
    setSelectedPlatforms(draft.platforms || []);
    setMediaPreview(draft.media && draft.media[0] ? draft.media[0] : null);
    setCloudMediaUrl(draft.media && draft.media[0] ? draft.media[0] : null);
    setMedia(null);
  };

  const handleDeleteDraft = async (draftId) => {
    await deleteDraft(draftId);
  };

  const handleCreateDraft = async (e) => {
    e.preventDefault();
    if (!content.trim() || !title.trim() || selectedPlatforms.length === 0) return;
    if (editDraft) {
      await updateDraft(editDraft.id, {
        content,
        platforms: selectedPlatforms,
        media: cloudMediaUrl ? [cloudMediaUrl] : editDraft.media || [],
      });
      setEditDraft(null);
    } else {
      await createDraft({
        content,
        platforms: selectedPlatforms,
        media: cloudMediaUrl ? [cloudMediaUrl] : [],
      });
    }
    setTitle("");
    setContent("");
    setSelectedPlatforms([]);
    setMedia(null);
    setMediaPreview(null);
    setCloudMediaUrl(null);
    setShowModal(false);
  };

  useEffect(() => {
    function handleClickOutside(event) {
      if (menuRef.current && !menuRef.current.contains(event.target)) {
        setMenuOpenDraftId(null);
      }
    }
    if (menuOpenDraftId !== null) {
      document.addEventListener('mousedown', handleClickOutside);
    }
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [menuOpenDraftId]);

  // Filter drafts by selected member
  const filteredDrafts = filterMember === 'all'
    ? drafts
    : drafts.filter(d => d.author && (d.author.id === filterMember || d.author.name === filterMember));

  return (
    <section>
      <div className="mb-4">
        <label className="mr-2 font-medium text-gray-700">Filter by member:</label>
        <select
          className="border rounded px-2 py-1 text-gray-700"
          value={filterMember}
          onChange={e => setFilterMember(e.target.value)}
        >
          <option value="all">All</option>
          {teamMembers && teamMembers.map(member => (
            <option key={member.id || member.name} value={member.id || member.name}>{member.name}</option>
          ))}
        </select>
      </div>
      <div className="mb-6">
        <button
          className="py-2 px-6 bg-blue-600 text-white rounded hover:bg-blue-700 transition font-semibold flex items-center gap-2"
          onClick={handleOpenModal}
        >
          + Add Draft
        </button>
      </div>
      <Modal open={showModal} onClose={handleCloseModal}>
        <div className="p-0 md:p-8 flex flex-col md:flex-row md:space-x-8">
          {/* Form Container */}
          <div className="flex-1 bg-white rounded-2xl shadow-lg p-8">
            <form onSubmit={handleCreateDraft} className="space-y-4">
              <label className="block text-base font-semibold text-gray-800 mb-2">Platforms</label>
              <PlatformSelector selectedPlatforms={selectedPlatforms} togglePlatform={handlePlatformToggle} />
              <label className="block text-base font-semibold text-gray-800 mb-1">Title</label>
              <input
                type="text"
                className="w-full border-0 bg-gray-100 rounded-xl px-4 py-3 text-xl text-gray-800 font-semibold focus:outline-none focus:ring-2 focus:ring-blue-400 placeholder-gray-500"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="Title (e.g. Spring Campaign)"
                maxLength={80}
                required
              />
              <label className="block text-base font-semibold text-gray-800 mb-1">Content</label>
              <textarea
                className="w-full border-0 bg-gray-100 rounded-xl px-4 py-3 text-base text-gray-800 focus:outline-none focus:ring-2 focus:ring-blue-400 placeholder-gray-500"
                value={content}
                onChange={(e) => setContent(e.target.value)}
                placeholder="Write your post content here..."
                rows={5}
                required
              />
              <label className="block text-base font-semibold text-gray-800 mb-1">Media</label>
              <input type="file" accept="image/*,video/*" onChange={handleMediaChange} />
              {mediaUploading && <div className="text-blue-600 text-sm mt-1">Uploading media...</div>}
              {mediaUploadError && <div className="text-red-600 text-sm mt-1">{mediaUploadError}</div>}
              {mediaPreview && (
                <div className="mt-2">
                  {mediaPreview.match(/\.(mp4|mov|avi|mkv|wmv|flv|webm)$/i) ? (
                    <video src={mediaPreview} controls className="max-w-xs h-32 object-contain rounded-xl border" />
                  ) : (
                    <img src={mediaPreview} alt="Media Preview" className="max-w-xs h-32 object-contain rounded-xl border" />
                  )}
                </div>
              )}
              <div className="flex justify-end mt-6">
                <button
                  type="submit"
                  className="px-6 py-2 bg-blue-600 text-white rounded-lg font-semibold hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-400"
                >
                  {editDraft ? 'Update Draft' : 'Create Draft'}
                </button>
              </div>
            </form>
          </div>
          {/* Preview Container */}
          <div className="hidden md:block flex-1">
            {selectedPlatforms.includes('facebook') && (
              <MiniFacebookPreview 
                task={{
                  id: 0,
                  title: title,
                  description: content,
                  author: currentUser || { name: 'User', avatar: '/default-avatar.png' },
                  status: 'Draft',
                  assigned: [],
                  reactions: { thumbsUp: 0 },
                  comments: [],
                  photo: mediaPreview,
                  platforms: selectedPlatforms,
                }}
                onReact={() => {}}
                showReactions={false}
              />
            )}
            {selectedPlatforms.includes('mastodon') && (
              <MiniMastodonPreview 
                task={{
                  id: 0,
                  title: title,
                  description: content,
                  author: currentUser || { name: 'User', avatar: '/default-avatar.png' },
                  status: 'Draft',
                  assigned: [],
                  reactions: { thumbsUp: 0 },
                  comments: [],
                  photo: mediaPreview,
                  platforms: selectedPlatforms,
                }}
                onReact={() => {}}
                showReactions={false}
              />
            )}
          </div>
        </div>
      </Modal>
      {/* Drafts List */}
      <div className="space-y-6">
        {loading && <div className="text-gray-500">Loading drafts...</div>}
        {error && <div className="text-red-500">Error: {error}</div>}
        {filteredDrafts.map((draft) => (
          <div key={draft.id} className="mb-6">
            {draft.platforms && draft.platforms.includes('facebook') && (
              <div className="max-w-md mx-auto">
                <MiniFacebookPreview
                  task={{
                    ...draft,
                    description: draft.content,
                    photo: draft.media && draft.media[0],
                    author: draft.author || currentUser || { name: 'Unknown', avatar: '/default-avatar.png' },
                    reactions: draft.reactions || { thumbsUp: 0 },
                    comments: draft.comments || [],
                  }}
                  showReactions={true}
                  showTitle={false}
                  onReact={() => {}}
                  fullWidth={true}
                  onEdit={() => handleEditDraft(draft)}
                  onDelete={() => handleDeleteDraft(draft.id)}
                  onPost={() => publishDraft(draft.id)}
                />
              </div>
            )}
            {draft.platforms && draft.platforms.includes('mastodon') && (
              <div className="max-w-md mx-auto">
                <MiniMastodonPreview
                  task={{
                    ...draft,
                    description: draft.content,
                    photo: draft.media && draft.media[0],
                    author: draft.author || currentUser || { name: 'Unknown', avatar: '/default-avatar.png' },
                    reactions: draft.reactions || { thumbsUp: 0 },
                    comments: draft.comments || [],
                  }}
                  showReactions={true}
                  showTitle={false}
                  onReact={() => {}}
                  fullWidth={true}
                  onEdit={() => handleEditDraft(draft)}
                  onDelete={() => handleDeleteDraft(draft.id)}
                  onPost={() => publishDraft(draft.id)}
                />
              </div>
            )}
          </div>
        ))}
      </div>
    </section>
  );
} 