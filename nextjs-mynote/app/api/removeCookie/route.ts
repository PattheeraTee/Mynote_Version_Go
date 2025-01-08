import { NextResponse } from 'next/server';

export async function GET() {
    try {
        // ตั้งค่าคุกกี้ jwt และ access_token ให้หมดอายุ
        const response = NextResponse.json({ message: 'Cookies jwt and access_token have been deleted' });

        response.headers.set(
            'Set-Cookie',
            [
                `jwt=; Path=/; Max-Age=0; HttpOnly; Secure;`,
                `access_token=; Path=/; Max-Age=0; HttpOnly; Secure;`,
            ].join(', ')
        );

        return response;
    } catch (error) {
        console.error('Error deleting cookies:', error);

        // ส่งข้อความตอบกลับเมื่อเกิดข้อผิดพลาด
        return new Response(JSON.stringify({ error: 'Failed to delete cookies' }), { status: 500 });
    }
}
