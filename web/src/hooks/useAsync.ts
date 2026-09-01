import { useCallback, useEffect, useRef, useState } from 'react'

export function useAsync<T>(load: () => Promise<T>, dependencies: unknown[]) {
	const loader = useRef(load)
	loader.current = load
  const [data, setData] = useState<T | null>(null)
  const [error, setError] = useState<Error | null>(null)
  const [loading, setLoading] = useState(true)
	const dependencyKey = JSON.stringify(dependencies)
  const reload = useCallback(() => { setLoading(true); setError(null); loader.current().then(setData).catch(setError).finally(() => setLoading(false)) }, [])
  useEffect(reload, [reload, dependencyKey])
  return { data, error, loading, reload }
}
