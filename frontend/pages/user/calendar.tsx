import { useState, useEffect, useMemo } from "react";
import Layout from "../../components/Layout";
import dynamic from "next/dynamic";
import { useRequireAuth } from "../../utils/useRequireAuth";
import { createBooking } from "@/utils/createBookingApi";

const Navbar = dynamic(() => import("../../components/Navbar"), {
    ssr: false,
});

export default function UserCalendar() {
    useRequireAuth("user");

    const today = new Date();
    const year = today.getFullYear();
    const month = today.getMonth(); // 0-indexed

    //–– STATE ––
    const [selectedDate, setSelectedDate] = useState<Date | null>(null);
    const [availableTimes, setAvailableTimes] = useState<
        { id: string; displayTime: string; startTime: string }[]
    >([]);
    const [selectedTime, setSelectedTime] = useState<string>("");

    //–– HELPERS ––
    // how many days in this month
    const daysInMonth = useMemo(
        () => new Date(year, month + 1, 0).getDate(),
        [year, month]
    );
    // which weekday does the 1st fall on? (0=Sun … 6=Sat)
    const startWeekday = useMemo(
        () => new Date(year, month, 1).getDay(),
        [year, month]
    );
    // build an array of Date | null for each calendar “cell”
    const calendarDays = useMemo(() => {
        const blanks = Array(startWeekday).fill(null);
        const days = Array(daysInMonth)
            .fill(0)
            .map((_, i) => new Date(year, month, i + 1));
        // pad to full weeks:
        const totalCells = Math.ceil((blanks.length + days.length) / 7) * 7;
        const trailing = Array(totalCells - blanks.length - days.length).fill(null);
        return [...blanks, ...days, ...trailing];
    }, [year, month, daysInMonth, startWeekday]);

    //–– FETCH AVAILABILITY WHEN DAY SELECTED ––
    useEffect(() => {
        if (!selectedDate) return;

        interface Slot {
            id: string;
            start_time: string;
            end_time: string;
        }

        const provider = "f2480f96-e1a3-4e33-9f26-b90910680bec"
        const iso = selectedDate.toISOString().slice(0, 10);
        fetch(`/api/availabilities/free?start=${iso}T00:00:00Z&end=${iso}T23:59:59Z&provider=${provider}`)
            .then((r) => r.json())
            .then((data: Slot[]) => {
                const formatter = new Intl.DateTimeFormat("en-US", {
                    hour: "numeric",
                    minute: "2-digit",
                    hour12: true,
                });

                const formatted = data.map((slot) => ({
                    id: slot.id,
                    displayTime: formatter.format(new Date(slot.start_time)),
                    startTime: slot.start_time,
                }));

                setAvailableTimes(formatted);
            })
            .catch(() => setAvailableTimes([]));
    }, [selectedDate]);

    return (
        <Layout>
            <Navbar />
            <div className="w-full max-w-3xl mx-auto mt-20 p-6 bg-white rounded-lg shadow-md text-center">
                <h1 className="text-3xl font-semibold mb-4">Calendar</h1>
                <p className="text-gray-600">Pick a day to see availability</p>

                {/* Month & Year Header */}
                <div className="mt-4 text-xl font-semibold">
                    {today.toLocaleString("default", { month: "long" })} {year}
                </div>

                {/* Weekday Labels */}
                <div className="grid grid-cols-7 mt-2 text-sm font-medium text-gray-600">
                    {["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"].map((d) => (
                        <div key={d}>{d}</div>
                    ))}
                </div>

                {/* Day Grid */}
                <div className="grid grid-cols-7 gap-2 mt-2">
                    {calendarDays.map((date, idx) => {
                        const isToday =
                            date &&
                            date.getFullYear() === today.getFullYear() &&
                            date.getMonth() === today.getMonth() &&
                            date.getDate() === today.getDate();
                        const isSelected =
                            date &&
                            selectedDate &&
                            date.toDateString() === selectedDate.toDateString();

                        return (
                            <button
                                key={idx}
                                disabled={!date}
                                onClick={() => date && setSelectedDate(date)}
                                className={`
                  p-2 rounded
                  ${!date ? "invisible" : "hover:bg-blue-100"}
                  ${isToday ? "bg-blue-200" : ""}
                  ${isSelected ? "bg-blue-400 text-white" : ""}
                `}
                            >
                                {date?.getDate()}
                            </button>
                        );
                    })}
                </div>

                {/* Availability Panel */}
                {selectedDate && (
                    <div className="mt-6 text-left">
                        <table className="min-w-2/3 bg-white">
                            <thead className="border-b">
                                <tr>
                                    <h2 className="text-2xl font-semibold mb-2">
                                        Times on {selectedDate.toLocaleDateString()}
                                    </h2>
                                </tr>
                            </thead>
                            <tbody>
                                {availableTimes.length > 0 ? (
                                    <div className="grid grid-cols-7 gap-4 mt-4">
                                        {availableTimes.map((slot) => (
                                            <button
                                                key={slot.id}
                                                onClick={async () => {
                                                    const confirmBooking = window.confirm(`Are you sure you want to book ${slot.displayTime}?`);
                                                    if (confirmBooking) {

                                                        setSelectedTime(slot.displayTime);

                                                        try {

                                                            const token = localStorage.getItem("booking_app_token");
                                                            if (!token) {
                                                                alert("Missing auth token");
                                                                return;
                                                            }

                                                            await createBooking(
                                                                {
                                                                    id: slot.id,
                                                                    appointmentStart: new Date(slot.startTime),
                                                                    durationMinutes: 60,
                                                                },
                                                                token
                                                            );

                                                            alert("Booking confirmed!");
                                                        } catch (err: any) {
                                                            console.error(err);
                                                            alert(err.message);
                                                        }
                                                    }
                                                }}
                                                className={`p-2 rounded ${!slot.displayTime ? "invisible" : "hover:bg-blue-100"}`}
                                            >
                                                {slot.displayTime}
                                            </button>
                                        ))}
                                    </div>

                                ) : (
                                    <p className="text-gray-500">No available times.</p>
                                )}
                            </tbody>
                        </table>
                    </div>
                )}
            </div>
        </Layout>
    );
}