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
                    <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">Last Name</th>
                    <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">First Name</th>
                    <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">Email</th>
                </tr>
            </thead>
            <tbody>
                {filtered.map((u) => (
                    <tr
                        key={u.id}
                        className={`border-b hover:bg-blue-100 ${selectedUser?.id === u.id ? "bg-blue-300" : "bg-white"}`}
                        onClick={() => onSelectUser(u)}
                    >
                        <td className="px-6 py-4 text-sm text-gray-800">{u.last_name}</td>
                        <td className="px-6 py-4 text-sm text-gray-800">{u.first_name}</td>
                        <td className="px-6 py-4 text-sm text-gray-800">{u.email}</td>
                    </tr>
                ))}
            </tbody>
        </table>
    );
}