import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { SettingsView } from './SettingsView'

describe('SettingsView', () => {
  it('renders settings title', () => {
    render(<SettingsView />)
    expect(screen.getByText(/settings/i)).toBeInTheDocument()
  })

  it('displays appearance section', () => {
    render(<SettingsView />)
    expect(screen.getByText(/appearance/i)).toBeInTheDocument()
  })

  it('displays theme setting', () => {
    render(<SettingsView />)
    expect(screen.getByText('Theme')).toBeInTheDocument()
    expect(screen.getByText('Dark')).toBeInTheDocument()
  })

  it('displays data section', () => {
    render(<SettingsView />)
    expect(screen.getByText('Data')).toBeInTheDocument()
  })

  it('displays database path info', () => {
    render(<SettingsView />)
    expect(screen.getByText(/database/i)).toBeInTheDocument()
  })

  it('displays about section', () => {
    render(<SettingsView />)
    expect(screen.getByText(/about/i)).toBeInTheDocument()
  })

  it('displays version info', () => {
    render(<SettingsView />)
    expect(screen.getByText('Version')).toBeInTheDocument()
    expect(screen.getByText('1.0.0')).toBeInTheDocument()
  })

  it('displays default view setting', () => {
    render(<SettingsView />)
    expect(screen.getByText(/default view/i)).toBeInTheDocument()
  })
})
