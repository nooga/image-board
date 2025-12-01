import { useState, useEffect, useCallback } from "react";
import { TopicCard } from "../components/TopicCard";
import { CreateTopicForm } from "../components/CreateTopicForm";
import { fetchTopics } from "../services/api";
import { useWebSocket } from "../hooks/useWebSocket";
import type { Topic, WSMessage, NewTopicPayload } from "../types";

export function HomePage() {
  const [topics, setTopics] = useState<Topic[]>([]);
  const [newTopicIds, setNewTopicIds] = useState<Set<string>>(new Set());
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showForm, setShowForm] = useState(false);

  const loadTopics = useCallback(async () => {
    try {
      const data = await fetchTopics();
      setTopics(data);
      setError(null);
    } catch {
      setError("Failed to load topics");
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    loadTopics();
  }, [loadTopics]);

  const handleWSMessage = useCallback((message: WSMessage) => {
    if (message.type === "new_topic") {
      const payload = message.payload as NewTopicPayload;
      setTopics((prev) => [payload.topic, ...prev]);
      setNewTopicIds((prev) => new Set(prev).add(payload.topic.id));

      // Remove "new" status after animation
      setTimeout(() => {
        setNewTopicIds((prev) => {
          const next = new Set(prev);
          next.delete(payload.topic.id);
          return next;
        });
      }, 2000);
    }
  }, []);

  useWebSocket("/ws/feed", handleWSMessage);

  const handleFormSuccess = () => {
    setShowForm(false);
  };

  if (isLoading) {
    return (
      <div className="page-container">
        <div className="loading">
          <div className="loading-spinner"></div>
          <span>Loading topics...</span>
        </div>
      </div>
    );
  }

  return (
    <div className="page-container">
      <div className="home-header">
        <h1 className="page-title">Latest Topics</h1>
        <button
          className={`btn-new-topic ${showForm ? "btn-new-topic--active" : ""}`}
          onClick={() => setShowForm(!showForm)}
        >
          {showForm ? "✕ Cancel" : "+ New Topic"}
        </button>
      </div>

      {showForm && (
        <div className="form-container">
          <CreateTopicForm onSuccess={handleFormSuccess} />
        </div>
      )}

      {error && <div className="error-message">{error}</div>}

      {topics.length === 0 ? (
        <div className="empty-state">
          <span className="empty-state__icon">◇</span>
          <p>No topics yet. Be the first to post!</p>
        </div>
      ) : (
        <div className="topics-grid">
          {topics.map((topic) => (
            <TopicCard
              key={topic.id}
              topic={topic}
              isNew={newTopicIds.has(topic.id)}
            />
          ))}
        </div>
      )}
    </div>
  );
}
