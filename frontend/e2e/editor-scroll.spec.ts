import { test, expect } from '@playwright/test'

function generateLongDocument(lineCount: number): string {
  const lines: string[] = []
  for (let i = 1; i <= lineCount; i++) {
    lines.push(`. Task number ${i} - this is a test entry for scrolling`)
  }
  return lines.join('\n')
}

function mockWailsBindings(document: string) {
  return `
    window.go = {
      wails: {
        App: {
          GetDayEntries: () => Promise.resolve([{
            date: new Date().toISOString(),
            entries: [],
            context: { location: '', mood: '', weather: '' },
          }]),
          GetOverdue: () => Promise.resolve([]),
          GetHabits: () => Promise.resolve({ habits: [] }),
          GetLists: () => Promise.resolve([]),
          GetGoals: () => Promise.resolve([]),
          GetOutstandingQuestions: () => Promise.resolve([]),
          GetEditableDocumentWithEntries: () => Promise.resolve({
            document: ${JSON.stringify(document)},
            entries: [],
          }),
          GetEditableDocument: () => Promise.resolve(${JSON.stringify(document)}),
          ValidateEditableDocument: () => Promise.resolve([]),
          ApplyEditableDocument: () => Promise.resolve({ success: true }),
          GetEntryContext: () => Promise.resolve([]),
          GetVersion: () => Promise.resolve('test'),
        },
      },
    };

    // Mock Wails runtime - must include all methods called by runtime.js
    window.runtime = {
      LogPrint: () => {},
      LogTrace: () => {},
      LogDebug: () => {},
      LogInfo: () => {},
      LogWarning: () => {},
      LogError: () => {},
      LogFatal: () => {},
      EventsOnMultiple: () => () => {},
      EventsOn: () => () => {},
      EventsOff: () => {},
      EventsOffAll: () => {},
      EventsEmit: () => {},
    };
  `
}

test.describe('CodeMirror editor scrolling', () => {
  test('mouse wheel scrolls the editor content when document is taller than viewport', async ({ page }) => {
    const longDoc = generateLongDocument(100)

    await page.addInitScript(mockWailsBindings(longDoc))
    await page.goto('/')

    // Wait for the CodeMirror editor to render with content
    const cmScroller = page.locator('.cm-scroller')
    await expect(cmScroller).toBeVisible({ timeout: 10000 })

    // Wait for content to load
    await expect(page.locator('.cm-content')).toContainText('Task number 1')

    // Get scroll dimensions - if scrollHeight > clientHeight, content is scrollable
    const scrollInfo = await cmScroller.evaluate((el) => ({
      scrollTop: el.scrollTop,
      scrollHeight: el.scrollHeight,
      clientHeight: el.clientHeight,
    }))

    // Precondition: content must be taller than the visible area
    expect(scrollInfo.scrollHeight).toBeGreaterThan(scrollInfo.clientHeight)
    expect(scrollInfo.scrollTop).toBe(0)

    // Scroll down using mouse wheel
    await cmScroller.hover()
    await page.mouse.wheel(0, 300)

    // Wait for scroll to take effect
    await page.waitForTimeout(200)

    // Verify scroll position changed
    const scrollTopAfter = await cmScroller.evaluate((el) => el.scrollTop)
    expect(scrollTopAfter).toBeGreaterThan(0)
  })

  test('editor scroller height is constrained (not expanding to content height)', async ({ page }) => {
    const longDoc = generateLongDocument(100)

    await page.addInitScript(mockWailsBindings(longDoc))
    await page.goto('/')

    const cmScroller = page.locator('.cm-scroller')
    await expect(cmScroller).toBeVisible({ timeout: 10000 })
    await expect(page.locator('.cm-content')).toContainText('Task number 1')

    // The scroller's clientHeight (visible area) must be less than its scrollHeight (content)
    // If they're equal, the editor has expanded to fit all content and won't scroll
    const { scrollHeight, clientHeight } = await cmScroller.evaluate((el) => ({
      scrollHeight: el.scrollHeight,
      clientHeight: el.clientHeight,
    }))

    expect(clientHeight).toBeLessThan(scrollHeight)

    // Also verify the scroller isn't taking the entire viewport
    const viewportHeight = await page.evaluate(() => window.innerHeight)
    expect(clientHeight).toBeLessThan(viewportHeight)
  })
})
