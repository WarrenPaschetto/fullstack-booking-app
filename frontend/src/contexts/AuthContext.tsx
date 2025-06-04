"use client";

import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { useRouter } from 'next/navigation'

// Define what a user object looks like
interface User {
    id: string
    email: string
    role: 'admin' | 'user'
}

// Shape of context value
interface AuthContextType {
    user: User | null
    loading: boolean
    login: (email: string, password: string) => Promise<void>
    register: (firstName: string, lastName: string, email: string, password: string) => Promise<void>
    logout: () => void
}

// Create context with placeholder defaults
const AuthContext = createContext<AuthContextType>({
    user: null,
    loading: true,
    login: async () => { },
    register: async () => { },
    logout: () => { },
})

export const AuthProvider = ({ children }: { children: ReactNode }) => {
    const router = useRouter()
    const [user, setUser] = useState<User | null>(null)
    const [loading, setLoading] = useState(true)

    // On mount, check for existing token
    useEffect(() => {
        const token = localStorage.getItem('token')
        if (token) {
            // Optionally fetch user profile here
            fetch('/api/me', {
                headers: { Authorization: `Bearer ${token}` },
            })
                .then(res => res.json())
                .then((data: User) => setUser(data))
                .catch(() => localStorage.removeItem('token'))
                .finally(() => setLoading(false))
        } else {
            setLoading(false)
        }
    }, [])

    // Login function
    const login = async (email: string, password: string) => {
        setLoading(true)
        const res = await fetch('/api/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password }),
        })
        if (!res.ok) throw new Error('Invalid credentials')
        const { token, user: userData } = await res.json()
        localStorage.setItem('token', token)
        setUser(userData)
        setLoading(false)
        // Redirect based on role
        router.push(userData.role === 'admin' ? '/admin' : '/calendar')
    }

    // Registration function
    const register = async (
        firstName: string,
        lastName: string,
        email: string,
        password: string
    ) => {
        setLoading(true)
        const res = await fetch('/api/register', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ first_name: firstName, last_name: lastName, email, password }),
        })
        if (!res.ok) throw new Error('Registration failed')
        // Automatically log in after registering
        await login(email, password)
        setLoading(false)
    }

    // Logout function
    const logout = () => {
        localStorage.removeItem('token')
        setUser(null)
        router.push('/login')
    }

    return (
        <AuthContext.Provider value={{ user, loading, login, register, logout }}>
            {children}
        </AuthContext.Provider>
    )
}

// Custom hook to consume auth context
export const useAuth = () => useContext(AuthContext)

/*
Usage:

In _app.tsx:
<AuthProvider>
  <Component {...pageProps} />
</AuthProvider>

In any component:
const { user, login, register, logout, loading } = useAuth()
*/
