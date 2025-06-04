import { useRouter } from "next/router";
import { useEffect } from "react";
import { isAuthenticated, userRole } from "./auth";

export function useRequireAuth(allowedRole: "user" | "admin") {
    const router = useRouter();

    useEffect(() => {
        if (!isAuthenticated()) {
            router.replace("/login");
        } else {
            const role = userRole();
            if (role !== allowedRole) {
                // if a normal user tries to access /admin, send them away
                // or if admin tries /user
                router.replace(role === "admin" ? "/admin/dashboard" : "/user/calendar");
            }
        }
    }, [router, allowedRole]);
}