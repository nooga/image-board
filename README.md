# Image Board

A real-time anonymous image board with live chat. Built with React/TypeScript frontend and Go backend.

## Features

- Post images to create discussion topics
- Real-time chat within each topic
- Live updates for new topics on the homepage
- Anonymous posting with optional nickname (stored in localStorage)
- Clean, minimalistic UI

## Tech Stack

### Frontend

- React 18 with TypeScript
- Vite for build tooling
- WebSocket for real-time updates

### Backend

- Go with Chi router
- MongoDB for persistence
- Redis for pub/sub messaging
- MinIO (S3-compatible) for image storage
- Gorilla WebSocket

## Project Structure

```
image-board/
├── frontend/          # React TypeScript frontend
│   ├── src/
│   │   ├── components/
│   │   ├── pages/
│   │   ├── hooks/
│   │   ├── services/
│   │   └── types/
│   └── package.json
├── backend/           # Go API server
│   ├── cmd/
│   ├── internal/
│   │   ├── api/
│   │   ├── models/
│   │   ├── repository/
│   │   ├── storage/
│   │   └── websocket/
│   └── go.mod
└── docker-compose.yml
```

## Development Setup

### Prerequisites

- Docker & Docker Compose
- Node.js 18+
- Go 1.21+

### Quick Start

1. Start infrastructure services:

```bash
docker-compose up -d mongo redis minio
```

2. Start the backend:

```bash
cd backend
go run cmd/server/main.go
```

3. Start the frontend:

```bash
cd frontend
npm install
npm run dev
```

4. Open http://localhost:5173 in your browser

### Environment Variables

Backend (`.env` or environment):

- `MONGO_URI` - MongoDB connection string (default: `mongodb://localhost:27017`)
- `REDIS_ADDR` - Redis address (default: `localhost:6379`)
- `MINIO_ENDPOINT` - MinIO endpoint (default: `localhost:9000`)
- `MINIO_ACCESS_KEY` - MinIO access key (default: `minioadmin`)
- `MINIO_SECRET_KEY` - MinIO secret key (default: `minioadmin`)
- `PORT` - Server port (default: `8080`)

## API Endpoints

### REST API

- `GET /api/topics` - List all topics
- `POST /api/topics` - Create a new topic (multipart form with image)
- `GET /api/topics/:id` - Get topic details with messages
- `POST /api/topics/:id/messages` - Post a message to a topic

### WebSocket

- `WS /ws/feed` - Real-time feed of new topics
- `WS /ws/topics/:id` - Real-time chat for a specific topic

## License

MIT
