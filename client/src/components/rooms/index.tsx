import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import toast, { Toaster } from "react-hot-toast";
import { useAuthStore } from "../../store/authStore";
import { RoomDetails } from "../types";

const RoomLayout = () => {
  const { user } = useAuthStore((state) => state);
  const [rooms, setRooms] = useState<RoomDetails[]>([]);
  const [loading, setLoading] = useState(true);
  const [roomName, setRoomName] = useState("");
  const navigate = useNavigate();

  useEffect(() => {
    const fetchRooms = async () => {
      try {
        const response = await fetch(
          //@ts-ignore
          `${import.meta.env.VITE_API_URL}/rooms`,
          {
            method: "GET",
            headers: {
              "Content-Type": "application/json",
              Authorization: `Bearer ${user?.token}`,
            },
          }
        );

        if (response.status >= 400 && response.status <= 500) {
          const errorData = await response.json();
          toast.error(errorData.message);
          return;
        }

        if (!response.ok) {
          throw new Error("Network response was not ok");
        }

        const data = await response.json();
        setRooms(data.data);
      } catch (error) {
        console.error("There was a problem with the fetch request:", error);
      } finally {
        setLoading(false);
      }
    };

    fetchRooms();
  }, [user]);

  const handleCreateRoom = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    try {
      const response = await fetch(
        //@ts-ignore
        `${import.meta.env.VITE_API_URL}/rooms`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${user?.token}`,
          },
          body: JSON.stringify({ room_name: roomName }),
        }
      );

      if (response.status >= 400 && response.status <= 500) {
        const errorData = await response.json();
        toast.error(errorData.message);
        return;
      }

      if (!response.ok) {
        throw new Error("Network response was not ok");
      }

      const data = await response.json();
      setRooms((prevRooms) => [...prevRooms, data.data]);
      toast.success("Room created successfully!");
      setRoomName("");
    } catch (error) {
      console.error("There was a problem with the create room request:", error);
      toast.error("An unexpected error occurred. Please try again.");
    }
  };

  const handleDeleteRoom = async (roomId: string) => {
    try {
      const response = await fetch(
        //@ts-ignore
        `${import.meta.env.VITE_API_URL}/rooms/${roomId}`,
        {
          method: "DELETE",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${user?.token}`,
          },
        }
      );

      if (response.status >= 400 && response.status <= 500) {
        const errorData = await response.json();
        toast.error(errorData.message);
        return;
      }

      if (!response.ok) {
        throw new Error("Network response was not ok");
      }

      setRooms((prevRooms) =>
        prevRooms.filter((room) => room.room_id !== roomId)
      );
      toast.success("Room deleted successfully!");
    } catch (error) {
      console.error("There was a problem with the delete room request:", error);
      toast.error("An unexpected error occurred. Please try again.");
    }
  };

  const handleRoomClick = (roomId: string) => {
    navigate(`/rooms/${roomId}`);
  };

  if (loading) {
    return <div>Loading...</div>;
  }

  return (
    <>
      <Toaster />
      <div className="flex flex-col items-center justify-center min-h-screen bg-gray-100 p-4">
        <div className="w-full max-w-4xl p-8 bg-white rounded-lg shadow-lg">
          <div className="flex items-center justify-between mb-4">
            <button
              onClick={() => navigate("/")}
              className="text-green-500 hover:text-green-700"
            >
              Back
            </button>
            <h1 className="text-2xl font-bold text-center text-black">
              Available Chat Rooms
            </h1>
            <div></div> {/* Placeholder to keep the title centered */}
          </div>
          <form onSubmit={handleCreateRoom} className="mb-8">
            <div className="flex items-center space-x-4">
              <input
                type="text"
                className="w-full py-2 px-4 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-green-500"
                placeholder="Enter room name"
                value={roomName}
                onChange={(e) => setRoomName(e.target.value)}
                required
              />
              <button
                type="submit"
                className="py-2 px-4 text-white bg-green-500 rounded-lg font-medium hover:bg-green-600 focus:outline-none focus:ring-2 focus:ring-green-500"
              >
                Create Room
              </button>
            </div>
          </form>
          {rooms.length === 0 ? (
            <p className="text-center text-gray-600">No rooms available.</p>
          ) : (
            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
              {rooms.map((room) => (
                <div
                  key={room.room_id}
                  className="p-4 bg-gray-200 rounded-lg shadow-md cursor-pointer"
                  onClick={() => handleRoomClick(room.room_id)}
                >
                  <h2 className="text-xl font-semibold">{room.room_name}</h2>
                  <p className="text-gray-700">Admin: {room.admin}</p>
                  {room.admin === user?.username && (
                    <button
                      className="mt-2 px-4 py-2 bg-red-500 text-white rounded"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleDeleteRoom(room.room_id);
                      }}
                    >
                      Delete
                    </button>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </>
  );
};

export default RoomLayout;
