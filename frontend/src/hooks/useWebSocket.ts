import { useEffect, useRef, useCallback } from "react";
import { WebSocketService, type WSCallback } from "../services/websocket";

export function useWebSocket(url: string, onMessage: WSCallback) {
  const wsRef = useRef<WebSocketService | null>(null);
  const callbackRef = useRef(onMessage);

  // Update callback ref when onMessage changes
  useEffect(() => {
    callbackRef.current = onMessage;
  }, [onMessage]);

  useEffect(() => {
    const ws = new WebSocketService(url);
    wsRef.current = ws;

    const unsubscribe = ws.subscribe((message) => {
      callbackRef.current(message);
    });

    ws.connect();

    return () => {
      unsubscribe();
      ws.disconnect();
    };
  }, [url]);

  const reconnect = useCallback(() => {
    wsRef.current?.connect();
  }, []);

  return { reconnect };
}
