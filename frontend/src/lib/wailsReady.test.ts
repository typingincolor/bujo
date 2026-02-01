import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { waitForWailsRuntime, isWailsReady, withRetry } from './wailsReady'

describe('wailsReady', () => {
  beforeEach(() => {
    ;(window as unknown as Record<string, unknown>).go = undefined
  })

  afterEach(() => {
    ;(window as unknown as Record<string, unknown>).go = undefined
    vi.useRealTimers()
  })

  describe('isWailsReady', () => {
    it('returns false when window.go is undefined', () => {
      expect(isWailsReady()).toBe(false)
    })

    it('returns false when window.go exists but wails.App is missing', () => {
      ;(window as unknown as { go: object }).go = {}
      expect(isWailsReady()).toBe(false)
    })

    it('returns true when window.go.wails.App exists', () => {
      ;(window as unknown as { go: { wails: { App: object } } }).go = { wails: { App: {} } }
      expect(isWailsReady()).toBe(true)
    })
  })

  describe('waitForWailsRuntime', () => {
    it('resolves immediately if Wails is already ready', async () => {
      ;(window as unknown as { go: { wails: { App: object } } }).go = { wails: { App: {} } }
      await expect(waitForWailsRuntime()).resolves.toBeUndefined()
    })

    it('waits and resolves when Wails becomes ready', async () => {
      vi.useFakeTimers()

      const promise = waitForWailsRuntime()

      setTimeout(() => {
        ;(window as unknown as { go: { wails: { App: object } } }).go = { wails: { App: {} } }
      }, 100)

      vi.advanceTimersByTime(150)
      await expect(promise).resolves.toBeUndefined()
    })

    it('rejects after timeout if Wails never becomes ready', async () => {
      vi.useFakeTimers()

      const promise = waitForWailsRuntime(500)

      vi.advanceTimersByTime(600)
      await expect(promise).rejects.toThrow('Wails runtime not available after 500ms')
    })
  })

  describe('withRetry', () => {
    beforeEach(() => {
      ;(window as unknown as { go: { wails: { App: object } } }).go = { wails: { App: {} } }
    })

    it('returns result on first success', async () => {
      const fn = vi.fn().mockResolvedValue('success')
      const result = await withRetry(fn)
      expect(result).toBe('success')
      expect(fn).toHaveBeenCalledTimes(1)
    })

    it('retries on failure and returns result on eventual success', async () => {
      const fn = vi.fn()
        .mockRejectedValueOnce(new Error('fail 1'))
        .mockRejectedValueOnce(new Error('fail 2'))
        .mockResolvedValue('success')

      const result = await withRetry(fn, 5, 10)
      expect(result).toBe('success')
      expect(fn).toHaveBeenCalledTimes(3)
    })

    it('throws after max attempts exceeded', async () => {
      const fn = vi.fn().mockRejectedValue(new Error('always fails'))

      await expect(withRetry(fn, 3, 10)).rejects.toThrow('always fails')
      expect(fn).toHaveBeenCalledTimes(3)
    })
  })
})
