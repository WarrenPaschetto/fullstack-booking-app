export async function deleteBookingById(id: string, token: string): Promise<void> {
    const res = await fetch(`${process.env.NEXT_PUBLIC_BACKEND_URL}/api/bookings/${id}`, {
        method: "DELETE",
        headers: { Authorization: `Bearer ${token}` },
    });
    if (!res.ok) {
        const text = await res.text();
        throw new Error(`Delete failed: ${res.status} ${text}`);
    }
}