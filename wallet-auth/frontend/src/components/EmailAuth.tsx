import { useState, useEffect } from 'react';
import { authAPI } from '../services/api';
import { useAuth } from '../hooks/useAuth';

interface EmailAuthProps {
  onSuccess?: () => void;
  mode?: 'signup' | 'add'; // 'signup' for new users, 'add' for adding email to existing wallet user
}

export const EmailAuth = ({ onSuccess, mode = 'signup' }: EmailAuthProps) => {
  const [email, setEmail] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  const [emailServiceEnabled, setEmailServiceEnabled] = useState<boolean | null>(null);
  const [verificationUrl, setVerificationUrl] = useState<string | null>(null);
  const [isResending, setIsResending] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setIsLoading(true);

    try {
      let response: {
        message: string;
        email_sent?: boolean;
        email_service_enabled?: boolean;
        verification_token?: string;
        verification_url?: string;
      };
      if (mode === 'signup') {
        response = await authAPI.emailSignup(email);
      } else {
        response = await authAPI.addEmail(email);
      }
      
      setEmailServiceEnabled(response.email_service_enabled ?? true);
      setVerificationUrl(response.verification_url || null);
      setSuccess(true);
      if (onSuccess) {
        onSuccess();
      }
    } catch (err: any) {
      setError(err.message || 'Failed to send verification email');
    } finally {
      setIsLoading(false);
    }
  };

  const handleResend = async () => {
    setIsResending(true);
    setError(null);
    try {
      const response = await authAPI.resendVerification(email);
      setEmailServiceEnabled(response.email_service_enabled ?? true);
      setVerificationUrl(response.verification_url || null);
    } catch (err: any) {
      setError(err.message || 'Failed to resend verification email');
    } finally {
      setIsResending(false);
    }
  };

  if (success) {
    return (
      <div className="max-w-md mx-auto mt-8 p-6 bg-white rounded-lg shadow-md">
        <div className="text-center">
          <div className="mb-4 text-green-600">
            <svg
              className="mx-auto h-12 w-12"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M5 13l4 4L19 7"
              />
            </svg>
          </div>
          <h2 className="text-2xl font-bold mb-2">Check your email</h2>
          
          {emailServiceEnabled === false && (
            <div className="mb-4 p-4 bg-yellow-50 border-2 border-yellow-400 rounded-lg">
              <div className="flex items-start">
                <svg className="h-5 w-5 text-yellow-600 mt-0.5 mr-2" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                </svg>
                <div className="text-left">
                  <p className="font-semibold text-yellow-800 mb-1">Email Service Disabled</p>
                  <p className="text-sm text-yellow-700 mb-2">
                    Email sending is not configured. Use the verification link below to verify your email.
                  </p>
                  {verificationUrl && (
                    <div className="mt-3 p-2 bg-white rounded border border-yellow-300">
                      <p className="text-xs text-gray-600 mb-1">Verification Link:</p>
                      <a
                        href={verificationUrl}
                        className="text-xs text-blue-600 hover:text-blue-800 break-all underline"
                        target="_blank"
                        rel="noopener noreferrer"
                      >
                        {verificationUrl}
                      </a>
                    </div>
                  )}
                </div>
              </div>
            </div>
          )}

          {emailServiceEnabled !== false && (
            <p className="text-gray-600 mb-4">
              We've sent a verification link to <strong>{email}</strong>
            </p>
          )}
          
          <p className="text-sm text-gray-500 mb-4">
            Click the link in the email to {mode === 'signup' ? 'sign in' : 'verify your email'}.
          </p>

          <div className="mt-4 pt-4 border-t border-gray-200">
            <p className="text-sm text-gray-600 mb-2">
              Didn't receive the email? Check your spam folder or resend.
            </p>
            <button
              onClick={handleResend}
              disabled={isResending}
              className="w-full bg-gray-100 hover:bg-gray-200 disabled:bg-gray-50 text-gray-700 font-medium py-2 px-4 rounded text-sm"
            >
              {isResending ? 'Resending...' : 'Resend Verification Email'}
            </button>
          </div>

          {error && (
            <div className="mt-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded text-sm">
              {error}
            </div>
          )}
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-md mx-auto mt-8 p-6 bg-white rounded-lg shadow-md">
      <h2 className="text-2xl font-bold mb-4 text-center">
        {mode === 'signup' ? 'Sign In with Email' : 'Add Email Address'}
      </h2>

      {error && (
        <div className="mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded">
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-1">
            Email Address
          </label>
          <input
            id="email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
            placeholder="you@example.com"
          />
        </div>
        <button
          type="submit"
          disabled={isLoading}
          className="w-full bg-blue-500 hover:bg-blue-600 disabled:bg-gray-400 text-white font-bold py-3 px-4 rounded"
        >
          {isLoading ? 'Sending...' : mode === 'signup' ? 'Send Magic Link' : 'Send Verification Email'}
        </button>
      </form>
    </div>
  );
};

// Email verification page component (for handling magic links)
export const EmailVerify = () => {
  const [status, setStatus] = useState<'verifying' | 'success' | 'error'>('verifying');
  const [error, setError] = useState<string | null>(null);
  const { loadUser } = useAuth();

  // Extract token from URL
  const getTokenFromURL = () => {
    const params = new URLSearchParams(window.location.search);
    return params.get('token');
  };

  // Verify email on mount
  useEffect(() => {
    const verifyEmail = async () => {
      const token = getTokenFromURL();
      if (!token) {
        setStatus('error');
        setError('No verification token found');
        return;
      }

      try {
        const { token: jwtToken } = await authAPI.verifyEmail(token);
        localStorage.setItem('token', jwtToken);
        // Load user data to update auth state
        await loadUser();
        setStatus('success');
        // Redirect to home after a short delay
        setTimeout(() => {
          window.location.href = '/';
        }, 2000);
      } catch (err: any) {
        setStatus('error');
        setError(err.message || 'Failed to verify email');
      }
    };

    verifyEmail();
  }, [loadUser]);

  if (status === 'verifying') {
    return (
      <div className="max-w-md mx-auto mt-8 p-6 bg-white rounded-lg shadow-md text-center">
        <div className="mb-4">
          <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
        </div>
        <p className="text-gray-600">Verifying your email...</p>
      </div>
    );
  }

  if (status === 'success') {
    return (
      <div className="max-w-md mx-auto mt-8 p-6 bg-white rounded-lg shadow-md text-center">
        <div className="mb-4 text-green-600">
          <svg
            className="mx-auto h-12 w-12"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M5 13l4 4L19 7"
            />
          </svg>
        </div>
        <h2 className="text-2xl font-bold mb-2">Email Verified!</h2>
        <p className="text-gray-600 mb-4">Redirecting you to the app...</p>
      </div>
    );
  }

  return (
    <div className="max-w-md mx-auto mt-8 p-6 bg-white rounded-lg shadow-md">
      <div className="text-center">
        <div className="mb-4 text-red-600">
          <svg
            className="mx-auto h-12 w-12"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        </div>
        <h2 className="text-2xl font-bold mb-2">Verification Failed</h2>
        <p className="text-gray-600 mb-4">{error}</p>
        <a
          href="/"
          className="text-blue-500 hover:text-blue-600 underline"
        >
          Return to home
        </a>
      </div>
    </div>
  );
};
