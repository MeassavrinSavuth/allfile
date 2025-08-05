export default function EmailSection({ email }) {
  return (
    <section className="space-y-3">
      <h2 className="text-xl font-semibold text-gray-800 border-b pb-2 mb-2">Email Address</h2>
      <div className="w-full">
        <label htmlFor="email-input" className="block text-sm font-medium text-gray-700 mb-1">Your Email</label>
        <input
          id="email-input"
          type="email"
          value={email}
          readOnly
          className="w-full p-3 border border-gray-300 rounded-lg bg-gray-50 text-gray-600 cursor-not-allowed shadow-sm text-base"
          aria-describedby="email-help-text"
        />
        <p id="email-help-text" className="mt-1 text-xs text-gray-500">
          Your email is used for login and notifications.
        </p>
      </div>
    </section>
  );
}
