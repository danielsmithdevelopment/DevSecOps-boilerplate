import axios from 'axios';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

const api = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add token to requests if available
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export interface ChallengeResponse {
  message: string;
  nonce: string;
}

export interface VerifyResponse {
  token: string;
  user: {
    id: string;
    wallet_address: string;
  };
}

export interface User {
  id: string;
  wallet_address: string;
  created_at: string;
  last_login?: string;
}

export const authAPI = {
  getChallenge: async (address: string): Promise<ChallengeResponse> => {
    const response = await api.get<ChallengeResponse>('/auth/challenge', {
      params: { address },
    });
    return response.data;
  },

  verifySignature: async (
    message: string,
    signature: string
  ): Promise<VerifyResponse> => {
    try {
      const response = await api.post<VerifyResponse>('/auth/verify', {
        message,
        signature,
      });
      return response.data;
    } catch (error: any) {
      // Extract error message from response
      if (error.response?.data?.error) {
        throw new Error(error.response.data.error);
      }
      throw error;
    }
  },

  getMe: async (): Promise<User> => {
    const response = await api.get<User>('/auth/me');
    return response.data;
  },
};

export default api;
