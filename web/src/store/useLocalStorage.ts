import { create } from "zustand";

interface LocalStorageState {
  username: string;
  setUsername: (username: string) => void;
  getUsername: () => string | null;
}

const useLocalStorage = create<LocalStorageState>((set) => ({
  username: localStorage.getItem("username") || "",
  setUsername: (username: string) => {
    localStorage.setItem("username", username);
    set({ username });
  },
  getUsername: () => {
    return localStorage.getItem("username");
  },
}));

export default useLocalStorage;
