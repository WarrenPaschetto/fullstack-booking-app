export async function createBooking(params: {
    id: string
    appointmentStart: Date
    durationMinutes: 60
}, token: string) {
    const url = `${process.env.NEXT_PUBLIC_BACKEND_URL}/api/bookings/create`
    const res = await fetch(url, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            "Authorization": `Bearer ${token}`,
        },
        body: JSON.stringify({
            appointment_start: params.appointmentStart.toISOString(),
            duration_minutes: params.durationMinutes,
        }),
    })
    if (!res.ok) {
        const text = await res.text()
        throw new Error(`Failed to create booking: ${res.status} ${text}`)
    }
}