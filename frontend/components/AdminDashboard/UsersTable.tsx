import { User } from "@/utils/fetchAllUsers";

interface Props {
    users: User[];
    selectedUser: User | null;
    onSelectUser: (user: User) => void;
}

export default function UsersTable({ users, selectedUser, onSelectUser }: Props) {
    const filtered = users.filter((u) => u.user_role !== "admin");

    return (
        <table className="min-w-full bg-white">
            <thead className="border-b">
                <tr>
                    <th className="px-6 py-3 text-left text-xl font-semibold text-blue-800">Last Name</th>
                    <th className="px-6 py-3 text-left text-xl font-semibold text-blue-800">First Name</th>
                    <th className="px-6 py-3 text-left text-xl font-semibold text-blue-800">Email</th>
                </tr>
            </thead>
            <tbody>
                {filtered.map((u) => (
                    <tr
                        key={u.id}
                        className={`border-b border-b-blue-800 hover:bg-blue-100 ${selectedUser?.id === u.id ? "bg-blue-300" : "bg-white"}`}
                        onClick={() => onSelectUser(u)}
                    >
                        <td className="px-6 py-4 text-md font-medium text-gray-900">{u.last_name}</td>
                        <td className="px-6 py-4 text-md font-medium text-gray-900">{u.first_name}</td>
                        <td className="px-6 py-4 text-md font-medium text-gray-900">{u.email}</td>
                    </tr>
                ))}
            </tbody>
        </table>
    );
}