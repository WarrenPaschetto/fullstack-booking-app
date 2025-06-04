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
        // Full-screen wrapper that centers the form vertically & horizontally
        <div className="flex items-center justify-center">
            {/* The bluish–purple “card” */}
            <form
                onSubmit={handleSubmit}
                className="
          bg-gradient-to-br from-blue-500 to-purple-600
          rounded-2xl
          shadow-2xl
          p-8
          w-full max-w-md
          text-white
          flex flex-col
          space-y-6
        "
            >
                {/* Heading */}
                <h2 className="text-3xl font-bold text-center">
                    {mode === "login" ? "Log In" : "Register"}
                </h2>

                {/* If registering, show First & Last Name fields */}
                {mode === "register" && (
                    <div className="space-y-4">
                        <label className="block">
                            <span className="block mb-1 text-sm font-medium">First Name</span>
                            <input
                                type="text"
                                required
                                value={firstName}
                                onChange={(e) => setFirstName(e.target.value)}
                                className="
                  w-full
                  px-3 py-2
                  rounded-lg
                  focus:outline-none focus:ring-2 focus:ring-blue-300
                  text-gray-900
                "
                            />
                        </label>
                        <label className="block">
                            <span className="block mb-1 text-sm font-medium">Last Name</span>
                            <input
                                type="text"
                                required
                                value={lastName}
                                onChange={(e) => setLastName(e.target.value)}
                                className="
                  w-full
                  px-3 py-2
                  rounded-lg
                  focus:outline-none focus:ring-2 focus:ring-blue-300
                  text-gray-900
                "
                            />
                        </label>
                    </div>
                )}

                {/* Email field */}
                <label className="block">
                    <span className="block mb-1 text-sm font-medium">Email</span>
                    <input
                        type="email"
                        required
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        className="
              w-full
              px-3 py-2
              rounded-lg
              focus:outline-none focus:ring-2 focus:ring-blue-300
              text-gray-900
            "
                    />
                </label>

                {/* Password field */}
                <label className="block">
                    <span className="block mb-1 text-sm font-medium">Password</span>
                    <input
                        type="password"
                        required
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        className="
              w-full
              px-3 py-2
              rounded-lg
              focus:outline-none focus:ring-2 focus:ring-blue-300
              text-gray-900
            "
                    />
                </label>

                {/* Error message, if any */}
                {errorMsg && <p className="text-red-300 text-sm text-center">{errorMsg}</p>}

                {/* Submit button */}
                <button
                    type="submit"
                    disabled={loading}
                    className="
            w-full
            text-blue-700
            bg-blue-200 bg-opacity-20
            hover:bg-opacity-30
            py-2 rounded-lg
            font-semibold
            transition
            disabled:opacity-50
            hover:bg-blue-400
          "
                >
                    {loading ? "Processing…" : mode === "login" ? "Log In" : "Register"}
                </button>
            </form>
        </div>
    );
}