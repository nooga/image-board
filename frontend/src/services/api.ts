import type { Topic, TopicWithMessages, Message } from "../types";

const API_BASE = "/api";

export async function fetchTopics(limit = 50, offset = 0): Promise<Topic[]> {
  const response = await fetch(
    `${API_BASE}/topics?limit=${limit}&offset=${offset}`
  );
  if (!response.ok) {
    throw new Error("Failed to fetch topics");
  }
  return response.json();
}

export async function fetchTopic(id: string): Promise<TopicWithMessages> {
  const response = await fetch(`${API_BASE}/topics/${id}`);
  if (!response.ok) {
    throw new Error("Failed to fetch topic");
  }
  return response.json();
}

export async function createTopic(
  title: string,
  image: File,
  author: string
): Promise<Topic> {
  const formData = new FormData();
  formData.append("title", title);
  formData.append("image", image);
  formData.append("author", author || "Anonymous");

  const response = await fetch(`${API_BASE}/topics`, {
    method: "POST",
    body: formData,
  });

  if (!response.ok) {
    throw new Error("Failed to create topic");
  }
  return response.json();
}

export async function createMessage(
  topicId: string,
  content: string,
  author: string
): Promise<Message> {
  const response = await fetch(`${API_BASE}/topics/${topicId}/messages`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      content,
      author: author || "Anonymous",
    }),
  });

  if (!response.ok) {
    throw new Error("Failed to create message");
  }
  return response.json();
}
