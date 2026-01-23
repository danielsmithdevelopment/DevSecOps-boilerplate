import { useState } from 'react';
import { useAuth } from '../hooks/useAuth';
import { EmailAuth } from './EmailAuth';

export const WalletConnect = () => {
  const { connect, signIn, signOut, addWallet, loadUser, isAuthenticated, isLoading, address, user, provider } = useAuth();
  const [error, setError] = useState<string | null>(null);
  const [isConnecting, setIsConnecting] = useState(false);
  const [isSigningIn, setIsSigningIn] = useState(false);
  const [authMode, setAuthMode] = useState<'wallet' | 'email'>('wallet');
  const [showAddWallet, setShowAddWallet] = useState(false);
  const [showAddEmail, setShowAddEmail] = useState(false);
  const [isAddingWallet, setIsAddingWallet] = useState(false);

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
      setShowAddWallet(false);
      setShowAddEmail(false);
      setAuthMode('wallet');
    } catch (err: any) {
      setError(err.message || 'Failed to sign out');
    }
  };

  const handleAddWallet = async () => {
    try {
      setError(null);
      setIsAddingWallet(true);
      
      // Connect wallet if not already connected
      let walletAddress = address;
      if (!walletAddress) {
        walletAddress = await connect();
      }
      
      // Add wallet to user (user should have email if they're adding wallet)
      if (walletAddress && user) {
        await addWallet(walletAddress);
        setShowAddWallet(false);
        await loadUser();
      }
    } catch (err: any) {
      setError(err.message || 'Failed to add wallet');
    } finally {
      setIsAddingWallet(false);
    }
  };

  const handleEmailAdded = async () => {
    setShowAddEmail(false);
    await loadUser();
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-gray-600">Loading...</div>
      </div>
    );
  }

  if (isAuthenticated && user) {
    const hasWallet = !!user.wallet_address;
    const hasEmail = !!user.email && user.email_verified;

    return (
      <div className="max-w-md mx-auto mt-8 p-6 bg-white rounded-lg shadow-md">
        <h2 className="text-2xl font-bold mb-4">Welcome!</h2>
        
        {error && (
          <div className="mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded">
            {error}
          </div>
        )}

        <div className="space-y-2 mb-4">
          {hasWallet && (
            <p>
              <span className="font-semibold">Wallet:</span>{' '}
              <span className="font-mono text-sm">{user.wallet_address}</span>
            </p>
          )}
          {hasEmail && (
            <p>
              <span className="font-semibold">Email:</span>{' '}
              <span className="text-sm">{user.email}</span>
              {user.email_verified && (
                <span className="ml-2 text-xs text-green-600">✓ Verified</span>
              )}
            </p>
          )}
          {user.last_login && (
            <p>
              <span className="font-semibold">Last Login:</span>{' '}
              {new Date(user.last_login).toLocaleString()}
            </p>
          )}
        </div>

        {/* Add second authentication method */}
        {!hasWallet && (
          <div className="mb-4 p-4 bg-blue-50 rounded border border-blue-200">
            <p className="text-sm text-gray-700 mb-2">
              Add a wallet address to your account for additional security.
            </p>
            {showAddWallet ? (
              <div className="space-y-2">
                {!address && (
                  <button
                    onClick={handleAddWallet}
                    disabled={isConnecting || !provider}
                    className="w-full bg-blue-500 hover:bg-blue-600 disabled:bg-gray-400 text-white font-bold py-2 px-4 rounded text-sm"
                  >
                    {isConnecting ? 'Connecting...' : 'Connect Wallet'}
                  </button>
                )}
                {address && (
                  <button
                    onClick={handleAddWallet}
                    disabled={isAddingWallet}
                    className="w-full bg-green-500 hover:bg-green-600 disabled:bg-gray-400 text-white font-bold py-2 px-4 rounded text-sm"
                  >
                    {isAddingWallet ? 'Adding Wallet...' : 'Add Wallet'}
                  </button>
                )}
                <button
                  onClick={() => setShowAddWallet(false)}
                  className="w-full bg-gray-300 hover:bg-gray-400 text-gray-800 font-bold py-2 px-4 rounded text-sm"
                >
                  Cancel
                </button>
              </div>
            ) : (
              <button
                onClick={() => setShowAddWallet(true)}
                className="w-full bg-blue-500 hover:bg-blue-600 text-white font-bold py-2 px-4 rounded text-sm"
              >
                Add Wallet
              </button>
            )}
          </div>
        )}

        {!hasEmail && (
          <div className="mb-4 p-4 bg-green-50 rounded border border-green-200">
            <p className="text-sm text-gray-700 mb-2">
              Add an email address to your account for passwordless login.
            </p>
            {showAddEmail ? (
              <EmailAuth mode="add" onSuccess={handleEmailAdded} />
            ) : (
              <button
                onClick={() => setShowAddEmail(true)}
                className="w-full bg-green-500 hover:bg-green-600 text-white font-bold py-2 px-4 rounded text-sm"
              >
                Add Email
              </button>
            )}
          </div>
        )}

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
      <h2 className="text-2xl font-bold mb-4 text-center">Sign In</h2>
      
      {/* Auth mode toggle */}
      <div className="mb-4 flex gap-2">
        <button
          onClick={() => setAuthMode('wallet')}
          className={`flex-1 py-2 px-4 rounded font-semibold ${
            authMode === 'wallet'
              ? 'bg-blue-500 text-white'
              : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
          }`}
        >
          Wallet
        </button>
        <button
          onClick={() => setAuthMode('email')}
          className={`flex-1 py-2 px-4 rounded font-semibold ${
            authMode === 'email'
              ? 'bg-blue-500 text-white'
              : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
          }`}
        >
          Email
        </button>
      </div>

      {/* Show email auth if selected */}
      {authMode === 'email' ? (
        <EmailAuth mode="signup" onSuccess={() => {}} />
      ) : (
        <>
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
        </>
      )}
    </div>
  );
};
