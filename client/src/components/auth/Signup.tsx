import React, { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAuthStore } from "../../store/authStore";
import toast, { Toaster } from "react-hot-toast";
import { AuthResponse } from "../types";

const Signup = () => {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const login = useAuthStore((state) => state.login);
  const navigate = useNavigate();

  const handleChangeUsername = (e: React.ChangeEvent<HTMLInputElement>) => {
    setUsername(e.target.value);
  };
  const handleChangePassword = (e: React.ChangeEvent<HTMLInputElement>) => {
    setPassword(e.target.value);
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    try {
      const response = await fetch(
        //@ts-ignore
        `${import.meta.env.VITE_API_URL}/auth/signup`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ username, password }),
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

      const data: AuthResponse = await response.json();
      login(data);
      toast.success(data.message);
      navigate("/rooms");
    } catch (error) {
      console.error("There was a problem with the signup request:", error);
      toast.error("An unexpected error occurred. Please try again.");
    }
  };

  return (
    <>
      <Toaster />
      <form
        className="w-full max-w-md mx-auto mt-10 p-6 sm:p-8 space-y-6 bg-white border border-gray-200 rounded-lg shadow-md"
        onSubmit={handleSubmit}
      >
        <h1 className="text-xl sm:text-2xl font-semibold text-center text-black">
          Welcome! Join us.
        </h1>
        <div className="space-y-4">
          <input
            className="w-full py-2 px-3 sm:px-4 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-green-500"
            placeholder="Username"
            onChange={handleChangeUsername}
            value={username}
            required
          />
          <input
            className="w-full py-2 px-3 sm:px-4 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-green-500"
            placeholder="Password"
            onChange={handleChangePassword}
            type="password"
            value={password}
            required
          />
        </div>
        <button
          className="w-full py-2 px-4 text-white bg-green-500 rounded-lg font-medium hover:bg-green-600 focus:outline-none focus:ring-2 focus:ring-green-500"
          type="submit"
        >
          Signup
        </button>
        <p className="text-center text-gray-800">
          Already have an account?{" "}
          <Link
            to="/auth/login"
            className="underline text-green-500 hover:text-green-700"
          >
            Login
          </Link>
        </p>
      </form>
    </>
  );
};

export default Signup;
