import Layout from "@/components/Layout";
import Navbar from "@/components/Navbar";
import { useRequireAuth } from "@/utils/useRequireAuth";
import { FormEvent, useCallback, useEffect, useState } from "react";
import "react-datepicker/dist/react-datepicker.css";
import UpdateBookingForm from "@/components/UpdateBookingForm";
import { updateBooking } from "@/utils/updateBookingApi";
import AvailabilityPatternForm from "@/components/AdminDashboard/AvailabilityPatternForm";
import { availabilityPattern } from "@/utils/availPatternApi";
import { Booking, fetchAllBookings } from "@/utils/fetchAllBookings";
import { fetchAllUsers, User } from "@/utils/fetchAllUsers";
import { formatError } from "@/utils/formatError";
import { deleteBookingById } from "@/utils/deleteBookingById";
import BookingsTable from "@/components/AdminDashboard/BookingsTable";
import UsersTable from "@/components/AdminDashboard/UsersTable";
import AdminToolbar from "@/components/AdminDashboard/AdminTollbar";

type View = "allBookings" | "users" | "userBookings" | "patterns" | "update";

export default function AdminDashboard() {
    useRequireAuth("admin");

    const [errorMsg, setErrorMsg] = useState<string | undefined>(undefined)
    const [view, setView] = useState<View>("allBookings");
    const [allUsers, setAllUsers] = useState<User[]>([]);
    const [selectedUser, setSelectedUser] = useState<User | null>(null);
    const [allBookings, setAllBookings] = useState<Booking[]>([]);
    const [selectedBooking, setSelectedBooking] = useState<Booking | null>(null)
    const [dateValue, setDateValue] = useState<Date | null>(null);
    const [durationValue, setDurationValue] = useState<number>(60);
    const [token, setToken] = useState<string>("");
    const [weekValue, setWeekValue] = useState<number>(0);
    const [startTimeValue, setStartTimeValue] = useState<Date | null>(null)
    const [endTimeValue, setEndTimeValue] = useState<Date | null>(null)

    useEffect(() => {
        const stored = localStorage.getItem("booking_app_token")
        if (stored) {
            setToken(stored)
        }
    }, [])

    const fetchBookings = useCallback(async () => {
        try {
            const data = await fetchAllBookings(token);
            setAllBookings(data);
        } catch (err) {
            setErrorMsg(formatError(err));
        }
    }, [token]);

    useEffect(() => {
        if (token) fetchBookings();
    }, [token, fetchBookings]);

    useEffect(() => {
        if (view !== "users" || !token) return;

        (async () => {
            setErrorMsg("");
            try {
                const data = await fetchAllUsers(token);
                setAllUsers(data);
            } catch (err) {
                setErrorMsg(formatError(err));
            }
        })();
    }, [view, token]);

    async function handleDeleteBooking() {
        if (!selectedBooking) return setErrorMsg("No booking selected");
        if (!window.confirm("Are you sure you want to delete this booking?")) return;

        try {
            await deleteBookingById(selectedBooking.ID, token);
            await fetchBookings();
            setSelectedBooking(null);
            setView("allBookings");
        } catch (err) {
            setErrorMsg(formatError(err));
        }
    }

    // Get all bookings and users upon startup
    useEffect(() => {
        if (token === "") return;
        fetchAllBookings(token).catch(console.error);
    }, [token, fetchAllBookings]);


    async function handleUpdateSubmit(e: FormEvent) {
        e.preventDefault();
        setErrorMsg(undefined);
        if (!selectedBooking || !dateValue) return;

        try {
            await updateBooking(
                { id: selectedBooking.ID, appointmentStart: dateValue, durationMinutes: durationValue },
                token
            );
            await fetchBookings();
            setSelectedBooking(null);
            setView("allBookings");
        } catch (err) {
            setErrorMsg(formatError(err));
        }
    }

    async function handleAvailPatternSubmit(e: FormEvent) {
        e.preventDefault();
        setErrorMsg(undefined);
        if (!startTimeValue || !endTimeValue) return;

        try {
            await availabilityPattern({ dayOfWeek: weekValue, startTime: startTimeValue, endTime: endTimeValue }, token);
        } catch (err) {
            setErrorMsg(formatError(err));
        }
    }

    return (
        <Layout>
            <Navbar />
            <div className="w-full max-w-5xl mx-auto mt-8">
                <h2 className="text-2xl font-semibold mb-6">Admin Dashboard</h2>

                <AdminToolbar onSetView={setView} />

                {view === "patterns" && (
                    <AvailabilityPatternForm
                        weekValue={weekValue}
                        setWeekValue={setWeekValue}
                        startTimeValue={startTimeValue}
                        setStartTimeValue={setStartTimeValue}
                        endTimeValue={endTimeValue}
                        setEndTimeValue={setEndTimeValue}
                        onSubmit={handleAvailPatternSubmit}
                        errorMsg={errorMsg}
                    />
                )}

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

                {view === "users" && (
                    <div className="bg-white shadow-md rounded-lg p-6">
                        <h3 className="text-xl font-medium mb-4">Current Users</h3>
                        {errorMsg && <p className="text-red-500 mb-4">{errorMsg}</p>}
                        <UsersTable
                            users={allUsers}
                            selectedUser={selectedUser}
                            onSelectUser={(user: User) => setSelectedUser(prev => prev?.id === user.id ? null : user)}
                        />
                        {selectedUser && (
                            <div className="mt-4 space-x-4">
                                <button className="px-4 py-2 bg-green-600 text-white rounded" onClick={() => setView("update")}>Update</button>
                                <button className="px-4 py-2 bg-blue-600 text-white rounded" onClick={() => setView("update")}>Create Appt.</button>
                                <button className="px-4 py-2 bg-red-600 text-white rounded" onClick={handleDeleteBooking}>Delete</button>
                            </div>
                        )}
                    </div>
                )}

                {view === "allBookings" && (
                    <div className="bg-white shadow-md rounded-lg p-6">
                        <h3 className="text-xl font-medium mb-4">All Bookings</h3>
                        <BookingsTable
                            bookings={allBookings}
                            selectedBooking={selectedBooking}
                            onSelectBooking={(booking: Booking) => setSelectedBooking(prev => prev?.ID === booking.ID ? null : booking)}
                        />
                        {selectedBooking && (
                            <div className="mt-4 space-x-4">
                                <button className="px-4 py-2 bg-green-600 text-white rounded" onClick={() => setView("update")}>Update</button>
                                <button className="px-4 py-2 bg-red-600 text-white rounded" onClick={handleDeleteBooking}>Delete</button>
                            </div>
                        )}
                    </div>
                )}
            </div>
        </Layout>
    );
}
