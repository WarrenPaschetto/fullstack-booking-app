export interface Booking {
    ID: string;
    AppointmentStart: string;
    DurationMinutes: number;
}

export async function fetchAllBookings(token: string): Promise<Booking[]> {
    const resp = await fetch(`${process.env.NEXT_PUBLIC_BACKEND_URL}/api/bookings/all`, {
        headers: { Authorization: `Bearer ${token}` },
    });
    if (!resp.ok) throw new Error("Failed to fetch bookings");
    return await resp.json();
}