type View = "userBookings" | "book";

interface Props {
    onSetView: (view: View) => void;
}

export default function UserToolbar({ onSetView }: Props) {
    return (
        <div className="flex flex-wrap gap-3 mb-6">
            <button
                className="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700"
                onClick={() => onSetView("userBookings")}
            >
                All Bookings
            </button>
            <button
                className="px-4 py-2 bg-yellow-600 text-white rounded hover:bg-yellow-700"
                onClick={() => onSetView("book")}
            >
                Book an Appointment
            </button>
        </div>
    );
}