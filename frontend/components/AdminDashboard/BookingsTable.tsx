import { Booking } from "@/utils/fetchAllBookings";

interface Props {
    bookings: Booking[];
    selectedBooking: Booking | null;
    onSelectBooking: (booking: Booking) => void;
}

export default function BookingsTable({ bookings, selectedBooking, onSelectBooking }: Props) {
    return (
        <table className="min-w-full bg-white">
            <thead className="border-b">
                <tr>
                    <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">Booking ID</th>
                    <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">Start Time</th>
                    <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">Duration (min)</th>
                </tr>
            </thead>
            <tbody>
                {bookings.map((b) => (
                    <tr
                        key={b.ID}
                        className={`border-b hover:bg-blue-100 ${selectedBooking?.ID === b.ID ? "bg-blue-300" : "bg-white"}`}
                        onClick={() => onSelectBooking(b)}
                    >
                        <td className="px-6 py-4 text-sm text-gray-800">{b.ID}</td>
                        <td className="px-6 py-4 text-sm text-gray-800">
                            {new Date(b.AppointmentStart).toLocaleString()}
                        </td>
                        <td className="px-6 py-4 text-sm text-gray-800">{b.DurationMinutes}</td>
                    </tr>
                ))}
            </tbody>
        </table>
    );
}