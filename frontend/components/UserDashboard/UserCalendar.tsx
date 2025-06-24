import React, { useState, useEffect, useMemo } from "react";
import { useRequireAuth } from "../../utils/useRequireAuth";
import { createBooking } from "@/utils/createBookingApi";
import { FormattedSlot, listFreeSlots } from "@/utils/listFreeSlots";
import toast from "react-hot-toast";

interface UserCalendarProps {
    onBack?: () => void;
    onBookingSuccess?: () => void;
}

const UserCalendar: React.FC<UserCalendarProps> = ({ onBack, onBookingSuccess }) => {
    useRequireAuth("user");

    const today = new Date();
    const [month, setMonth] = useState(today.getMonth());
    const [year, setYear] = useState(today.getFullYear());

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

    // Helper functions to scroll through months
    const goToPreviousMonth = () => {
        if (month === 0) {
            setMonth(11);
            setYear((prev) => prev - 1);
        } else {
            setMonth((prev) => prev - 1);
        }
    };

    const goToNextMonth = () => {
        if (month === 11) {
            setMonth(0);
            setYear((prev) => prev + 1);
        } else {
            setMonth((prev) => prev + 1);
        }
    };

    //–– FETCH AVAILABILITY WHEN DAY SELECTED ––
    useEffect(() => {
        if (!selectedDate) return;
        const provider = "f2480f96-e1a3-4e33-9f26-b90910680bec";

        listFreeSlots(selectedDate, provider)
            .then((formatted: FormattedSlot[]) => {
                setAvailableTimes(formatted);
            })
            .catch(() => setAvailableTimes([]));
    }, [selectedDate]);

    return (
        <div className="w-full max-w-3xl mx-auto mt-20 p-6 bg-white rounded-lg shadow-md text-center">
            <h1 className="text-3xl font-semibold text-blue-800 mb-4">Calendar</h1>
            <p className="text-gray-900 font-medium">Pick a day to see availability</p>

            {/* Month & Year Header */}
            <div className="mt-4 text-blue-800 text-2xl px-2 font-semibold flex items-center justify-center gap-4">
                <button className="hover:text-green-600 text-2xl px-2" onClick={goToPreviousMonth}>&lt;</button>
                <span>{new Date(year, month).toLocaleString("default", { month: "long" })} {year}</span>
                <button className="hover:text-green-600 text-2xl px-2" onClick={goToNextMonth}>&gt;</button>
            </div>

            {/* Weekday Labels */}
            <div className="grid grid-cols-7 mt-2 text-xl font-semibold text-blue-800">
                {["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"].map((d) => (
                    <div key={d}>{d}</div>
                ))}
            </div>

            {/* Day Grid */}
            <div className="grid grid-cols-7 gap-2 mt-2">
                {calendarDays.map((date, idx) => {
                    const isToday = date?.toDateString() === new Date().toDateString();
                    const isSelected =
                        date &&
                        selectedDate &&
                        date.toDateString() === selectedDate.toDateString();

                    return (
                        <button
                            key={idx}
                            disabled={!date}
                            onClick={() => {
                                if (date) {
                                    setSelectedDate(date);
                                    setSelectedTime(""); // reset the selected time
                                }
                            }}
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
                    <h2 className="text-2xl font-semibold text-blue-800 mb-2">
                        Times on {selectedDate.toLocaleDateString()}
                    </h2>

                    {Array.isArray(availableTimes) && availableTimes.length > 0 ? (
                        <div className="grid grid-cols-4 gap-4 mt-4">
                            {availableTimes.map((slot) => (
                                <button
                                    key={slot.id}
                                    onClick={async () => {
                                        const confirmBooking = window.confirm(
                                            `Are you sure you want to book ${slot.displayTime}?`
                                        );
                                        if (confirmBooking) {
                                            setSelectedTime(slot.displayTime);

                                            try {
                                                const token = localStorage.getItem("booking_app_token");
                                                if (!token) {
                                                    toast.error("Missing auth token");
                                                    return;
                                                }

                                                await toast.promise(
                                                    createBooking(
                                                        {
                                                            id: slot.id,
                                                            appointmentStart: new Date(slot.startTime),
                                                            durationMinutes: 60,
                                                        },
                                                        token
                                                    ),
                                                    {
                                                        loading: "Booking...",
                                                        success: "Booking confirmed!",
                                                        error: (err) => err.message || "Booking failed.",
                                                    }
                                                );

                                                if (onBookingSuccess) {
                                                    await onBookingSuccess(); // make this async if needed
                                                }

                                                // Refresh available times after booking
                                                const provider = "f2480f96-e1a3-4e33-9f26-b90910680bec";
                                                listFreeSlots(selectedDate, provider)
                                                    .then((formatted) => setAvailableTimes(formatted))
                                                    .catch(() => setAvailableTimes([]));

                                            } catch (err) {
                                                toast.error("Something went wrong.");
                                                console.error(err);
                                            }

                                        }
                                    }}
                                    className={`p-2 rounded ${selectedTime === slot.displayTime
                                        ? "invisible"
                                        : "hover:bg-blue-100"
                                        }`}
                                >
                                    {slot.displayTime}
                                </button>
                            ))}
                        </div>
                    ) : (
                        <p className="text-gray-900">No available times.</p>
                    )}
                </div>
            )}
            {onBack && (
                <button
                    onClick={onBack}
                    className="text-blue-800 underline mb-4 hover:text-blue-500"
                >
                    ← Back to Dashboard
                </button>
            )}
        </div>
    );
}

export default UserCalendar;