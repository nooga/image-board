import { useState, useEffect, useCallback } from "react";

const NICKNAME_KEY = "imageboard_nickname";

export function useNickname() {
  const [nickname, setNicknameState] = useState<string>(() => {
    return localStorage.getItem(NICKNAME_KEY) || "";
  });

  useEffect(() => {
    const stored = localStorage.getItem(NICKNAME_KEY);
    if (stored) {
      setNicknameState(stored);
    }
  }, []);

  const setNickname = useCallback((newNickname: string) => {
    const trimmed = newNickname.trim();
    if (trimmed) {
      localStorage.setItem(NICKNAME_KEY, trimmed);
    } else {
      localStorage.removeItem(NICKNAME_KEY);
    }
    setNicknameState(trimmed);
  }, []);

  const displayName = nickname || "Anonymous";

  return { nickname, setNickname, displayName };
}
