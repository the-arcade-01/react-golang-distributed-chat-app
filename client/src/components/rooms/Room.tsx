import React, { useEffect, useState, useRef } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useAuthStore } from "../../store/authStore";
import toast, { Toaster } from "react-hot-toast";
import { RoomDetails, Message } from "../types";

const Room = () => {
  const { roomId } = useParams<{ roomId: string }>();
  const { user } = useAuthStore((state) => state);
  const [roomDetails, setRoomDetails] = useState<RoomDetails>();
  const [loading, setLoading] = useState(true);
  const [messages, setMessages] = useState<Message[]>([]);
  const [newMessage, setNewMessage] = useState("");
  const navigate = useNavigate();
  const ws = useRef<WebSocket | null>(null);

  useEffect(() => {
    const fetchRoomDetails = async () => {
      try {
        const response = await fetch(`http://localhost:8080/rooms/${roomId}`, {
          method: "GET",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${user?.token}`,
          },
        });

        if (response.status >= 400 && response.status <= 500) {
          const errorData = await response.json();
          toast.error(errorData.message);
          return;
        }

        if (!response.ok) {
          throw new Error("Network response was not ok");
        }

        const data = await response.json();
        setRoomDetails(data.data);
      } catch (error) {
        console.error("There was a problem with the fetch request:", error);
      } finally {
        setLoading(false);
      }
    };

    fetchRoomDetails();
  }, [roomId, user]);

  useEffect(() => {
    if (roomDetails && user?.token) {
      const url = `ws://localhost:8080/ws?room_id=${roomDetails.room_id}&token=${user.token}`;
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
  }, [roomDetails, user?.token]);

  const handleSendMessage = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (ws.current && newMessage.trim() !== "") {
      const message: Message = {
        user: user?.username || "Anonymous",
        type: "CHAT",
        content: newMessage,
      };
      ws.current.send(JSON.stringify(message));
      setNewMessage("");
    }
  };

  if (loading) {
    return <div>Loading...</div>;
  }

  if (!roomDetails) {
    return <div>No room details available.</div>;
  }

  return (
    <>
      <Toaster />
      <div className="flex flex-col items-center min-h-screen bg-gray-100">
        <div className="w-full max-w-2xl p-4 bg-white shadow-md">
          <div className="flex items-center justify-between">
            <button
              onClick={() => navigate("/rooms")}
              className="text-green-500 hover:text-green-700"
            >
              Back
            </button>
            <h1 className="text-xl font-bold text-black">
              {roomDetails.room_name}
            </h1>
            <div className="relative group">
              <button className="text-green-500 hover:text-green-700">
                Info
              </button>
              <div className="absolute right-0 mt-2 w-48 p-4 bg-white border border-gray-300 rounded-lg shadow-lg opacity-0 group-hover:opacity-100 transition-opacity">
                <p className="text-gray-700">Room ID: {roomDetails.room_id}</p>
                <p className="text-gray-700">
                  Active Users: {roomDetails.active_users}
                </p>
              </div>
            </div>
          </div>
        </div>
        <div className="flex-1 w-full max-w-2xl p-4 overflow-y-auto bg-white shadow-md">
          {messages.map((message, index) => (
            <div
              key={index}
              className={`mb-2 p-2 rounded-lg ${
                message.user === user?.username
                  ? "bg-green-200 text-right"
                  : "bg-gray-200"
              }`}
            >
              {message.type === "CHAT" ? (
                <>
                  <strong>{message.user}: </strong>
                  {message.content}
                </>
              ) : (
                <em>{message.content}</em>
              )}
            </div>
          ))}
        </div>
        <div className="w-full max-w-2xl p-4 bg-white shadow-md">
          <form
            onSubmit={handleSendMessage}
            className="flex items-center space-x-4"
          >
            <input
              type="text"
              className="flex-1 py-2 px-4 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-green-500"
              placeholder="Type your message..."
              value={newMessage}
              onChange={(e) => setNewMessage(e.target.value)}
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
    </>
  );
};

export default Room;
