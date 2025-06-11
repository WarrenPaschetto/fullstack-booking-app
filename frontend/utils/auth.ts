import { jwtDecode } from "jwt-decode";

const TOKEN_KEY = "booking_app_token";

export function saveToken(token: string) {
    if (typeof window !== "undefined") {
        localStorage.setItem(TOKEN_KEY, token);
    }
}

export function getToken(): string | null {
    if (typeof window === "undefined") return null;
    return localStorage.getItem(TOKEN_KEY);
}

export function clearToken() {
    if (typeof window !== "undefined") {
        localStorage.removeItem(TOKEN_KEY);
    }
}

export interface DecodedJWT {
    sub: string;
    role: "user" | "admin";
    firstName: string;
    exp: number;
    iat: number;
}

export function getDecodedToken(): DecodedJWT | null {
    const token = getToken();
    if (!token) return null;
    try {
        const decoded = jwtDecode<DecodedJWT>(token);
        return decoded;
    } catch {
        return null;
    }
}

export function isAuthenticated(): boolean {
    const decoded = getDecodedToken();
    if (!decoded) return false;
    return decoded.exp * 1000 > Date.now();
}

export function userRole(): "user" | "admin" | null {
    const decoded = getDecodedToken();
    return decoded?.role || null;
}
