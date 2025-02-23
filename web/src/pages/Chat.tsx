import React, { useState, useEffect, useRef } from "react";
import { useNavigate } from "react-router";
import useLocalStorage from "../store/useLocalStorage";

interface Message {
  timestamp: number;
  username: string;
  type: "JOIN" | "LEAVE" | "CHAT";
  content: string;
}

const Chat: React.FC = () => {
  const { username } = useLocalStorage();
  const navigate = useNavigate();
  const [message, setMessage] = useState("");
  const [messages, setMessages] = useState<Message[]>([]);
  const ws = useRef<WebSocket | null>(null);

  if (!username) {
    navigate("/");
    return null;
  }

  useEffect(() => {
    if (username) {
      const url = `${
        import.meta.env.VITE_API_URL
      }?username=${encodeURIComponent(username)}`;
      ws.current = new WebSocket(url);

      ws.current.onopen = () => {
        console.log("Connected to WebSocket server");
      };

      ws.current.onmessage = (event) => {
        const message: Message = JSON.parse(event.data);
        console.log(message);
        setMessages((prevMessages) => [...prevMessages, message]);
      };

      ws.current.onerror = (error) => {
        console.error("WebSocket error:", error);
      };

      ws.current.onclose = () => {
        console.log("Disconnected from WebSocket server");
      };

      return () => {
        ws.current?.close();
      };
    }
  }, [username]);

  const handleSendMessage = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (ws.current && message.trim() !== "") {
      const msg: Message = {
        timestamp: Date.now(),
        username,
        type: "CHAT",
        content: message,
      };
      ws.current.send(JSON.stringify(msg));
      setMessage("");
    }
  };

  const handleBack = () => {
    navigate("/");
  };

  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-gray-100 p-4">
      <div className="w-full max-w-2xl p-8 bg-white rounded-lg shadow-lg">
        <h1 className="text-3xl font-bold text-center text-blue-600 mb-4">
          Chat Room
        </h1>
        <div className="flex justify-between mb-4">
          <button
            onClick={handleBack}
            className="py-2 px-4 text-white bg-blue-500 rounded-lg hover:bg-blue-600"
          >
            Back to Home
          </button>
        </div>
        <div className="flex flex-col space-y-4">
          <div className="flex flex-col space-y-2 overflow-y-auto h-[600px] p-4 border border-gray-300 rounded-lg">
            {messages.map((msg, index) => (
              <div
                key={index}
                className={`p-2 rounded-lg ${
                  msg.type === "JOIN"
                    ? "bg-green-200 text-green-800"
                    : msg.type === "LEAVE"
                    ? "bg-red-200 text-red-800"
                    : "bg-gray-200"
                }`}
              >
                <div className="text-xs text-gray-500">
                  {new Date(msg.timestamp).toLocaleTimeString()}
                </div>
                {msg.type === "CHAT" ? (
                  <>
                    <strong>{msg.username}: </strong>
                    {msg.content}
                  </>
                ) : (
                  <em>{msg.content}</em>
                )}
              </div>
            ))}
          </div>
          <form
            onSubmit={handleSendMessage}
            className="flex items-center space-x-4"
          >
            <input
              type="text"
              className="flex-grow p-2 border border-blue-300 rounded-lg"
              placeholder="Type your message..."
              value={message}
              onChange={(e) => {
                if (e.target.value.length <= 200) {
                  setMessage(e.target.value);
                }
              }}
            />
            <button
              type="submit"
              className="py-2 px-4 text-white bg-green-500 rounded-lg font-medium hover:bg-green-600 focus:outline-none focus:ring-2 focus:ring-green-500"
            >
              Send
            </button>
          </form>
        </div>
      </div>
    </div>
  );
};

export default Chat;
