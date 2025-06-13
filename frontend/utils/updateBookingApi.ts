export async function updateBooking(params: {
    id: string
    appointmentStart: Date
    durationMinutes: number
}, token: string) {
    const url = `${process.env.NEXT_PUBLIC_BACKEND_URL}/api/bookings/${params.id}`
    const res = await fetch(url, {
        method: "PUT",
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
        throw new Error(`Failed to update booking: ${res.status} ${text}`)
    }
}
