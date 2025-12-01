import { useState } from "react";
import { Link } from "react-router-dom";
import { useNickname } from "../hooks/useNickname";

export function Header() {
  const { nickname, setNickname, displayName } = useNickname();
  const [isEditing, setIsEditing] = useState(false);
  const [inputValue, setInputValue] = useState(nickname);

  const handleSave = () => {
    setNickname(inputValue);
    setIsEditing(false);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      handleSave();
    } else if (e.key === "Escape") {
      setInputValue(nickname);
      setIsEditing(false);
    }
  };

  return (
    <header className="header">
      <Link to="/" className="logo">
        <span className="logo-icon">◈</span>
        <span className="logo-text">board</span>
      </Link>

      <div className="user-info">
        {isEditing ? (
          <div className="nickname-edit">
            <input
              type="text"
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="Enter nickname"
              autoFocus
              maxLength={24}
            />
            <button onClick={handleSave} className="btn-save">
              ✓
            </button>
            <button
              onClick={() => {
                setInputValue(nickname);
                setIsEditing(false);
              }}
              className="btn-cancel"
            >
              ✕
            </button>
          </div>
        ) : (
          <button
            className="nickname-display"
            onClick={() => {
              setInputValue(nickname);
              setIsEditing(true);
            }}
            title="Click to change nickname"
          >
            <span className="nickname-label">posting as</span>
            <span className="nickname-value">{displayName}</span>
          </button>
        )}
      </div>
    </header>
  );
}
