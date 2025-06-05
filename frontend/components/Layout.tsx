import { ReactNode } from "react";

export default function Layout({ children }: { children: ReactNode }) {
    return (
        // Full-screen wrapper to center the “card” vertically & horizontally
        <div
            className="min-h-screen flex items-center justify-center bg-cover bg-center"
            style={{
                backgroundImage: `url('/images/blue-abstract.jpg')`,
            }}>
            <div
                className="
          border-5
          border-blue-200
          bg-gradient-to-br from-blue-900 via-blue-700 to-purple-800
          rounded-2xl
          shadow-blue-800
          shadow-2xl
          p-2
          w-[60%]
          text-black
          flex flex-col
          items-center
          space-y-6
        "
            >
                {children}
            </div>
        </div>
    );
}