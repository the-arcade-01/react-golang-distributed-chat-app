import React from "react";
import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import { AuthLayout } from "./components/auth";
import Login from "./components/auth/Login";
import Signup from "./components/auth/Signup";
import Home from "./components/home/Home";
import RoomLayout from "./components/rooms";

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/auth" element={<AuthLayout />}>
          <Route path="login" element={<Login />} />
          <Route path="signup" element={<Signup />} />
        </Route>
        <Route path="/rooms" element={<RoomLayout />} />
      </Routes>
    </Router>
  );
}

export default App;
