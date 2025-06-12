//import { Timestamp } from "next/dist/server/lib/cache-handlers/types";
import Layout from "../../components/Layout";
import Navbar from "../../components/Navbar";
import { useRequireAuth } from "../../utils/useRequireAuth";
import { FormEvent, useEffect, useState } from "react";
import DatePicker from "react-datepicker";
import "react-datepicker/dist/react-datepicker.css";
import { formatAppointment } from "../../utils/dateConversion";

interface Booking {
    ID: string;
    AppointmentStart: string;
    DurationMinutes: number;
}

//interface User {
//    ID: string;
//    FirstName: string;
//    LastName: string;
//    Email: string;
//}

type View = "allBookings" | "users" | "userBookings" | "patterns" | "update";

export default function AdminDashboard() {
    useRequireAuth("admin");

    const [view, setView] = useState<View>("allBookings");
    //const [users, setUsers] = useState<User[]>([]);
    const [allBookings, setAllBookings] = useState<Booking[]>([]);
    //const [selectedUserBookings, setSelectedUserBookings] = useState<Booking[]>([]);
    //const [patterns, setPatterns] = useState<any[]>([]);
    const [selectedBooking, setSelectedBooking] = useState<Booking | null>(null)
    const [dateValue, setDateValue] = useState<Date | null>(null);
    const [durationValue, setDurationValue] = useState<number>(60);
    const [bookingId, setBookingId] = useState("");

    const token = localStorage.getItem("booking_app_token");
    if (!token) return;

    //const API = process.env.NEXT_PUBLIC_BACKEND_URL || "http://localhost:8080";

    async function fetchAllBookings() {

        const resp = await fetch(`${process.env.NEXT_PUBLIC_BACKEND_URL || "http://localhost:8080"}/api/bookings/all`, {
            headers: { Authorization: `Bearer ${token}` },
        });
        if (resp.ok) {
            const data = await resp.json();
            setAllBookings(data);
        }
    }

    async function handleSubmit(e: FormEvent) {
        e.preventDefault();
        if (!selectedBooking || !dateValue) return;
        const newStart = dateValue.toISOString();
        setSelectedBooking({ ...selectedBooking, AppointmentStart: newStart, DurationMinutes: durationValue });
        const res = await fetch(
            `${process.env.NEXT_PUBLIC_BACKEND_URL || "http://localhost:8080"}/api/bookings/${bookingId}`,
            {
                method: "PUT",
                headers: {
                    "Content-Type": "application/json",
                    "Authorization": `Bearer ${token}`,
                },
                body: JSON.stringify({
                    appointment_start: newStart,
                    duration_minutes: durationValue,
                }),
            }
        );

        if (!res.ok) {
            const errText = await res.text();
            throw new Error(`Failed to update booking: ${res.status} ${errText}`);
        }

        // Refresh bookings table
        await fetchAllBookings();

        setView("allBookings");
        setSelectedBooking(null);
    }

    // Get all bookings upon startup
    useEffect(() => {
        fetchAllBookings().catch(console.error);
    }, []);

    // Select a booking in order to update or delete it
    useEffect(() => {
        if (!selectedBooking) return;
        const id = selectedBooking.ID.toString();
    }, [selectedBooking]);

    // Update tempBooking
    useEffect(() => {

    }, [selectedBooking])


    return (
        <Layout>
            <Navbar />
            <div className="w-full max-w-5xl mx-auto mt-8">
                <h2 className="text-2xl font-semibold mb-6">Admin Dashboard</h2>
                {/* toolbar */}
                <div className="flex flex-wrap gap-3">
                    <button
                        className="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700"
                        onClick={() => setView("allBookings")}
                    >
                        All Bookings
                    </button>
                    <button
                        className="px-4 py-2 bg-yellow-600 text-white rounded hover:bg-yellow-700"
                        onClick={() => setView("users")}
                    >
                        List All Users
                    </button>
                    <button
                        className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700"
                        onClick={() => setView("allBookings")}
                    >
                        Create Availability Pattern
                    </button>
                    <button
                        className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
                        onClick={() => setView("patterns")}
                    >
                        Availability Patterns
                    </button>
                </div>

                {/* PATTERNS view */}
                {view === "patterns" && (
                    <div className="bg-white shadow-md rounded-lg p-6">
                        <h3 className="text-xl mb-4">Availability Patterns</h3>
                        {/* ...same as before */}
                    </div>
                )}

                {/* Update view */}
                {view === "update" && (
                    <div className="flex items-center justify-center">
                        {/* The bluish–purple “card” */}
                        <form
                            onSubmit={handleSubmit}
                            className="bg-gradient-to-br from-blue-400 to-purple-700 rounded-2xl shadow-2xl p-8 w-full max-w-md
                          text-white flex flex-col space-y-6 m-10"
                        >
                            {/* Heading */}
                            <h2 className="text-3xl text-blue-50 font-bold text-center">
                                Update Booking
                            </h2>
                            <div className="space-y-4">
                                <label className="block  mb-1 text-lg font-medium text-blue-50">
                                    <span className="block"> From: {selectedBooking && formatAppointment(selectedBooking.AppointmentStart)}</span>
                                    <span>To: </span>
                                    <DatePicker
                                        selected={dateValue}
                                        onChange={date => setDateValue(date)}
                                        showTimeSelect
                                        timeIntervals={30}
                                        dateFormat="Pp"
                                        className="bg-blue-200 text-blue-800 font-bold text-lg ml-6"
                                    />
                                </label>
                                <label className="block">
                                    <span className="block mb-1 text-lg text-blue-50 font-medium">Duration in Minutes</span>
                                    <select
                                        required
                                        value={durationValue}
                                        onChange={e => setDurationValue(Number(e.target.value))}
                                        className="bg-blue-200 text-blue-800 font-bold text-lg ml-6"
                                    >
                                        <option value="" disabled>
                                            Select duration
                                        </option>
                                        <option value={30}>30</option>
                                        <option value={60}>60</option>
                                    </select>
                                </label>
                            </div>
                            {/* Error message, if any */}
                            {/*errorMsg && <p className="text-red-300 text-sm text-center">{errorMsg}</p>*/}

                            {/* Submit button */}
                            <button
                                type="submit"
                                className="w-full text-blue-700 bg-blue-200 bg-opacity-20 hover:bg-opacity-30 py-2 
                                rounded-lg font-semibold transition disabled:opacity-50 hover:bg-blue-400"
                            > Submit Changes
                            </button>
                        </form>
                    </div>
                )}


                {/* ALL BOOKINGS view */}
                {view === "allBookings" && (

                    <div className="bg-white shadow-md rounded-lg p-6">
                        <h3 className="text-xl font-medium mb-4">All Bookings</h3>
                        <table className="min-w-full bg-white">
                            <thead className="border-b">
                                <tr>
                                    <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">Booking ID</th>
                                    <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">Start Time</th>
                                    <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">Duration (min)</th>
                                </tr>
                            </thead>
                            <tbody>
                                {allBookings.map((b => {
                                    return (
                                        <tr key={b.ID} className={`border-b hover:bg-blue-100 ${selectedBooking == b ? "bg-blue-300" : "bg-white"}`} onClick={() => {
                                            setSelectedBooking((prev) =>
                                                prev?.ID === b.ID ? null : b
                                            )
                                            setBookingId(b.ID);
                                        }
                                        }>
                                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-800 ">{b.ID}</td>
                                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-800">
                                                {new Date(b.AppointmentStart).toLocaleString()}
                                            </td>
                                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-800">{b.DurationMinutes}</td>
                                        </tr>
                                    );
                                }
                                ))}
                            </tbody>
                        </table>
                        {selectedBooking && (
                            <div className="bg-white shadow-md rounded-lg p-6">
                                <button
                                    className="px-4 py-2 bg-green-600 text-white rounded"
                                    onClick={() => setView("update")}
                                >
                                    Update
                                </button>
                                <button
                                    className="px-4 py-2 bg-red-600 text-white rounded"
                                    onClick={() => setView("patterns")}
                                >
                                    Delete
                                </button>
                            </div>
                        )}
                    </div>
                )}
            </div>
        </Layout>
    );
}
