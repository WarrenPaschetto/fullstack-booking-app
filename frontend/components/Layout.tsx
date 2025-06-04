import { ReactNode } from "react";

export default function Layout({ children }: { children: ReactNode }) {
    return (
        // Full-screen wrapper to center the “card” vertically & horizontally
        <div className="min-h-screen flex flex-center items-center justify-center bg-gray-100">
            {/* The bluish–purple “card” container */}
            <div
                className="
          bg-gradient-to-br from-blue-500 to-purple-600
          rounded-2xl
          shadow-2xl
          p-6
          w-full max-w-md
          text-white
          flex flex-col
          space-y-6
        "
            >
                {children}
            </div>
        </div>
    );
}