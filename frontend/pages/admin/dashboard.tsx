import Layout from "../../components/Layout";
import Navbar from "../../components/Navbar";
import { useRequireAuth } from "../../utils/useRequireAuth";
import { useEffect, useState } from "react";

interface Booking {
    ID: string;
    AppointmentStart: string;
    DurationMinutes: number;
}

interface User {
    ID: string;
    FirstName: string;
    LastName: string;
    Email: string;
}

type View = "allBookings" | "users" | "userBookings" | "patterns";

export default function AdminDashboard() {
    useRequireAuth("admin");

    const [view, setView] = useState<View>("allBookings");
    const [users, setUsers] = useState<User[]>([]);
    const [allBookings, setAllBookings] = useState<Booking[]>([]);
    const [selectedUserBookings, setSelectedUserBookings] = useState<Booking[]>([]);
    const [patterns, setPatterns] = useState<any[]>([]);

    //const API = process.env.NEXT_PUBLIC_BACKEND_URL || "http://localhost:8080";

    useEffect(() => {
        async function fetchAllBookings() {
            const token = localStorage.getItem("booking_app_token");
            if (!token) return;

            const resp = await fetch(`${process.env.NEXT_PUBLIC_BACKEND_URL || "http://localhost:8080"}/api/bookings/all`, {
                headers: { Authorization: `Bearer ${token}` },
            });
            if (resp.ok) {
                const data = await resp.json();
                setAllBookings(data);
            }
        }
        fetchAllBookings();
    }, []);



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
                                    console.log(b);
                                    return (
                                        <tr key={b.ID} className="border-b hover:bg-gray-50">
                                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-800">{b.ID}</td>
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
                    </div>
                )}
            </div>
        </Layout>
    );
}
