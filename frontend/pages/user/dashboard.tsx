import BookingsTable from "@/components/BookingsTable";
import Layout from "@/components/Layout";
import Navbar from "@/components/Navbar";
import UserCalendar from "@/components/UserDashboard/UserCalendar";
import UserToolbar from "@/components/UserDashboard/UserToolbar";
import { deleteBookingById } from "@/utils/deleteBookingById";
import { Booking } from "@/utils/fetchAllBookings";
import { fetchUserBookings } from "@/utils/fetchUserBookings";
import { formatError } from "@/utils/formatError";
import { useRequireAuth } from "@/utils/useRequireAuth";
import { useEffect, useState } from "react";


type View = "userBookings" | "book";

export default function UserDashboard() {
    useRequireAuth("user");

    const [errorMsg, setErrorMsg] = useState<string | undefined>(undefined);
    const [view, setView] = useState<View>("userBookings");
    const [allBookings, setAllBookings] = useState<Booking[]>([]);
    const [selectedBooking, setSelectedBooking] = useState<Booking | null>(null)
    const [token, setToken] = useState<string>("");

    useEffect(() => {
        const stored = localStorage.getItem("booking_app_token")
        if (stored) {
            setToken(stored)
        }
    }, [])

    async function fetchBookings() {
        try {
            const data = await fetchUserBookings(token);
            setAllBookings(data);
        } catch (err) {
            setErrorMsg(formatError(err));
        }
    };

    useEffect(() => {
        if (!token) return;
        fetchBookings();
    }, [token]);

    async function handleDeleteBooking() {
        if (!selectedBooking) return setErrorMsg("No booking selected");
        if (!window.confirm("Are you sure you want to delete this booking?")) return;

        try {
            await deleteBookingById(selectedBooking.ID, token);
            await fetchBookings();
            setSelectedBooking(null);
            setView("userBookings");
        } catch (err) {
            setErrorMsg(formatError(err));
        }
    }

    return (
        <Layout>
            <Navbar />
            <div className="w-full max-w-5xl mx-auto mt-8">
                <h2 className="text-2xl font-semibold mb-6">User Dashboard</h2>

                <UserToolbar onSetView={setView} />

                {errorMsg && (
                    <div className="mb-4 p-4 bg-red-100 border border-red-400 text-red-700 rounded">
                        {errorMsg}
                    </div>
                )}

                {view === "userBookings" && (
                    <div className="bg-white shadow-md rounded-lg p-6">
                        <h3 className="text-xl font-medium mb-4">All Bookings</h3>
                        <BookingsTable
                            bookings={allBookings}
                            selectedBooking={selectedBooking}
                            onSelectBooking={(booking: Booking) => setSelectedBooking(prev => prev?.ID === booking.ID ? null : booking)}
                        />
                        {selectedBooking && (
                            <div className="mt-4 space-x-4">
                                <button className="px-4 py-2 bg-red-600 text-white rounded" onClick={handleDeleteBooking}>Delete</button>
                            </div>
                        )}
                    </div>
                )}

                {view === "book" && (
                    <UserCalendar onBack={() => setView("userBookings")} />
                )}

            </div>
        </Layout>
    );
}

