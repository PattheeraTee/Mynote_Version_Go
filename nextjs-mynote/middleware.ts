// import { NextResponse } from 'next/server'
// import type { NextRequest } from 'next/server'

// export function middleware(request: NextRequest): NextResponse | undefined {
//   const token = request.cookies.get('authjs.session-token');
//   console.log('Token from cookie in middleware:', token?.value);

//   if (token) {
//     // ถ้ามีคุกกี้ชื่อ 'token'
//     return;
//   } else {
//     // ถ้าไม่มีคุกกี้ชื่อ 'token'
//     return NextResponse.redirect(new URL('/', request.url));
//   }
// }

// export const config = {
//   matcher: '/note/:path*', // เปลี่ยนเส้นทางสำหรับเส้นทางที่ตรงตามนี้
// };
import {NextResponse} from 'next/server';
import type {NextRequest} from 'next/server';

export function middleware(request: NextRequest): NextResponse | undefined {
  const token = request.cookies.get('jwt');
  console.log('Token from cookie in middleware:', token?.value);

  if (token) {
    // ถ้ามีคุกกี้ชื่อ '
    return;
  }else{
    // ถ้าไม่มีคุกกร้ชื่อ jwt
    return  NextResponse.redirect(new URL('/', request.url));
  }

}

export const config = {
    matcher: '/note/:path*', // เปลี่ยนเส้นทางสำหรับเส้นทางที่ตรงตามนี้
};