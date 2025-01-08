'use client';

import { useState, useEffect } from "react";
import axios from "axios";
import Swal from "sweetalert2";

export default function ChangePassword({ params }: { params: { token: string } }) {
  const [newPassword, setnewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [token, setToken] = useState("");

  // Resolve the `params` promise
  useEffect(() => {
    async function resolveParams() {
      const resolvedParams = await params; // Ensure params are awaited if required
      setToken(resolvedParams.token);
    }
    resolveParams();
  }, [params]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    // Check if passwords match
    if (newPassword !== confirmPassword) {
      Swal.fire({
        icon: "error",
        title: "Error",
        text: "Passwords do not match.",
      });
      return;
    }

    try {
      const response = await axios.post("http://localhost:8000/reset-password", {
        token,
        newPassword,
        confirmPassword
      });

      if (response.status === 200) {
        Swal.fire({
          icon: "success",
          title: "Password Changed",
          text: "Your password has been successfully changed!",
        }).then(() => {
          window.location.href = "/"; // Redirect to login page
        });
      }
    } catch (error) {
      Swal.fire({
        icon: "error",
        title: "Error",
        text: "Failed to change password. Please try again.",
      });
    }
  };

  return (
    <div className="flex h-screen bg-gray-100">
      {/* Left Side */}
      <div className="flex-1 flex flex-col justify-center items-center bg-white px-10">
        <div className="w-full max-w-sm">
          <h2 className="text-3xl font-extrabold text-center mb-6 text-gray-800">
            Change Password
          </h2>
          <p className="text-base text-gray-600 text-center mb-8">
            Enter and confirm your new password.
          </p>

          <form onSubmit={handleSubmit} className="space-y-6">
            <div>
              <label
                htmlFor="password"
                className="block text-base font-medium text-gray-700 mb-2"
              >
                New Password
              </label>
              <input
                type="password"
                id="password"
                value={newPassword}
                onChange={(e) => setnewPassword(e.target.value)}
                className="mt-1 block w-full rounded-lg border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 p-2.5 text-base"
                placeholder="Enter your new password"
                required
              />
            </div>

            <div>
              <label
                htmlFor="confirmPassword"
                className="block text-base font-medium text-gray-700 mb-2"
              >
                Confirm Password
              </label>
              <input
                type="password"
                id="confirmPassword"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                className="mt-1 block w-full rounded-lg border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 p-2.5 text-base"
                placeholder="Confirm your new password"
                required
              />
            </div>

            <button
              type="submit"
              className="w-full bg-indigo-600 text-white py-2.5 px-4 rounded-lg text-base font-semibold hover:bg-indigo-500 transition"
            >
              Change Password
            </button>
          </form>
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
