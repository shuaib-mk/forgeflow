import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import { api } from '../services/api'
import type { User } from '../types/models'

type AuthState = { user: User | null; loading: boolean; login(email: string, password: string): Promise<void>; logout(): Promise<void> }
const AuthContext = createContext<AuthState | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(Boolean(sessionStorage.getItem('forgeflow.token')))
  useEffect(() => { if (!loading) return; api.me().then(setUser).catch(() => sessionStorage.removeItem('forgeflow.token')).finally(() => setLoading(false)) }, [loading])
  const login = async (email: string, password: string) => { const session = await api.login(email, password); sessionStorage.setItem('forgeflow.token', session.token); setUser(session.user) }
  const logout = async () => { try { await api.logout() } finally { sessionStorage.removeItem('forgeflow.token'); setUser(null) } }
  return <AuthContext.Provider value={{ user, loading, login, logout }}>{children}</AuthContext.Provider>
}

export function useAuth() { const value = useContext(AuthContext); if (!value) throw new Error('useAuth must be used inside AuthProvider'); return value }

