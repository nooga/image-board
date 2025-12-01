import type { Message } from "../types";

interface ChatMessageProps {
  message: Message;
  isNew?: boolean;
}

export function ChatMessage({ message, isNew = false }: ChatMessageProps) {
  const time = new Date(message.createdAt).toLocaleTimeString([], {
    hour: "2-digit",
    minute: "2-digit",
  });

  return (
    <div className={`chat-message ${isNew ? "chat-message--new" : ""}`}>
      <div className="chat-message__header">
        <span className="chat-message__author">{message.author}</span>
        <span className="chat-message__time">{time}</span>
      </div>
      <div className="chat-message__content">{message.content}</div>
    </div>
  );
}
