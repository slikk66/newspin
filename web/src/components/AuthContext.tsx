import React, { createContext, useContext, useState, useEffect } from "react";
import { api, setAuth, clearAuth, getToken, getUsername } from "@/lib/api";

interface AuthContextType {
  isLoggedIn: boolean;
  username: string | null;
  login: (username: string, password: string) => Promise<void>;
  register: (username: string, password: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType>({
  isLoggedIn: false,
  username: null,
  login: async () => {},
  register: async () => {},
  logout: () => {},
});

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [username, setUsername] = useState<string | null>(null);

  useEffect(() => {
    const token = getToken();
    const user = getUsername();
    if (token && user) {
      setIsLoggedIn(true);
      setUsername(user);
    }
  }, []);

  const login = async (user: string, password: string) => {
    const data = await api.post("/api/login", { username: user, password });
    setAuth(data.token, user);
    setIsLoggedIn(true);
    setUsername(user);
  };

  const register = async (user: string, password: string) => {
    await api.post("/api/register", { username: user, password });
    await login(user, password);
  };

  const logout = () => {
    clearAuth();
    setIsLoggedIn(false);
    setUsername(null);
  };

  return (
    <AuthContext.Provider value={{ isLoggedIn, username, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export const useAuth = () => useContext(AuthContext);
