import Link from "next/link";
import { useRouter } from "next/router";
import { clearToken, getDecodedToken, isAuthenticated } from "../utils/auth";

export default function Navbar() {
    const router = useRouter();
    const auth = isAuthenticated();
    const decoded = getDecodedToken();

    function handleLogout() {
        clearToken();
        router.push("/login");
    }

    return (
        <nav className="w-full bg-white shadow-md py-3 px-6 flex justify-between items-center">
            <Link href="/" className="font-bold text-xl">
                BookingApp
            </Link>

            {auth ? (
                <div className="flex items-center space-x-4">
                    <span className="text-gray-600">{decoded?.sub}</span>
                    <button
                        onClick={handleLogout}
                        className="text-red-500 hover:text-red-700 text-sm"
                    >
                        Log out
                    </button>
                </div>
            ) : (
                <div className="space-x-4">
                    <Link href="/login" className="text-blue-600 hover:underline">
                        Log In
                    </Link>
                    <Link href="/register" className="text-blue-600 hover:underline">
                        Register
                    </Link>
                </div>
            )}
        </nav>
    );
}