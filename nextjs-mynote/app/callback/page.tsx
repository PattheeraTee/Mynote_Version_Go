// "use client";

// import { useRouter } from "next/navigation";
// import { useEffect } from "react";

// const Callback = () => {
//   const router = useRouter();

//   useEffect(() => {
//     const params = new URLSearchParams(window.location.search);
//     const code = params.get("code");
//     const state = params.get("state");
//     const error = params.get("error");

//     if (error) {
//       console.error("OAuth Error:", error);
//       return;
//     }

//     if (code && state) {
//       // ส่ง code และ state ไปยัง backend ผ่าน GET
//       fetch(`http://localhost:8000/callback?code=${encodeURIComponent(code)}&state=${encodeURIComponent(state)}`, {
//         method: "GET", // ใช้ GET เพื่อส่ง code และ state
//         credentials: "include", // ส่ง cookie session
//       })
//         .then((res) => {
//           if (!res.ok) {
//             throw new Error(`Server responded with ${res.status}`);
//           }
//           return res.json();
//         })
//         .then((data) => {
//           console.log("OAuth Callback Success:", data);
//           // Redirect ไปที่หน้าที่ต้องการ (อาจเป็นหน้าแสดงผลลัพธ์)
//           router.push("/success-page"); // เปลี่ยน URL หลังสำเร็จ
//         })
//         .catch((err) => {
//           console.error("Error handling OAuth Callback:", err);
//         });
//     }
//   }, [router]);

//   return <div>Processing OAuth...</div>;
// };

// export default Callback;
