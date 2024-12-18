import React, { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAuthStore } from "../../store/authStore";
import toast, { Toaster } from "react-hot-toast";
import { AuthResponse } from "../types";

const Login = () => {
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
        `${import.meta.env.VITE_API_URL}/auth/login`,
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
      console.error("There was a problem with the login request:", error);
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
          Welcome back!
        </h1>
        <div className="space-y-4">
          <input
            className="w-full py-2 px-3 sm:px-4 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-green-500"
            placeholder="Username"
            type="text"
            value={username}
            onChange={handleChangeUsername}
            required
          />
          <input
            className="w-full py-2 px-3 sm:px-4 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-green-500"
            placeholder="Password"
            type="password"
            value={password}
            onChange={handleChangePassword}
            required
          />
        </div>
        <button
          type="submit"
          className="w-full py-2 px-4 text-white bg-green-500 rounded-lg font-medium hover:bg-green-600 focus:outline-none focus:ring-2 focus:ring-green-500"
        >
          Login
        </button>
        <p className="text-center text-gray-800">
          Don't have an account?{" "}
          <Link
            to="/auth/signup"
            className="underline text-green-500 hover:text-green-700"
          >
            Signup
          </Link>
        </p>
      </form>
    </>
  );
};

export default Login;
