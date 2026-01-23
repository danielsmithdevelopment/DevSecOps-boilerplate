import { useState, useEffect, useCallback } from 'react';
import { EthereumProvider } from '@walletconnect/ethereum-provider';
import { authAPI, User } from '../services/api';

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  provider: Awaited<ReturnType<typeof EthereumProvider.init>> | null;
  address: string | null;
}

export const useAuth = () => {
  const [state, setState] = useState<AuthState>({
    user: null,
    isAuthenticated: false,
    isLoading: true,
    provider: null,
    address: null,
  });

  const projectId = import.meta.env.VITE_WALLETCONNECT_PROJECT_ID || '';

  // Initialize provider
  useEffect(() => {
    const initProvider = async () => {
      if (!projectId) {
        console.warn('WalletConnect Project ID not set');
        setState((prev) => ({ ...prev, isLoading: false }));
        return;
      }

      try {
        const provider = await EthereumProvider.init({
          projectId,
          chains: [1], // Ethereum mainnet
          optionalChains: [137, 42161, 10, 56], // Polygon, Arbitrum, Optimism, BSC
          showQrModal: true,
          metadata: {
            name: 'Wallet Auth',
            description: 'Sign in with your crypto wallet',
            url: window.location.origin,
            icons: [`${window.location.origin}/favicon.ico`],
          },
        });

        setState((prev) => ({ ...prev, provider }));

        // Check if already connected
        if (provider.accounts.length > 0) {
          const address = provider.accounts[0];
          setState((prev) => ({ ...prev, address }));
          // Load user if token exists
          const token = localStorage.getItem('token');
          if (token) {
            try {
              const user = await authAPI.getMe();
              setState((prev) => ({
                ...prev,
                user,
                isAuthenticated: true,
                isLoading: false,
              }));
            } catch (error) {
              console.error('Failed to load user:', error);
              localStorage.removeItem('token');
              setState((prev) => ({ ...prev, isLoading: false }));
            }
          } else {
            setState((prev) => ({ ...prev, isLoading: false }));
          }
        } else {
          setState((prev) => ({ ...prev, isLoading: false }));
        }
      } catch (error) {
        console.error('Failed to initialize provider:', error);
        setState((prev) => ({ ...prev, isLoading: false }));
      }
    };

    initProvider();
  }, [projectId]);

  // Load user from API
  const loadUser = useCallback(async () => {
    try {
      const token = localStorage.getItem('token');
      if (!token) {
        setState((prev) => ({ ...prev, isLoading: false, isAuthenticated: false }));
        return;
      }

      const user = await authAPI.getMe();
      setState((prev) => ({
        ...prev,
        user,
        isAuthenticated: true,
        isLoading: false,
      }));
    } catch (error) {
      console.error('Failed to load user:', error);
      localStorage.removeItem('token');
      setState((prev) => ({
        ...prev,
        user: null,
        isAuthenticated: false,
        isLoading: false,
      }));
    }
  }, []);

  // Connect wallet
  const connect = useCallback(async () => {
    if (!state.provider) {
      throw new Error('Provider not initialized');
    }

    try {
      const accounts = await state.provider.enable();
      if (accounts.length === 0) {
        throw new Error('No accounts found');
      }

      const address = accounts[0];
      setState((prev) => ({ ...prev, address }));
      return address;
    } catch (error) {
      console.error('Failed to connect wallet:', error);
      throw error;
    }
  }, [state.provider]);

  // Sign in with wallet
  const signIn = useCallback(async () => {
    if (!state.provider || !state.address) {
      throw new Error('Wallet not connected');
    }

    try {
      // Get challenge from backend
      const { message } = await authAPI.getChallenge(state.address);

      // Request signature from wallet
      // personal_sign standard: [message, address]
      console.log('Signing message:', message);
      const signature = await state.provider.request({
        method: 'personal_sign',
        params: [message, state.address],
      });
      console.log('Received signature:', signature);

      // Verify signature with backend
      const { token, user } = await authAPI.verifySignature(message, signature as string);

      // Store token
      localStorage.setItem('token', token);

      // Load full user data
      const fullUser = await authAPI.getMe();

      // Update state
      setState((prev) => ({
        ...prev,
        user: fullUser,
        isAuthenticated: true,
      }));

      return user;
    } catch (error) {
      console.error('Failed to sign in:', error);
      throw error;
    }
  }, [state.provider, state.address]);

  // Sign out
  const signOut = useCallback(async () => {
    if (state.provider) {
      try {
        await state.provider.disconnect();
      } catch (error) {
        console.error('Failed to disconnect provider:', error);
      }
    }

    localStorage.removeItem('token');
    setState({
      user: null,
      isAuthenticated: false,
      isLoading: false,
      provider: state.provider,
      address: null,
    });
  }, [state.provider]);

  return {
    ...state,
    connect,
    signIn,
    signOut,
    loadUser,
    provider: state.provider,
  };
};
