'use client';

import { useState } from "react";
import axios from "axios";
import Swal from "sweetalert2";
import Googleicon from "../../public/google-icon.svg";
import Image from "next/image";
import { useRouter } from "next/navigation";

export default function Register() {
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const router = useRouter();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    // Check for empty fields
    if (!username || !email || !password) {
      Swal.fire({
        icon: "warning",
        title: "Incomplete Form",
        text: "All fields are required.",
      });
      return;
    }

    const payload = {
      username,
      email,
      password,
    };

    try {
      const response = await axios.post("http://localhost:8000/register", payload);

      if (response.status === 201) {
        // Registration successful
        Swal.fire({
          icon: "success",
          title: "Registration Successful",
          text: "Your account has been created successfully!",
        }).then(() => {
          router.push("/"); // Redirect to login page
        });
      }
    } catch (error) {
      Swal.fire({
        icon: "error",
        title: "Register Failed",
        text: "This email has already been registered.",
      });
    }
  };

  return (
    <div className="flex h-screen bg-gray-100">
      {/* Left Side */}
      <div className="flex-1 flex flex-col justify-center items-center bg-white px-10">
        <div className="w-full max-w-sm"> {/* Reduced max width to match the login form */}
          <h2 className="text-3xl font-extrabold text-center mb-6 text-gray-800">
            Join Us!
          </h2>
          <p className="text-base text-gray-600 text-center mb-8">
            Create your account to get started.
          </p>

          <form onSubmit={handleSubmit} className="space-y-6"> {/* Adjusted spacing */}
            <div>
              <label
                htmlFor="username"
                className="block text-base font-medium text-gray-700 mb-2"
              >
                Username
              </label>
              <input
                type="text"
                id="username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="mt-1 block w-full rounded-lg border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 p-2.5 text-base"
                placeholder="Enter your username"
                required
              />
            </div>

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
                placeholder="Create a password"
                required
              />
            </div>

            <button
              type="submit"
              className="w-full bg-indigo-600 text-white py-2.5 px-4 rounded-lg text-base font-semibold hover:bg-indigo-500 transition"
            >
              Register
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
            Already a member?{" "}
            <a
              href="/"
              className="text-indigo-600 font-semibold hover:text-indigo-500"
            >
              Login here
            </a>
          </p>
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
