'use client';

import { useState } from "react";
import axios from "axios";
import Swal from "sweetalert2";
import Googleicon from "../../public/google-icon.svg";
import Image from "next/image";

export default function ForgotPassword() {
  const [email, setEmail] = useState("");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!email) {
      Swal.fire({
        icon: "warning",
        title: "Incomplete Form",
        text: "Email field is required.",
      });
      return;
    }

    try {
      const response = await axios.post("http://localhost:8000/forgot-password", {
        email,
      });

      if (response.status === 200) {
        Swal.fire({
          icon: "success",
          title: "Email Sent",
          text: "Please check your email for the reset link.",
        });
      }
    } catch (error) {
      Swal.fire({
        icon: "error",
        title: "Error",
        text: "Failed to send reset email. Please try again.",
      });
    }
  };

  return (
    <div className="flex h-screen bg-gray-100">
      {/* Left Side */}
      <div className="flex-1 flex flex-col justify-center items-center bg-white px-10">
        <div className="w-full max-w-sm"> {/* Reduced max width to match the login form */}
          <h2 className="text-3xl font-extrabold text-center mb-6 text-gray-800">
            Forgot Password?
          </h2>
          <p className="text-base text-gray-600 text-center mb-8">
            Enter your email to receive a reset link.
          </p>

          <form onSubmit={handleSubmit} className="space-y-6"> {/* Adjusted spacing */}
            <div>
              <label
                htmlFor="email"
                className="block text-base font-medium text-gray-700 mb-2"
              >
                Email
              </label>
              <input
                type="email"
                id="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="mt-1 block w-full rounded-lg border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 p-2.5 text-base"
                placeholder="Enter your email"
                required
              />
            </div>

            <button
              type="submit"
              className="w-full bg-indigo-600 text-white py-2.5 px-4 rounded-lg text-base font-semibold hover:bg-indigo-500 transition"
            >
              Send Reset Link
            </button>
          </form>

          <div className="mt-8 text-center"> {/* Adjusted margin */}
            <p className="text-base text-gray-600">
              Remember your password?{" "}
              <a
                href="/"
                className="text-indigo-600 font-semibold hover:text-indigo-500"
              >
                Login here
              </a>
            </p>
          </div>
        </div>
      </div>

      {/* Right Side */}
      <div className="flex-1 flex flex-col items-center justify-center bg-purple-300">
        <div className="relative">
          {/* Illustration */}
          <div className="flex flex-col items-center">
            <div className="rounded-xl p-6 relative">
              <img
                src="https://cdn.prod.website-files.com/603c87adb15be3cb0b3ed9b5/670cd5385947824a1a82c844_108.webp"
                alt="Illustration"
                className=""
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}