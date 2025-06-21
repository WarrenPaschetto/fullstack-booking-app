import { useRouter } from "next/router";
import { useEffect } from "react";
import AuthForm from "../components/AuthForm";
import Layout from "../components/Layout";
import Navbar from "../components/Navbar";
import { isAuthenticated, saveToken, userRole } from "../utils/auth";

export default function LoginPage() {
    const router = useRouter();

    useEffect(() => {
        if (isAuthenticated()) {
            // If already logged in, redirect
            const role = userRole();
            router.replace(role === "admin" ? "/admin/dashboard" : "/user/dashboard");
        }
    }, [router]);

    function handleSuccess(token: string) {
        saveToken(token);
        const role = userRole(); // decode immediately
        if (role === "admin") {
            router.push("/admin/dashboard");
        } else {
            router.push("/user/dashboard");
        }
    }

    return (
        <Layout>
            <Navbar />
            <AuthForm mode="login" onSuccess={handleSuccess} />
        </Layout>
    );
}