import { render, screen } from '@testing-library/react'
import { EmptyState, ErrorState, StatusPill } from './States'

describe('shared states', () => {
  it('announces errors accessibly', () => { render(<ErrorState error={new Error('Database unavailable')} />); expect(screen.getByRole('alert')).toHaveTextContent('Database unavailable') })
  it('renders actionable empty states', () => { render(<EmptyState title="No projects" body="Create one." action={<button>Create</button>} />); expect(screen.getByRole('button', { name: 'Create' })).toBeEnabled() })
  it('formats machine statuses for people', () => { render(<StatusPill status="in_progress" />); expect(screen.getByText('in progress')).toBeVisible() })
})

