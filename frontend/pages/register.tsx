import { useRouter } from "next/router";
import { useEffect } from "react";
import AuthForm from "../components/AuthForm";
import Layout from "../components/Layout";
import Navbar from "../components/Navbar";
import { isAuthenticated, saveToken, userRole } from "../utils/auth";

export default function RegisterPage() {
    const router = useRouter();

    useEffect(() => {
        if (isAuthenticated()) {
            const role = userRole();
            router.replace(role === "admin" ? "/admin/dashboard" : "/user/calendar");
        }
    }, [router]);

    function handleSuccess(token: string) {
        saveToken(token);
        const role = userRole();
        if (role === "admin") {
            router.push("/admin/dashboard");
        } else {
            router.push("/user/calendar");
        }
    }

    return (
        <Layout>
            <Navbar />
            <AuthForm mode="register" onSuccess={handleSuccess} />
        </Layout>
    );
}