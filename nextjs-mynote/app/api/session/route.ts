import { cookies } from "next/headers";

// Define the GET function for retrieving session_id
export async function GET() {
  // Await cookies to resolve the Promise
  const cookieStore = await cookies(); // Resolve the Promise

  // Retrieve the session_id cookie
  const cookie = cookieStore.get("session_id");

  if (cookie) {
    // If session_id is found, return it
    const sessionId = cookie.value;
    return new Response(JSON.stringify({ session_id: sessionId }), { status: 200 });
  } else {
    // If session_id is not found, return an error response
    return new Response(JSON.stringify({ error: "Session ID not found" }), { status: 404 });
  }
}
