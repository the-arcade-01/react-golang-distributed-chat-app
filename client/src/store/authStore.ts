// src/store/authStore.ts
import { create } from "zustand";
import { AuthResponse } from "../components/types";

interface AuthState {
  user: { user_id: number; username: string; token: string } | null;
  login: (data: AuthResponse) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  login: (data: AuthResponse) => set({ user: data.data }),
  logout: () => set({ user: null }),
}));
