'use client';

import { useState } from "react";
import { useRouter } from "next/navigation";
import axios from "axios";
import Swal from "sweetalert2";
import Googleicon from "../public/google-icon.svg";
import Image from "next/image";

export default function Login() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const router = useRouter();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    const payload = {
      email,
      password,
    };

    try {
      const response = await axios.post("http://localhost:8000/login", payload,{withCredentials: true});

      if (response.status === 200) {
        // Login successful
        Swal.fire({
          icon: "success",
          title: "Login Successful",
          text: "Welcome back!",
        }).then(() => {
          router.push("/note"); // Navigate to /note
        });
      }
    } catch (error) {
      // Login failed
      Swal.fire({
        icon: "error",
        title: "Login Failed",
        text: "Invalid email or password.",
      });
    }
  };

  return (
    <div className="flex h-screen bg-gray-100">
      {/* Left Side */}
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

      {/* Right Side */}
      <div className="flex-1 flex flex-col justify-center items-center bg-white px-10">
        <div className="w-full max-w-sm"> {/* Reduced max width */}
          <h2 className="text-3xl font-extrabold text-center mb-6 text-gray-800">
            Hello!
          </h2>
          <p className="text-base text-gray-600 text-center mb-8">
            Welcome back to Mynote!
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

            <div>
              <label
                htmlFor="password"
                className="block text-base font-medium text-gray-700 mb-2"
              >
                Password
              </label>
              <input
                type="password"
                id="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="mt-1 block w-full rounded-lg border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 p-2.5 text-base"
                placeholder="Enter your password"
                required
              />
              <div className="mt-2 text-right">
                <a
                  href="/forgot_password"
                  className="text-indigo-600 text-sm hover:text-indigo-500"
                >
                  Forgot Password?
                </a>
              </div>
            </div>

            <button
              type="submit"
              className="w-full bg-indigo-600 text-white py-2.5 px-4 rounded-lg text-base font-semibold hover:bg-indigo-500 transition"
            >
              Sign In
            </button>
          </form>

          <div className="mt-6 text-center"> {/* Adjusted margin */}
            <p className="text-base text-gray-600">Or continue with</p>
            <div className="flex justify-center mt-4">
              <button className="bg-gray-100 p-2.5 rounded-full hover:bg-gray-200">
                <Image
                  src={Googleicon}
                  alt="Google"
                  width={20}
                  height={20}
                />
              </button>
            </div>
          </div>

          <p className="mt-8 text-center text-base text-gray-600">
            Not a member?{" "}
            <a
              href="/register"
              className="text-indigo-600 font-semibold hover:text-indigo-500"
            >
              Register now
            </a>
          </p>
        </div>
      </div>
    </div>
  );
}
