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
        <nav className="w-full bg-blue-300 bg-opacity-30 shadow-md py-3 px-6 flex flex-col justify-between items-center">
            <Link href="/" className="font-bold text-3xl text-blue-800 py-2">
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
                <div className="space-x-8 flex flex-row justify-center w-full">
                    <Link href="/login" className="text-white hover:text-blue-700">
                        Log In
                    </Link>
                    <Link href="/register" className="text-white hover:text-blue-700">
                        Register
                    </Link>
                </div>
            )}
        </nav>
    );
}