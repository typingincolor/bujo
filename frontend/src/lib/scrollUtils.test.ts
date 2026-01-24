import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { hasScrollIntoView, scrollToElement, scrollToPosition, SCROLL_INTO_VIEW_DELAY_MS } from './scrollUtils'

describe('scrollUtils', () => {
  describe('hasScrollIntoView', () => {
    it('returns true for elements with scrollIntoView method', () => {
      const element = document.createElement('div')
      element.scrollIntoView = vi.fn()
      expect(hasScrollIntoView(element)).toBe(true)
    })

    it('returns false for elements without scrollIntoView method', () => {
      const element = {} as Element
      expect(hasScrollIntoView(element)).toBe(false)
    })
  })

  describe('scrollToElement', () => {
    beforeEach(() => {
      vi.useFakeTimers()
      Element.prototype.scrollIntoView = vi.fn()
    })

    afterEach(() => {
      vi.useRealTimers()
    })

    it('scrolls to element with default options', () => {
      const element = document.createElement('div')
      element.setAttribute('data-test', 'target')
      document.body.appendChild(element)

      scrollToElement('[data-test="target"]')
      vi.advanceTimersByTime(SCROLL_INTO_VIEW_DELAY_MS)

      expect(element.scrollIntoView).toHaveBeenCalledWith({
        behavior: 'smooth',
        block: 'center'
      })

      document.body.removeChild(element)
    })

    it('scrolls to element with custom options', () => {
      const element = document.createElement('div')
      element.setAttribute('data-test', 'target')
      document.body.appendChild(element)

      scrollToElement('[data-test="target"]', {
        delay: 200,
        behavior: 'auto',
        block: 'start'
      })
      vi.advanceTimersByTime(200)

      expect(element.scrollIntoView).toHaveBeenCalledWith({
        behavior: 'auto',
        block: 'start'
      })

      document.body.removeChild(element)
    })

    it('does nothing if element not found', () => {
      scrollToElement('[data-test="nonexistent"]')
      vi.advanceTimersByTime(SCROLL_INTO_VIEW_DELAY_MS)
      // Should not throw
    })

    it('does nothing if element lacks scrollIntoView', () => {
      const element = document.createElement('div')
      element.setAttribute('data-test', 'target')
      delete (element as any).scrollIntoView
      document.body.appendChild(element)

      scrollToElement('[data-test="target"]')
      vi.advanceTimersByTime(SCROLL_INTO_VIEW_DELAY_MS)
      // Should not throw

      document.body.removeChild(element)
    })
  })

  describe('scrollToPosition', () => {
    beforeEach(() => {
      vi.useFakeTimers()
      window.scrollTo = vi.fn()
    })

    afterEach(() => {
      vi.useRealTimers()
    })

    it('scrolls to position with default options using RAF', () => {
      scrollToPosition(500)

      // Advance through requestAnimationFrame
      vi.runAllTimers()

      expect(window.scrollTo).toHaveBeenCalledWith({
        top: 500,
        behavior: 'smooth'
      })
    })

    it('scrolls to position with custom behavior', () => {
      scrollToPosition(300, { behavior: 'auto' })

      vi.runAllTimers()

      expect(window.scrollTo).toHaveBeenCalledWith({
        top: 300,
        behavior: 'auto'
      })
    })

    it('uses requestAnimationFrame for smooth animation', () => {
      const rafSpy = vi.spyOn(window, 'requestAnimationFrame')

      scrollToPosition(100)

      expect(rafSpy).toHaveBeenCalled()

      rafSpy.mockRestore()
    })

    it('respects custom delay', () => {
      scrollToPosition(200, { delay: 500 })

      // Should not scroll immediately
      vi.advanceTimersByTime(100)
      expect(window.scrollTo).not.toHaveBeenCalled()

      // Should scroll after full delay
      vi.runAllTimers()
      expect(window.scrollTo).toHaveBeenCalled()
    })

    it('uses RAF immediately when delay is 0', () => {
      scrollToPosition(100, { delay: 0 })

      vi.runAllTimers()

      expect(window.scrollTo).toHaveBeenCalledWith({
        top: 100,
        behavior: 'smooth'
      })
    })
  })
})
