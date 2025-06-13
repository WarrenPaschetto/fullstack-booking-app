import Layout from "@/components/Layout";
import Navbar from "@/components/Navbar";
import { useRequireAuth } from "@/utils/useRequireAuth";
import { FormEvent, useCallback, useEffect, useState } from "react";
import "react-datepicker/dist/react-datepicker.css";
import UpdateBookingForm from "@/components/UpdateBookingForm";
import { updateBooking } from "@/utils/updateBookingApi";

interface Booking {
    ID: string;
    AppointmentStart: string;
    DurationMinutes: number;
}

interface User {
    id: string;
    first_name: string;
    last_name: string;
    email: string;
    user_role: string;
}

type View = "allBookings" | "users" | "userBookings" | "patterns" | "update";

export default function AdminDashboard() {
    useRequireAuth("admin");

    const [errorMsg, setErrorMsg] = useState<string | undefined>(undefined)
    const [view, setView] = useState<View>("allBookings");
    const [allUsers, setAllUsers] = useState<User[]>([]);
    const [selectedUser, setSelectedUser] = useState<User | null>(null);
    const [allBookings, setAllBookings] = useState<Booking[]>([]);
    //const [patterns, setPatterns] = useState<any[]>([]);
    const [selectedBooking, setSelectedBooking] = useState<Booking | null>(null)
    const [dateValue, setDateValue] = useState<Date | null>(null);
    const [durationValue, setDurationValue] = useState<number>(60);
    const [token, setToken] = useState<string>("");

    useEffect(() => {
        const stored = localStorage.getItem("booking_app_token")
        if (stored) {
            setToken(stored)
        }
    }, [])

    //const API = process.env.NEXT_PUBLIC_BACKEND_URL || "http://localhost:8080";

    const fetchAllBookings = useCallback(async () => {
        const resp = await fetch(
            `${process.env.NEXT_PUBLIC_BACKEND_URL}/api/bookings/all`,
            { headers: { Authorization: `Bearer ${token}` } }
        );
        if (resp.ok) {
            const data = await resp.json();
            setAllBookings(data);
        }
    }, [token]);

    useEffect(() => {
        if (view !== "users" || !token) return;

        (async () => {
            setErrorMsg("");
            try {
                const resp = await fetch(
                    `${process.env.NEXT_PUBLIC_BACKEND_URL || "http://localhost:8080"}/api/users/all`,
                    { headers: { Authorization: `Bearer ${token}` } }
                );
                if (!resp.ok) {
                    const text = await resp.text();
                    throw new Error(`Error ${resp.status}: ${text}`);
                }
                const data: User[] = await resp.json();
                console.log(data);
                setAllUsers(data);
            } catch (err: unknown) {
                console.error(err);
                setErrorMsg(err instanceof Error ? err.message : String(err));
            }
        })();
    }, [view, token]);

    // Admin delete booking by ID
    async function handleDeleteBooking() {
        setErrorMsg(undefined);

        if (!selectedBooking) {
            setErrorMsg("No booking selected");
            return;
        }

        if (!window.confirm("Are you sure you want to delete this booking?")) {
            return;
        }

        try {
            const res = await fetch(
                `${process.env.NEXT_PUBLIC_BACKEND_URL}/api/bookings/${selectedBooking.ID}`,
                {
                    method: "DELETE",
                    headers: { Authorization: `Bearer ${token}` }
                }
            );

            if (!res.ok) {
                const text = await res.text();
                throw new Error(`Delete failed: ${res.status} ${text}`);
            }

            await fetchAllBookings();
            setView("allBookings");
            setSelectedBooking(null);
        } catch (err: unknown) {
            if (err instanceof Error) {
                setErrorMsg(err.message);
            } else {
                setErrorMsg(String(err));
            }
        }

    }

    // Get all bookings and users upon startup
    useEffect(() => {
        if (token === "") return;
        fetchAllBookings().catch(console.error);
    }, [token, fetchAllBookings]);


    async function handleUpdateSubmit(e: FormEvent) {
        e.preventDefault()

        setErrorMsg(undefined)
        if (!selectedBooking || !dateValue) return

        try {
            await updateBooking(
                { id: selectedBooking.ID, appointmentStart: dateValue, durationMinutes: durationValue },
                token
            )
            await fetchAllBookings()
            setView("allBookings")
            setSelectedBooking(null)
        } catch (err: unknown) {
            if (err instanceof Error) {
                setErrorMsg(err.message);
            } else {
                // fallback for non-Error throws
                setErrorMsg(String(err));
            }
        }
    }

    // Select a booking in order to update or delete it
    //useEffect(() => {
    //    if (!selectedBooking) return;
    //    const id = selectedBooking.ID.toString();
    //}, [selectedBooking]);

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
                {view === "update" && selectedBooking && (
                    <UpdateBookingForm
                        key={selectedBooking.ID}
                        selectedBooking={selectedBooking}
                        dateValue={dateValue}
                        setDateValue={setDateValue}
                        durationValue={durationValue}
                        setDurationValue={setDurationValue}
                        onSubmit={handleUpdateSubmit}
                        errorMsg={errorMsg}
                    />
                )}

                {/* ALL USERS view */}
                {view === "users" && (
                    <div className="bg-white shadow-md rounded-lg p-6">
                        <h3 className="text-xl font-medium mb-4">Current Users</h3>

                        {errorMsg && (
                            <p className="text-red-500 mb-4">{errorMsg}</p>
                        )}

                        {allUsers.length === 0 ? (
                            <p>No users found.</p>
                        ) : (
                            <table className="min-w-full bg-white">
                                <thead className="border-b">
                                    <tr>
                                        <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">Last Name</th>
                                        <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">First Name</th>
                                        <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">Email</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {allUsers.filter(u => (u.user_role !== "admin")).map((u => {
                                        return (
                                            <tr key={u.id} className={`border-b hover:bg-blue-100 ${selectedUser == u ? "bg-blue-300" : "bg-white"}`} onClick={() => {
                                                setSelectedUser((prev) =>
                                                    prev?.id === u.id ? null : u
                                                )
                                            }
                                            }>
                                                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-800 ">{u.last_name}</td>
                                                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-800">
                                                    {u.first_name}
                                                </td>
                                                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-800">{u.email}</td>
                                            </tr>
                                        );
                                    }
                                    ))}
                                </tbody>
                            </table>
                        )}
                        {selectedUser && (
                            <div className="bg-white shadow-md rounded-lg p-6">
                                <button
                                    className="px-4 py-2 bg-green-600 text-white rounded"
                                    onClick={() => setView("update")}
                                >
                                    Update
                                </button>
                                <button
                                    className="px-4 py-2 bg-blue-600 text-white rounded"
                                    onClick={() => setView("update")}
                                >
                                    Create Appt.
                                </button>
                                <button
                                    className="px-4 py-2 bg-red-600 text-white rounded"
                                    onClick={handleDeleteBooking}
                                >
                                    Delete
                                </button>
                            </div>
                        )}
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
                                    onClick={handleDeleteBooking}
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
