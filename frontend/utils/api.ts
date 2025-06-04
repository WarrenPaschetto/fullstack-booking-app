const BACKEND_URL = process.env.NEXT_PUBLIC_BACKEND_URL || "http://localhost:8080";

export async function post<T>(path: string, body: any, token?: string): Promise<T> {
    const res = await fetch(`${BACKEND_URL}${path}`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        body: JSON.stringify(body),
    });

    if (!res.ok) {
        const errorBody = await res.json().catch(() => ({}));
        throw new Error(errorBody.message || `API POST ${path} failed with status ${res.status}`);
    }

    return (await res.json()) as T;
}

export async function get<T>(path: string, token?: string): Promise<T> {
    const res = await fetch(`${BACKEND_URL}${path}`, {
        headers: {
            "Content-Type": "application/json",
            ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
    });

    if (!res.ok) {
        const errorBody = await res.json().catch(() => ({}));
        throw new Error(errorBody.message || `API GET ${path} failed with status ${res.status}`);
    }

    return (await res.json()) as T;
}