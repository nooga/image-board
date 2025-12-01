import { useState, useEffect, useCallback, useRef } from 'react';
import { useParams, Link } from 'react-router-dom';
import { ChatMessage } from '../components/ChatMessage';
import { fetchTopic, createMessage } from '../services/api';
import { useWebSocket } from '../hooks/useWebSocket';
import { useNickname } from '../hooks/useNickname';
import type { Topic, Message, WSMessage, NewMessagePayload } from '../types';

export function TopicPage() {
  const { id } = useParams<{ id: string }>();
  const { displayName } = useNickname();
  const [topic, setTopic] = useState<Topic | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [newMessageIds, setNewMessageIds] = useState<Set<string>>(new Set());
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [messageInput, setMessageInput] = useState('');
  const [isSending, setIsSending] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const chatContainerRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = useCallback(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, []);

  useEffect(() => {
    if (!id) return;

    const loadTopic = async () => {
      try {
        const data = await fetchTopic(id);
        setTopic(data.topic);
        setMessages(data.messages);
        setError(null);
      } catch {
        setError('Failed to load topic');
      } finally {
        setIsLoading(false);
      }
    };

    loadTopic();
  }, [id]);

  useEffect(() => {
    scrollToBottom();
  }, [messages.length, scrollToBottom]);

  const handleWSMessage = useCallback((wsMessage: WSMessage) => {
    if (wsMessage.type === 'new_message') {
      const payload = wsMessage.payload as NewMessagePayload;
      setMessages((prev) => [...prev, payload.message]);
      setNewMessageIds((prev) => new Set(prev).add(payload.message.id));
      
      setTimeout(() => {
        setNewMessageIds((prev) => {
          const next = new Set(prev);
          next.delete(payload.message.id);
          return next;
        });
      }, 2000);
    }
  }, []);

  useWebSocket(id ? `/ws/topics/${id}` : '', handleWSMessage);

  const handleSendMessage = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!messageInput.trim() || !id || isSending) return;

    setIsSending(true);
    try {
      await createMessage(id, messageInput.trim(), displayName);
      setMessageInput('');
    } catch {
      setError('Failed to send message');
    } finally {
      setIsSending(false);
    }
  };

  if (isLoading) {
    return (
      <div className="page-container">
        <div className="loading">
          <div className="loading-spinner"></div>
          <span>Loading topic...</span>
        </div>
      </div>
    );
  }

  if (error || !topic) {
    return (
      <div className="page-container">
        <div className="error-state">
          <span className="error-state__icon">⚠</span>
          <p>{error || 'Topic not found'}</p>
          <Link to="/" className="btn-back">← Back to home</Link>
        </div>
      </div>
    );
  }

  return (
    <div className="topic-page">
      <div className="topic-content">
        <Link to="/" className="back-link">← Back</Link>
        
        <div className="topic-image-container">
          <img 
            src={topic.imageUrl} 
            alt={topic.title}
            className="topic-image"
          />
        </div>

        <div className="topic-info">
          <h1 className="topic-title">{topic.title}</h1>
          <div className="topic-meta">
            <span className="topic-author">{topic.author}</span>
            <span className="topic-separator">·</span>
            <span className="topic-date">
              {new Date(topic.createdAt).toLocaleDateString()}
            </span>
          </div>
        </div>
      </div>

      <div className="chat-section">
        <div className="chat-header">
          <h2>Discussion</h2>
          <span className="message-count">{messages.length} messages</span>
        </div>

        <div className="chat-messages" ref={chatContainerRef}>
          {messages.length === 0 ? (
            <div className="chat-empty">
              <span>No messages yet. Start the conversation!</span>
            </div>
          ) : (
            messages.map((message) => (
              <ChatMessage
                key={message.id}
                message={message}
                isNew={newMessageIds.has(message.id)}
              />
            ))
          )}
          <div ref={messagesEndRef} />
        </div>

        <form onSubmit={handleSendMessage} className="chat-input-form">
          <input
            type="text"
            value={messageInput}
            onChange={(e) => setMessageInput(e.target.value)}
            placeholder={`Message as ${displayName}...`}
            className="chat-input"
            maxLength={1000}
            disabled={isSending}
          />
          <button 
            type="submit" 
            className="btn-send"
            disabled={!messageInput.trim() || isSending}
          >
            {isSending ? '...' : 'Send'}
          </button>
        </form>
      </div>
    </div>
  );
}

