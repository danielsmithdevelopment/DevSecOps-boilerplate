import { useState } from 'react';
import { useAuth } from '../hooks/useAuth';

export const WalletConnect = () => {
  const { connect, signIn, signOut, isAuthenticated, isLoading, address, user, provider } = useAuth();
  const [error, setError] = useState<string | null>(null);
  const [isConnecting, setIsConnecting] = useState(false);
  const [isSigningIn, setIsSigningIn] = useState(false);

  const handleConnect = async () => {
    try {
      setError(null);
      setIsConnecting(true);
      await connect();
    } catch (err: any) {
      setError(err.message || 'Failed to connect wallet');
    } finally {
      setIsConnecting(false);
    }
  };

  const handleSignIn = async () => {
    try {
      setError(null);
      setIsSigningIn(true);
      await signIn();
    } catch (err: any) {
      setError(err.message || 'Failed to sign in');
    } finally {
      setIsSigningIn(false);
    }
  };

  const handleSignOut = async () => {
    try {
      await signOut();
    } catch (err: any) {
      setError(err.message || 'Failed to sign out');
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-gray-600">Loading...</div>
      </div>
    );
  }

  if (isAuthenticated && user) {
    return (
      <div className="max-w-md mx-auto mt-8 p-6 bg-white rounded-lg shadow-md">
        <h2 className="text-2xl font-bold mb-4">Welcome!</h2>
        <div className="space-y-2 mb-4">
          <p>
            <span className="font-semibold">Wallet:</span>{' '}
            <span className="font-mono text-sm">{user.wallet_address}</span>
          </p>
          {user.last_login && (
            <p>
              <span className="font-semibold">Last Login:</span>{' '}
              {new Date(user.last_login).toLocaleString()}
            </p>
          )}
        </div>
        <button
          onClick={handleSignOut}
          className="w-full bg-red-500 hover:bg-red-600 text-white font-bold py-2 px-4 rounded"
        >
          Sign Out
        </button>
      </div>
    );
  }

  return (
    <div className="max-w-md mx-auto mt-8 p-6 bg-white rounded-lg shadow-md">
      <h2 className="text-2xl font-bold mb-4 text-center">Sign In with Wallet</h2>
      
      {error && (
        <div className="mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded">
          {error}
        </div>
      )}

      {!address ? (
        <>
          {!provider && (
            <div className="mb-4 p-3 bg-yellow-100 border border-yellow-400 text-yellow-700 rounded">
              <p className="font-semibold">WalletConnect not configured</p>
              <p className="text-sm mt-1">
                Please set VITE_WALLETCONNECT_PROJECT_ID in your environment variables.
                <br />
                Get a free project ID at{' '}
                <a href="https://cloud.walletconnect.com/" target="_blank" rel="noopener noreferrer" className="underline">
                  cloud.walletconnect.com
                </a>
              </p>
            </div>
          )}
          <button
            onClick={handleConnect}
            disabled={isConnecting || !provider}
            className="w-full bg-blue-500 hover:bg-blue-600 disabled:bg-gray-400 text-white font-bold py-3 px-4 rounded"
          >
            {isConnecting ? 'Connecting...' : 'Connect Wallet'}
          </button>
        </>
      ) : (
        <div className="space-y-4">
          <div className="p-3 bg-gray-100 rounded">
            <p className="text-sm text-gray-600">Connected:</p>
            <p className="font-mono text-sm break-all">{address}</p>
          </div>
          <button
            onClick={handleSignIn}
            disabled={isSigningIn}
            className="w-full bg-green-500 hover:bg-green-600 disabled:bg-gray-400 text-white font-bold py-3 px-4 rounded"
          >
            {isSigningIn ? 'Signing In...' : 'Sign In'}
          </button>
        </div>
      )}
    </div>
  );
};
