export interface User {
    id: string;
    first_name: string;
    last_name: string;
    email: string;
    user_role: string;
}

export async function fetchAllUsers(token: string): Promise<User[]> {
    const resp = await fetch(`${process.env.NEXT_PUBLIC_BACKEND_URL}/api/admin/users/all`, {
        headers: { Authorization: `Bearer ${token}` },
    });
    if (!resp.ok) {
        const text = await resp.text();
        throw new Error(`Error ${resp.status}: ${text}`);
    }
    return await resp.json();
}