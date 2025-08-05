'use client';
import React, { useState, useRef, Fragment } from 'react';
import { FaPlus, FaTimes, FaCloudUploadAlt, FaTag, FaTrash, FaEdit } from 'react-icons/fa';
import { useMedia } from '../../hooks/api/useMedia';

const FILTERS = [
  { key: 'all', label: 'All' },
  { key: 'image', label: 'Photos' },
  { key: 'video', label: 'Videos' },
];

function AddMediaCard({ onUpload, tags, onCancel }) {
  const [file, setFile] = useState(null);
  const [filePreview, setFilePreview] = useState(null);
  const [tagInput, setTagInput] = useState('');
  const [selectedTags, setSelectedTags] = useState([]);
  const [error, setError] = useState('');
  const [dragActive, setDragActive] = useState(false);
  const fileInputRef = useRef(null);

  const handleFileChange = (e) => {
    const f = e.target.files[0];
    setFile(f);
    setError('');
    if (f) {
      setFilePreview(URL.createObjectURL(f));
    } else {
      setFilePreview(null);
    }
  };
  const handleAddTag = () => {
    const tag = tagInput.trim().toLowerCase();
    if (tag && !selectedTags.includes(tag)) {
      setSelectedTags([...selectedTags, tag]);
    }
    setTagInput('');
  };
  const handleRemoveTag = (tag) => {
    setSelectedTags(selectedTags.filter(t => t !== tag));
  };
  const handleTagInputKeyDown = (e) => {
    if (e.key === 'Enter' || e.key === ',') {
      e.preventDefault();
      handleAddTag();
    } else if (e.key === 'Backspace' && !tagInput && selectedTags.length > 0) {
      setSelectedTags(selectedTags.slice(0, -1));
    }
  };
  const handleTagDropdown = (tag) => {
    if (!selectedTags.includes(tag)) {
      setSelectedTags([...selectedTags, tag]);
    }
  };
  const handleDrop = (e) => {
    e.preventDefault();
    setDragActive(false);
    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      handleFileChange({ target: { files: e.dataTransfer.files } });
    }
  };
  const handleDragOver = (e) => {
    e.preventDefault();
    setDragActive(true);
  };
  const handleDragLeave = (e) => {
    e.preventDefault();
    setDragActive(false);
  };
  const handleSubmit = (e) => {
    e.preventDefault();
    if (!file) {
      setError('Please select a file.');
      return;
    }
    let tagsToSubmit = selectedTags;
    if (tagInput.trim()) {
      const newTag = tagInput.trim().toLowerCase();
      if (!tagsToSubmit.includes(newTag)) {
        tagsToSubmit = [...tagsToSubmit, newTag];
      }
    }
    onUpload(file, tagsToSubmit);
    setFile(null);
    setFilePreview(null);
    setSelectedTags([]);
    setTagInput('');
    setError('');
  };
  // Filter tag suggestions
  const tagSuggestions = tags.filter(t => !selectedTags.includes(t) && t.includes(tagInput.toLowerCase()) && tagInput);

  return (
    <div className="relative animate-fade-in mb-8 w-full max-w-2xl mx-auto bg-white rounded-3xl shadow-2xl p-10 border border-gray-100 flex flex-col gap-6 transition-all duration-300">
      <button type="button" className="absolute top-4 right-4 text-gray-400 hover:text-gray-700 text-2xl" onClick={onCancel} aria-label="Close">
        <FaTimes />
      </button>
      <form onSubmit={handleSubmit}>
        <div className="flex flex-col md:flex-row gap-8">
          {/* Drag-and-drop area */}
          <div className="flex-1 flex flex-col items-center justify-center">
            <label
              htmlFor="media-upload-input"
              className={`w-full h-44 flex flex-col items-center justify-center border-2 border-dashed rounded-2xl cursor-pointer transition-all duration-200 ${dragActive ? 'border-blue-400 bg-blue-50' : 'border-gray-200 bg-gray-50 hover:border-blue-300'}`}
              onDrop={handleDrop}
              onDragOver={handleDragOver}
              onDragLeave={handleDragLeave}
            >
              {filePreview ? (
                file && file.type && file.type.startsWith('video') ? (
                  <video src={filePreview} controls className="mx-auto max-w-xs h-32 object-contain rounded-xl border" />
                ) : (
                  <img src={filePreview} alt="Preview" className="mx-auto max-w-xs h-32 object-contain rounded-xl border" />
                )
              ) : (
                <>
                  <FaCloudUploadAlt className="text-4xl text-blue-400 mb-2" />
                  <span className="text-lg font-medium text-gray-700">Click or drag to upload</span>
                  <span className="text-xs text-gray-400 mt-1">PNG, JPG, GIF, MP4, MOV, etc.</span>
                </>
              )}
              <input
                id="media-upload-input"
                type="file"
                accept="image/*,video/*"
                onChange={handleFileChange}
                className="hidden"
                ref={fileInputRef}
              />
            </label>
          </div>
          {/* Tag input */}
          <div className="flex-1 flex flex-col justify-between">
            <label className="block text-base font-semibold text-gray-800 mb-2 flex items-center gap-2"><FaTag className="text-blue-400" /> Tags</label>
            <div className="flex flex-wrap gap-2 mb-3 min-h-[40px]">
              {selectedTags.map(tag => (
                <span key={tag} className="flex items-center px-3 py-1 bg-blue-100 text-blue-700 rounded-full text-sm font-semibold border border-blue-200 mr-1">
                  {tag}
                  <button type="button" className="ml-2 text-blue-400 hover:text-blue-700" onClick={() => handleRemoveTag(tag)}>&times;</button>
                </span>
              ))}
            </div>
            <input
              type="text"
              className="bg-gray-50 border border-gray-300 rounded px-2 py-1 text-base text-gray-900 min-w-[80px] placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-200 focus:border-blue-400 transition"
              placeholder={selectedTags.length === 0 ? 'Add tag...' : ''}
              value={tagInput}
              onChange={e => setTagInput(e.target.value)}
              onKeyDown={handleTagInputKeyDown}
              autoFocus={selectedTags.length === 0}
            />
            {/* Tag suggestions dropdown */}
            {tagSuggestions.length > 0 && (
              <div className="absolute mt-1 bg-white border border-gray-200 rounded shadow z-10 w-48">
                {tagSuggestions.map(tag => (
                  <button key={tag} type="button" className="block w-full text-left px-4 py-2 text-sm hover:bg-blue-50" onClick={() => handleTagDropdown(tag)}>{tag}</button>
                ))}
              </div>
            )}
          </div>
        </div>
        {error && <div className="text-red-500 text-sm mb-2">{error}</div>}
        <div className="flex gap-2 mt-8 justify-end">
          <button type="button" className="px-5 py-2 bg-gray-100 text-gray-700 rounded-full font-semibold hover:bg-gray-200 text-lg" onClick={onCancel}>Cancel</button>
          <button type="submit" className="px-6 py-2 bg-blue-600 text-white rounded-full font-semibold hover:bg-blue-700 text-lg flex items-center gap-2"><FaCloudUploadAlt /> Upload</button>
        </div>
      </form>
    </div>
  );
}

