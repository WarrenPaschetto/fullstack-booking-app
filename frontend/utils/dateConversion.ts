
export function formatAppointment(iso: string): string {
    const dt = new Date(iso);
    const date = dt.toLocaleDateString("en-CA", {
        year: "numeric",
        month: "2-digit",
        day: "2-digit",
    });
    const time = dt.toLocaleTimeString("en-US", {
        hour: "numeric",
        minute: "2-digit",
        hour12: true,
    });
    return `${date} at ${time}`;
}