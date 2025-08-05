import { FaSave } from 'react-icons/fa';

export default function NameSection({ name, onNameChange, onSave }) {
  return (
    <section className="space-y-3">
      <h2 className="text-xl font-semibold text-gray-800 border-b pb-2 mb-2">Name</h2>
      <div className="flex flex-col sm:flex-row space-y-3 sm:space-y-0 sm:space-x-4 items-end">
        <div className="flex-1 w-full">
          <label htmlFor="name-input" className="block text-sm font-medium text-gray-700 mb-1">Your Name</label>
          <input
            id="name-input"
            type="text"
            value={name}
            onChange={onNameChange}
            className="w-full p-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-200 focus:border-indigo-500 text-gray-900 shadow-sm text-base"
            placeholder="Enter your name"
          />
        </div>
        <button
          onClick={onSave}
          className="w-full sm:w-auto px-5 py-2.5 bg-indigo-600 text-white font-medium rounded-lg hover:bg-indigo-700 transition-all duration-200 flex items-center justify-center space-x-2 shadow-sm hover:shadow transform hover:-translate-y-0.5 text-base"
        >
          <FaSave className="text-sm" />
          <span>Save Name</span>
        </button>
      </div>
    </section>
  );
}
