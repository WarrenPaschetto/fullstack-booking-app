"use client"

import { useState, useEffect } from "react";
import Link from "next/link";
import { useRouter } from "next/router";
import { clearToken, getDecodedToken, isAuthenticated } from "../utils/auth";

export default function Navbar() {
    const router = useRouter();
    const [auth, setAuth] = useState(false);
    const [decoded, setDecoded] = useState<{ firstName: string } | null>(null);

    useEffect(() => {
        const ok = isAuthenticated();
        setAuth(ok);
        if (ok) setDecoded(getDecodedToken());
    }, []);

    function handleLogout() {
        clearToken();
        router.push("/login");
    }

    return (
        <nav className="w-full bg-blue-200 bg-opacity-30 shadow-md rounded-2xl py-3 px-6 flex flex-col lg:flex-row justify-between items-center">
            <Link href="/" className=" w-1/2 font-bold text-2xl sm:text-4xl text-blue-800 py-2 flex flex-row justify-items-start">
                BookingApp Demo
            </Link>
            <div>
                <Link href="https://github.com/WarrenPaschetto/fullstack-booking-app/tree/main" className=" w-1/2 font-bold text-xl sm:text-xl text-gray-800 hover:text-blue-800 py-2 flex flex-row justify-items-start">
                    Click here for demo instructions
                </Link>
            </div>

            {auth ? (
                <div className="flex items-center space-x-4">
                    <span className="text-gray-600 font-semibold text-lg ml-10">Welcome {decoded?.firstName}</span>
                    <button
                        onClick={handleLogout}
                        className="text-red-500 hover:text-red-700 font-semibold text-lg"
                    >
                        Log out
                    </button>
                </div>
            ) : (
                <div className="space-x-8 flex flex-row justify-center lg:justify-end w-1/2">
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