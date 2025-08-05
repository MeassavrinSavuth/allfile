import ChangePasswordForm from './ChangePasswordForm';

export default function PasswordSection({ provider }) {
  return (
    <section className="space-y-3">
      <h2 className="text-xl font-semibold text-gray-800 border-b pb-2 mb-2">Password</h2>
      {provider === null ? (
        <ChangePasswordForm />
      ) : (
        <p className="text-gray-700 bg-gray-50 p-3 rounded-md border border-gray-200 text-sm shadow-sm">
          You signed up with <strong className="text-indigo-600">{provider}</strong>. Please manage your password through your {provider} account settings.
        </p>
      )}
    </section>
  );
}
