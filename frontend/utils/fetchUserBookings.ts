import { Booking } from "./fetchAllBookings";

export async function fetchUserBookings(token: string): Promise<Booking[]> {
    const resp = await fetch(`${process.env.NEXT_PUBLIC_BACKEND_URL}/api/bookings/user`, {
        headers: { Authorization: `Bearer ${token}` },
    });

    if (!resp.ok) {
        const errorText = await resp.text(); // read once
        console.error("Fetch failed with:", errorText);
        throw new Error("Failed to fetch bookings");
    }

    const data = await resp.json(); // ✅ only read if not already read
    console.log("Raw response:", data);
    return data;
}
