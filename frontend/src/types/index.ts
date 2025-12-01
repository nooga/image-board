export interface Topic {
  id: string;
  title: string;
  imageUrl: string;
  author: string;
  createdAt: string;
  updatedAt: string;
}

export interface Message {
  id: string;
  topicId: string;
  content: string;
  author: string;
  createdAt: string;
}

export interface TopicWithMessages {
  topic: Topic;
  messages: Message[];
}

export type WSMessageType = "new_topic" | "new_message" | "error";

export interface WSMessage<T = unknown> {
  type: WSMessageType;
  payload: T;
}

export interface NewTopicPayload {
  topic: Topic;
}

export interface NewMessagePayload {
  message: Message;
}
