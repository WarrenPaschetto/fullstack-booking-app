import Layout from "../../components/Layout";
import Navbar from "../../components/Navbar";
import { useRequireAuth } from "../../utils/useRequireAuth";

export default function UserCalendar() {
    // Protect the route—only allow logged-in “user” roles
    useRequireAuth("user");

    return (
        <Layout>
            <Navbar />
            <div className="w-full max-w-3xl mx-auto mt-20 p-6 bg-white rounded-lg shadow-md text-center">
                <h1 className="text-3xl font-semibold mb-4">Calendar</h1>
                <p className="text-gray-600">Calendar page coming soon...</p>
            </div>
        </Layout>
    );
}