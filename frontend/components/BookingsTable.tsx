import { Booking } from "@/utils/fetchAllBookings";

interface Props {
    bookings: Booking[];
    selectedBooking: Booking | null;
    onSelectBooking: (booking: Booking) => void;
}

export default function BookingsTable({ bookings, selectedBooking, onSelectBooking }: Props) {
    return (
        <table className="min-w-full bg-white">
            <thead className="border-b border-b-blue-800">
                <tr>
                    <th className="px-6 py-3 text-left text-xl font-semibold text-blue-800">Start Time</th>
                    <th className="px-6 py-3 text-left text-xl font-semibold text-blue-800">Duration (min)</th>
                </tr>
            </thead>
            <tbody>
                {Array.isArray(bookings) && bookings.length > 0 ? (
                    bookings.map((b) => (
                        <tr
                            key={b.ID}
                            className={`border-b border-b-blue-800 hover:bg-blue-100 ${selectedBooking?.ID === b.ID ? "bg-blue-300" : "bg-white"}`}
                            onClick={() => onSelectBooking(b)}
                        >
                            <td className="px-6 py-4 text-md font-medium text-gray-900">
                                {new Date(b.AppointmentStart).toLocaleString()}
                            </td>
                            <td className="px-6 py-4 text-md font-medium text-gray-900">{b.DurationMinutes}</td>
                        </tr>
                    ))
                ) : (
                    <tr>
                        <td colSpan={2} className="text-center text-gray-500">
                            No bookings found.
                        </td>
                    </tr>
                )}
            </tbody>
        </table>
    );
}