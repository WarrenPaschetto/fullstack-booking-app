import { useEffect } from "react";
import { useRouter } from "next/router";
import Layout from "../components/Layout";
import Link from "next/link";
import Navbar from "../components/Navbar";
import { isAuthenticated, userRole } from "../utils/auth";

export default function Home() {
    const router = useRouter();

    useEffect(() => {
        if (isAuthenticated()) {
            const role = userRole();
            if (role === "admin") {
                router.replace("/admin/dashboard");
            } else {
                router.replace("/user/dashboard");
            }
        }
        // otherwise, stay on this page
    }, [router]);

    return (
        <Layout>
            <Navbar />
            <div className="flex flex-col justify-center bg-blue-100 border-5 border-blue-200 p-8 rounded-lg shadow-lg max-w-md w-full text-center">
                <h1 className="text-2xl font-semibold mb-4 text-blue-800">Welcome to BookingApp</h1>
                <p className="mb-6 text-blue-600">
                    Please log in or register to continue.
                </p>
                <div className="flex justify-center space-x-4">
                    <Link
                        href="/login"
                        className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700"
                    >
                        Log In
                    </Link>
                    <Link
                        href="/register"
                        className="border border-blue-600 text-blue-600 px-4 py-2 rounded-md hover:bg-blue-50"
                    >
                        Register
                    </Link>
                </div>
            </div>
        </Layout>
    );
}