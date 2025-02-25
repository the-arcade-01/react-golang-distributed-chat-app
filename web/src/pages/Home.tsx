import React, { useState } from "react";
import { useNavigate } from "react-router";
import useLocalStorage from "../store/useLocalStorage";

const Home: React.FC = () => {
  const { username, setUsername } = useLocalStorage();
  const [inputValue, setInputValue] = useState(username);
  const navigate = useNavigate();

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setInputValue(e.target.value);
  };

  const handleSave = () => {
    setUsername(inputValue);
    navigate("/chat");
  };

  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-gray-100 p-4">
      <div className="w-full max-w-lg p-8 bg-white rounded-lg shadow-lg">
        <h1 className="text-3xl font-bold text-center text-blue-600 mb-4">
          Welcome to Distributed <br /> Chat Room App
        </h1>
        <div className="flex flex-col items-center space-y-4">
          <label className="block mb-4 text-blue-600 w-full">
            {username ? "Change Username:" : "Your Username:"}
            <input
              type="text"
              value={inputValue}
              onChange={handleInputChange}
              className="mt-2 p-2 border border-blue-300 rounded-lg w-full"
            />
          </label>
          <button
            onClick={handleSave}
            className="w-full py-3 px-6 text-lg text-white bg-green-500 rounded-full font-semibold hover:bg-green-600 focus:outline-none focus:ring-2 focus:ring-green-500"
          >
            {username ? "Change" : "Save"}
          </button>
        </div>
      </div>
    </div>
  );
};

export default Home;
