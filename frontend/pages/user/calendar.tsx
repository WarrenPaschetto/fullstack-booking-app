"use client"
import { useState, useEffect, useMemo } from "react";
import Layout from "../../components/Layout";
import Navbar from "../../components/Navbar";
import { useRequireAuth } from "../../utils/useRequireAuth";

export default function UserCalendar() {
    useRequireAuth("user");

    const today = new Date();
    const year = today.getFullYear();
    const month = today.getMonth(); // 0-indexed

    //–– STATE ––
    const [selectedDate, setSelectedDate] = useState<Date | null>(null);
    const [availableTimes, setAvailableTimes] = useState<string[]>([]);

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
        const iso = selectedDate.toISOString().slice(0, 10);
        fetch(`/api/availabilities/free?start=${iso}T00:00:00Z&end=${iso}T23:59:59Z`)
            .then((r) => r.json())
            .then((data) => {
                // assuming your API returns { times: string[] }
                setAvailableTimes(data.times || []);
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
                        <h2 className="text-2xl font-semibold mb-2">
                            Times on {selectedDate.toLocaleDateString()}
                        </h2>
                        {availableTimes.length > 0 ? (
                            <ul className="list-disc list-inside space-y-1">
                                {availableTimes.map((t) => (
                                    <li key={t}>{t}</li>
                                ))}
                            </ul>
                        ) : (
                            <p className="text-gray-500">No available times.</p>
                        )}
                    </div>
                )}
            </div>
        </Layout>
    );
}