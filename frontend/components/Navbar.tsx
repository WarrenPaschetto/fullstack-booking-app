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
        <nav className="w-full bg-blue-200 bg-opacity-30 shadow-md rounded-2xl py-3 px-6 flex flex-col lg:flex-row justify-between items-center">
            <div className="flex flex-row justify-end w-full"></div>
            <Link href="/" className="font-bold text-2xl sm:text-4xl text-blue-800 py-2 flex flex-row justify-center w-full">
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
                <div className="space-x-8 flex flex-row justify-center lg:justify-end w-full">
                    <Link href="/login" className="font-bold text-blue-800 hover:text-white">
                        Log In
                    </Link>
                    <Link href="/register" className="font-bold text-blue-800 hover:text-white">
                        Register
                    </Link>
                </div>
            )}
        </nav>
    );
}