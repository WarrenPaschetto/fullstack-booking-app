import { FormEvent, useState } from "react";

type AuthMode = "login" | "register";

interface AuthFormProps {
    mode: AuthMode;
    onSuccess: (token: string) => void;
}

// Helper: given an unknown value, return `message` if that property is a string.
function extractMessage(body: unknown): string | undefined {
    if (
        body !== null &&
        typeof body === "object" &&
        // “in” check ensures a “message” key exists
        "message" in (body as Record<string, unknown>) &&
        typeof (body as Record<string, unknown>).message === "string"
    ) {
        // Now it’s safe to cast to Record<string, string> for that property
        return (body as Record<string, string>).message;
    }
    return undefined;
}

export default function AuthForm({ mode, onSuccess }: AuthFormProps) {
    const [firstName, setFirstName] = useState("");
    const [lastName, setLastName] = useState("");
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [errorMsg, setErrorMsg] = useState<string | null>(null);
    const [loading, setLoading] = useState(false);

    async function handleSubmit(e: FormEvent) {
        e.preventDefault();
        setErrorMsg(null);
        setLoading(true);

        try {
            if (mode === "register") {
                const payload = { first_name: firstName, last_name: lastName, email, password };
                const resp = await fetch(
                    `${process.env.NEXT_PUBLIC_BACKEND_URL || "http://localhost:8080"}/api/register`,
                    {
                        method: "POST",
                        headers: { "Content-Type": "application/json" },
                        body: JSON.stringify(payload),
                    }
                );

                if (!resp.ok) {
                    // Parse response as unknown, then safely extract “message” if present.
                    const unknownBody: unknown = await resp.json().catch(() => ({} as unknown));
                    let errMsg = `Register failed: ${resp.status}`;
                    const parsed = extractMessage(unknownBody);
                    if (parsed) errMsg = parsed;
                    throw new Error(errMsg);
                }

                // On success, backend should return { token: "<JWT>" }
                const data = await resp.json();
                onSuccess(data.token);
            } else {
                // login mode
                const payload = { email, password };
                const resp = await fetch(
                    `${process.env.NEXT_PUBLIC_BACKEND_URL || "http://localhost:8080"}/api/login`,
                    {
                        method: "POST",
                        headers: { "Content-Type": "application/json" },
                        body: JSON.stringify(payload),
                    }
                );

                if (!resp.ok) {
                    const unknownBody: unknown = await resp.json().catch(() => ({} as unknown));
                    let errMsg = `Login failed: ${resp.status}`;
                    const parsed = extractMessage(unknownBody);
                    if (parsed) errMsg = parsed;
                    throw new Error(errMsg);
                }

                const data = await resp.json();
                onSuccess(data.token);
            }
        } catch (err: unknown) {
            setErrorMsg(err instanceof Error ? err.message : "An unexpected error occurred");
        } finally {
            setLoading(false);
        }
    }

    return (
        <form onSubmit={handleSubmit} className="bg-white rounded-lg shadow-lg p-8 w-full max-w-md">
            <h2 className="text-2xl font-semibold mb-6 text-center">
                {mode === "login" ? "Log In" : "Register"}
            </h2>

            {mode === "register" && (
                <>
                    <label className="block mb-2">
                        <span className="text-gray-700">First Name</span>
                        <input
                            type="text"
                            required
                            value={firstName}
                            onChange={(e) => setFirstName(e.target.value)}
                            className="mt-1 block w-full px-3 py-2 border rounded-md focus:outline-none focus:ring focus:border-blue-300"
                        />
                    </label>

                    <label className="block mb-2">
                        <span className="text-gray-700">Last Name</span>
                        <input
                            type="text"
                            required
                            value={lastName}
                            onChange={(e) => setLastName(e.target.value)}
                            className="mt-1 block w-full px-3 py-2 border rounded-md focus:outline-none focus:ring focus:border-blue-300"
                        />
                    </label>
                </>
            )}

            <label className="block mb-2">
                <span className="text-gray-700">Email</span>
                <input
                    type="email"
                    required
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    className="mt-1 block w-full px-3 py-2 border rounded-md focus:outline-none focus:ring focus:border-blue-300"
                />
            </label>

            <label className="block mb-4">
                <span className="text-gray-700">Password</span>
                <input
                    type="password"
                    required
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    className="mt-1 block w-full px-3 py-2 border rounded-md focus:outline-none focus:ring focus:border-blue-300"
                />
            </label>

            {errorMsg && <p className="text-red-500 mb-4 text-sm">{errorMsg}</p>}

            <button
                type="submit"
                disabled={loading}
                className="w-full bg-blue-600 text-white py-2 rounded-md hover:bg-blue-700 disabled:opacity-50"
            >
                {loading ? "Processing…" : mode === "login" ? "Log In" : "Register"}
            </button>
        </form>
    );
}
