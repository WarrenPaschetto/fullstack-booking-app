export interface Slot {
    id: string;
    start_time: string;
    end_time: string;
}

export interface FormattedSlot {
    id: string;
    displayTime: string;
    startTime: string;
}

export async function listFreeSlots(selectedDate: Date, provider: string): Promise<FormattedSlot[]> {
    const iso = selectedDate.toISOString().slice(0, 10);

    try {
        const res = await fetch(
            `${process.env.NEXT_PUBLIC_API_URL}/api/availabilities/free?start=${iso}T00:00:00Z&end=${iso}T23:59:59Z&provider=${provider}`
        );

        if (!res.ok) throw new Error("Failed to fetch slots");

        const data: Slot[] = await res.json();

        const formatter = new Intl.DateTimeFormat("en-US", {
            hour: "numeric",
            minute: "2-digit",
            hour12: true,
        });

        return data.map((slot) => ({
            id: slot.id,
            displayTime: formatter.format(new Date(slot.start_time)),
            startTime: slot.start_time,
        }));
    } catch (err) {
        console.error("Error fetching free slots:", err);
        return [];
    }
}