import { Link } from 'react-router-dom';
import type { Topic } from '../types';

interface TopicCardProps {
  topic: Topic;
  isNew?: boolean;
}

export function TopicCard({ topic, isNew = false }: TopicCardProps) {
  const timeAgo = formatTimeAgo(new Date(topic.createdAt));

  return (
    <Link to={`/topic/${topic.id}`} className={`topic-card ${isNew ? 'topic-card--new' : ''}`}>
      <div className="topic-card__image-wrapper">
        <img 
          src={topic.imageUrl} 
          alt={topic.title}
          className="topic-card__image"
          loading="lazy"
        />
      </div>
      <div className="topic-card__content">
        <h3 className="topic-card__title">{topic.title}</h3>
        <div className="topic-card__meta">
          <span className="topic-card__author">{topic.author}</span>
          <span className="topic-card__separator">·</span>
          <span className="topic-card__time">{timeAgo}</span>
        </div>
      </div>
    </Link>
  );
}

function formatTimeAgo(date: Date): string {
  const now = new Date();
  const diffInSeconds = Math.floor((now.getTime() - date.getTime()) / 1000);

  if (diffInSeconds < 60) {
    return 'just now';
  }

  const diffInMinutes = Math.floor(diffInSeconds / 60);
  if (diffInMinutes < 60) {
    return `${diffInMinutes}m ago`;
  }

  const diffInHours = Math.floor(diffInMinutes / 60);
  if (diffInHours < 24) {
    return `${diffInHours}h ago`;
  }

  const diffInDays = Math.floor(diffInHours / 24);
  if (diffInDays < 7) {
    return `${diffInDays}d ago`;
  }

  return date.toLocaleDateString();
}

