import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import { api } from '../services/api'
import type { OrganizationMembership, User } from '../types/models'

type RegistrationInput = { email: string; displayName: string; password: string; organizationName: string; organizationSlug: string }
type AuthState = {
  user: User | null
  organizations: OrganizationMembership[]
  organizationId: string
  loading: boolean
  login(email: string, password: string): Promise<void>
  register(input: RegistrationInput): Promise<void>
  selectOrganization(id: string): void
  logout(): Promise<void>
}

const AuthContext = createContext<AuthState | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [organizations, setOrganizations] = useState<OrganizationMembership[]>([])
  const [organizationId, setOrganizationId] = useState(localStorage.getItem('forgeflow.organization') ?? '')
  const [loading, setLoading] = useState(true)

  const selectOrganization = (id: string) => {
    localStorage.setItem('forgeflow.organization', id)
    setOrganizationId(id)
  }

  const loadOrganizations = async (preferred = organizationId) => {
    const result = await api.organizations()
    setOrganizations(result.items)
    const selected = result.items.some((item) => item.organization.id === preferred) ? preferred : (result.items[0]?.organization.id ?? '')
    if (selected) selectOrganization(selected)
  }

  useEffect(() => {
    const token = sessionStorage.getItem('forgeflow.token')
    if (!token) {
      setLoading(false)
      return
    }
    Promise.all([api.me(), loadOrganizations()])
      .then(([currentUser]) => setUser(currentUser))
      .catch(() => {
        sessionStorage.removeItem('forgeflow.token')
        setUser(null)
      })
      .finally(() => setLoading(false))
    // Session restoration runs once when the application starts.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const login = async (email: string, password: string) => {
    const session = await api.login(email, password)
    sessionStorage.setItem('forgeflow.token', session.token)
    try {
      await loadOrganizations()
      setUser(session.user)
    } catch (error) {
      sessionStorage.removeItem('forgeflow.token')
      throw error
    }
  }

  const register = async (input: RegistrationInput) => {
    const registration = await api.register(input)
    selectOrganization(registration.organization.id)
    await login(input.email, input.password)
  }

  const logout = async () => {
    try {
      await api.logout()
    } finally {
      sessionStorage.removeItem('forgeflow.token')
      setUser(null)
      setOrganizations([])
    }
  }

  return (
    <AuthContext.Provider value={{ user, organizations, organizationId, loading, login, register, selectOrganization, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const value = useContext(AuthContext)
  if (!value) throw new Error('useAuth must be used inside AuthProvider')
  return value
}
