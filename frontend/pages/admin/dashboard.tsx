import Layout from "../../components/Layout";
import Navbar from "../../components/Navbar";
import { useRequireAuth } from "../../utils/useRequireAuth";
import { useEffect, useState } from "react";

interface Booking {
    id: string;
    user_id: string;
    appointment_start: string;
    duration_minutes: number;
}

export default function AdminDashboard() {
    useRequireAuth("admin");

    const [allBookings, setAllBookings] = useState<Booking[]>([]);

    useEffect(() => {
        async function fetchAllBookings() {
            const token = localStorage.getItem("booking_app_token");
            if (!token) return;

            const resp = await fetch(`${process.env.NEXT_PUBLIC_BACKEND_URL || "http://localhost:8080"}/api/bookings`, {
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
                <div className="bg-white shadow-md rounded-lg p-6">
                    <h3 className="text-xl font-medium mb-4">All Bookings</h3>
                    <table className="min-w-full bg-white">
                        <thead className="border-b">
                            <tr>
                                <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">Booking ID</th>
                                <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">User ID</th>
                                <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">Start Time</th>
                                <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">Duration (min)</th>
                            </tr>
                        </thead>
                        <tbody>
                            {allBookings.map((b) => (
                                <tr key={b.id} className="border-b hover:bg-gray-50">
                                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-800">{b.id}</td>
                                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-800">{b.user_id}</td>
                                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-800">
                                        {new Date(b.appointment_start).toLocaleString()}
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-800">{b.duration_minutes}</td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            </div>
        </Layout>
    );
}
