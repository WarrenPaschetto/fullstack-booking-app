type View = "allBookings" | "users" | "userBookings" | "patterns" | "update";

interface Props {
    onSetView: (view: View) => void;
}

export default function AdminToolbar({ onSetView }: Props) {
    return (
        <div className="flex flex-wrap gap-3 mb-6">
            <button
                className="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700"
                onClick={() => onSetView("allBookings")}
            >
                All Bookings
            </button>
            <button
                className="px-4 py-2 bg-yellow-600 text-white rounded hover:bg-yellow-700"
                onClick={() => onSetView("users")}
            >
                List All Users
            </button>
            <button
                className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700"
                onClick={() => onSetView("patterns")}
            >
                Availability Patterns
            </button>
        </div>
    );
}