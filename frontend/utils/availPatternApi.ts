export async function availabilityPattern(params: {
    dayOfWeek: number
    startTime: Date
    endTime: Date
}, token: string) {
    const url = `${process.env.NEXT_PUBLIC_BACKEND_URL}/api/admin/avail-pattern/create`
    const res = await fetch(url, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            "Authorization": `Bearer ${token}`,
        },
        body: JSON.stringify({
            day_of_week: params.dayOfWeek,
            start_time: params.startTime.toISOString(),
            end_time: params.endTime.toISOString(),
        }),
    })
    if (!res.ok) {
        const text = await res.text()
        throw new Error(`Failed to create availability pattern: ${res.status} ${text}`)
    }
}
