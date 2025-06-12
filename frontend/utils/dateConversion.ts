export function dateTimeToString(date: Date) {
    const day = date.getDay;
    const month = date.getMonth;
    const year = date.getFullYear;
    const hour = date.getHours;
    const minutes = date.getMinutes;
    return `${year}${month}${day}T${hour}:${minutes}Z`
}

export function stringToDate(date: string) {
    const year = date.slice(0, 4)
    const month = date.slice(4, 6)
}

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