export default function MediaLibraryGrid({ workspaceId }) {
  const [selectedType, setSelectedType] = useState('all');
  const [search, setSearch] = useState('');
  const [selectedTag, setSelectedTag] = useState('');
  const fileInputRef = useRef(null);
  const [filterDropdownOpen, setFilterDropdownOpen] = useState(false);
  const [addCardOpen, setAddCardOpen] = useState(false);
  
  const { media, loading, error, uploadMedia, deleteMedia, updateMediaTags } = useMedia(workspaceId);

  // Dynamically compute unique tags from the media list
  const tags = Array.from(
    new Set(
      (media || []).flatMap(m => Array.isArray(m.tags) ? m.tags : [])
    )
  );

  // Deduplicate media by id
  const dedupedMedia = Array.from(
    new Map((media || []).map(m => [m.id, m])).values()
  );

  const handleUploadClick = () => setAddCardOpen((open) => !open);
  const handleCancelAdd = () => setAddCardOpen(false);

  const handleDeleteMedia = async (mediaId) => {
    if (window.confirm('Are you sure you want to delete this media?')) {
      try {
        await deleteMedia(mediaId);
      } catch (error) {
        console.error('Failed to delete media:', error);
        // You could add a toast notification here
      }
    }
  };

  const handleUpload = async (file, newTags) => {
    try {
      await uploadMedia(file, newTags);
    setAddCardOpen(false);
    } catch (error) {
      console.error('Failed to upload media:', error);
      // You could add a toast notification here
    }
  };

  const filteredMedia = dedupedMedia.filter((m) => {
    const matchesType = selectedType === 'all' || m.file_type === selectedType;
    const matchesTag = !selectedTag || m.tags.includes(selectedTag);
    const matchesSearch = !search || m.original_name.toLowerCase().includes(search.toLowerCase());
    return matchesType && matchesTag && matchesSearch;
  });

  return (
    <div className="w-full">
      <div className="mb-6">
        <button
          className="py-2 px-6 bg-blue-600 text-white rounded hover:bg-blue-700 transition font-semibold flex items-center gap-2"
          onClick={handleUploadClick}
        >
          <FaPlus className="text-base" /> Add Media
        </button>
      </div>
      {addCardOpen && (
        <AddMediaCard onUpload={handleUpload} tags={tags} onCancel={handleCancelAdd} />
      )}
      {/* Modern Toolbar */}
      <div className="flex flex-col md:flex-row md:items-center gap-3 md:gap-4 mb-8 bg-white rounded-xl shadow p-4 border border-gray-100">
        {/* Filter by type */}
        <div className="flex gap-2">
          {FILTERS.map(tab => (
            <button
              key={tab.key}
              className={`px-3 py-1 rounded text-sm font-semibold transition ${selectedType === tab.key ? 'bg-blue-100 text-blue-700' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'}`}
              onClick={() => setSelectedType(tab.key)}
            >
              {tab.label}
            </button>
          ))}
        </div>
        {/* Dropdown Filter for tags */}
        <div className="relative">
          <button
            className="px-4 py-1.5 rounded text-sm font-semibold bg-gray-100 text-gray-700 hover:bg-gray-200 flex items-center gap-2 border border-gray-200 shadow-sm"
            onClick={() => setFilterDropdownOpen((open) => !open)}
            aria-haspopup="listbox"
            aria-expanded={filterDropdownOpen}
          >
            <span className="truncate max-w-[80px]">{selectedTag ? tags.find(t => t === selectedTag) : 'Filter'}</span>
            <svg className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" d="M19 9l-7 7-7-7" /></svg>
          </button>
          {filterDropdownOpen && (
            <div className="absolute left-0 mt-2 w-36 bg-white border border-gray-200 rounded-lg shadow-lg z-20">
              <button
                className={`block w-full text-left px-4 py-2 text-sm rounded hover:bg-blue-50 transition ${selectedTag === '' ? 'font-bold text-blue-700' : 'text-gray-700'}`}
                onClick={() => { setSelectedTag(''); setFilterDropdownOpen(false); }}
              >
                All
              </button>
              {tags.map(tag => (
                <button
                  key={tag}
                  className={`block w-full text-left px-4 py-2 text-sm rounded hover:bg-blue-50 transition ${selectedTag === tag ? 'font-bold text-blue-700' : 'text-gray-700'}`}
                  onClick={() => { setSelectedTag(tag); setFilterDropdownOpen(false); }}
                >
                  {tag.charAt(0).toUpperCase() + tag.slice(1)}
                </button>
              ))}
            </div>
          )}
        </div>
        {/* Search */}
        <input
          type="text"
          className="border border-gray-300 rounded px-3 py-1.5 text-sm flex-1 min-w-[180px] focus:ring-2 focus:ring-blue-200 focus:border-blue-400 transition"
          placeholder="Search media..."
          value={search}
          onChange={e => setSearch(e.target.value)}
        />
      </div>
      {/* Media Grid */}
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-8">
        {loading && (
          <div className="col-span-full text-center text-gray-500 py-12">Loading media...</div>
        )}
        {error && (
          <div className="col-span-full text-center text-red-500 py-12">Error loading media: {error}</div>
        )}
        {!loading && !error && filteredMedia.length === 0 && (
          <div className="col-span-full text-center text-gray-400 py-12">No media found.</div>
        )}
        {filteredMedia.map(media => (
          <div key={media.id} className="bg-white rounded-2xl shadow-lg p-4 flex flex-col items-center border border-gray-100 hover:border-blue-400 hover:shadow-xl transition group relative">
            <div className="w-full h-40 flex items-center justify-center bg-gray-100 rounded-xl mb-3 overflow-hidden">
              {media.file_type === 'image' ? (
                <img src={media.file_url} alt={media.original_name} className="max-h-full max-w-full object-contain" />
              ) : (
                <video src={media.file_url} controls className="max-h-full max-w-full object-contain" />
              )}
            </div>
            
            {/* Action buttons */}
            <div className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity duration-200 flex gap-1">
              <button
                onClick={() => handleDeleteMedia(media.id)}
                className="p-1 bg-red-500 text-white rounded-full hover:bg-red-600 transition"
                title="Delete media"
              >
                <FaTrash className="text-xs" />
              </button>
            </div>
            
            <div className="w-full flex flex-col items-start">
              <span className="font-semibold text-sm text-gray-800 truncate max-w-[140px]" title={media.original_name}>{media.original_name}</span>
              <div className="flex flex-wrap gap-1 mt-1 mb-2">
                {media.tags.map(tag => (
                  <span key={tag} className="px-2 py-0.5 rounded-full bg-blue-50 text-blue-700 text-xs font-semibold border border-blue-100">{tag}</span>
                ))}
              </div>
              <span className="text-xs text-gray-400">by {media.uploader_name} • {new Date(media.created_at).toLocaleDateString()}</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
} 