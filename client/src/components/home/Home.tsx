import React from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAuthStore } from "../../store/authStore";

const Home = () => {
  const navigate = useNavigate();
  const { user } = useAuthStore((state) => state);

  const handleCreateRoom = () => {
    if (user) {
      navigate("/rooms");
    } else {
      navigate("/auth/login");
    }
  };

  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-gray-100 p-4">
      <div className="w-full max-w-lg p-8 bg-white rounded-lg shadow-lg">
        <h1 className="text-3xl font-bold text-center text-blue-600 mb-4">
          Welcome to Distributed Chat Room App
        </h1>
        <p className="text-center text-gray-600 mb-8">
          Our app is a Distributed Chat room app, where users can create rooms
          and join them to chat with people.
        </p>
        <div className="flex flex-col items-center space-y-4">
          {!user ? (
            <Link
              to="/auth/login"
              className="w-full py-3 px-6 text-lg text-white bg-blue-500 rounded-full font-semibold hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              Get Started
            </Link>
          ) : (
            <button
              onClick={handleCreateRoom}
              className="w-full py-3 px-6 text-lg text-white bg-green-500 rounded-full font-semibold hover:bg-green-600 focus:outline-none focus:ring-2 focus:ring-green-500"
            >
              Go To Chat Room
            </button>
          )}
        </div>
      </div>
    </div>
  );
};

export default Home;
