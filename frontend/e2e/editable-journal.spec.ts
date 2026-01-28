import { test, expect, Page } from '@playwright/test'
import { execFileSync } from 'child_process'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const TEST_DB_PATH = path.join(__dirname, 'test.db')

function clearDatabase(): void {
  execFileSync('sqlite3', [TEST_DB_PATH, 'DELETE FROM entries;'])
}

async function seedEntry(page: Page, content: string): Promise<void> {
  const contentArea = page.locator('.cm-content[contenteditable="true"]')
  await contentArea.fill(content)
  await page.keyboard.press('Meta+s')
  await page.waitForTimeout(300)
  await page.reload()
  await page.locator('.cm-editor').waitFor({ state: 'visible' })
}

async function focusEditor(page: Page): Promise<void> {
  const contentArea = page.locator('.cm-content[contenteditable="true"]')
  await contentArea.click()
}

async function setEditorContent(page: Page, content: string): Promise<void> {
  const contentArea = page.locator('.cm-content[contenteditable="true"]')
  await contentArea.fill(content)
}

async function getEditorContent(page: Page): Promise<string> {
  return await page.locator('.cm-editor').textContent() || ''
}

test.describe('Editable Journal View', () => {
  test.beforeEach(async ({ page }) => {
    // Clear entries from database before each test
    clearDatabase()

    await page.goto('/')

    // Dismiss any draft restoration banner if present
    const discardButton = page.getByRole('button', { name: 'Discard Draft' })
    if (await discardButton.isVisible({ timeout: 1000 }).catch(() => false)) {
      await discardButton.click()
    }

    await page.locator('.cm-editor').waitFor({ state: 'visible' })
  })

  test.describe('E1-E4: Basic Editing', () => {
    test('E1: edit content - user can modify entry text and save', async ({ page }) => {
      await seedEntry(page, '. Test task for editing')

      const editor = page.locator('.cm-editor')
      await expect(editor).toBeVisible()

      // Wait for editor to fully load content from backend
      await page.waitForTimeout(500)

      // Use fill() to replace content since keyboard navigation doesn't work reliably with CodeMirror
      await focusEditor(page)
      await setEditorContent(page, '. Test task for editing modified')

      await page.keyboard.press('Meta+s')
      await page.waitForTimeout(300)

      await page.reload()
      await page.locator('.cm-editor').waitFor({ state: 'visible' })

      await expect(page.locator('.cm-editor')).toContainText('modified')
    })

    test('E2: change entry type - changing symbol marks entry as done', async ({ page }) => {
      await seedEntry(page, '. Task to mark done')

      // Wait for editor to fully load content from backend
      await page.waitForTimeout(500)

      // Use fill() to replace the first character since keyboard navigation doesn't work with CodeMirror
      await focusEditor(page)
      await setEditorContent(page, 'x Task to mark done')

      await page.keyboard.press('Meta+s')
      await page.waitForTimeout(300)

      await page.reload()
      await page.locator('.cm-editor').waitFor({ state: 'visible' })

      const content = await page.locator('.cm-editor').textContent()
      expect(content).toMatch(/^x\s/)
    })

    test.describe('E2: Entry type change combinations', () => {
      const entryTypes = [
        { symbol: '.', name: 'task' },
        { symbol: '-', name: 'note' },
        { symbol: 'o', name: 'event' },
        { symbol: 'x', name: 'done' },
        { symbol: '~', name: 'cancelled' },
        { symbol: '?', name: 'question' },
        { symbol: '>', name: 'migrated' },
      ]

      async function testTypeChange(
        page: Page,
        fromSymbol: string,
        toSymbol: string,
        fromName: string,
        toName: string
      ) {
        await seedEntry(page, `${fromSymbol} Test ${fromName} entry`)

        // Wait for editor to fully load content from backend
        await page.waitForTimeout(500)

        // Use fill() to replace content since keyboard navigation doesn't work with CodeMirror
        await focusEditor(page)
        await setEditorContent(page, `${toSymbol} Test ${fromName} entry`)

        await page.keyboard.press('Meta+s')
        await page.waitForTimeout(300)

        await page.reload()
        await page.locator('.cm-editor').waitFor({ state: 'visible' })

        const content = await page.locator('.cm-editor').textContent()
        const regex = new RegExp(`^${toSymbol.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}\\s`)
        expect(content).toMatch(regex)
      }

      // Task to other types
      test('task to note', async ({ page }) => {
        await testTypeChange(page, '.', '-', 'task', 'note')
      })

      test('task to event', async ({ page }) => {
        await testTypeChange(page, '.', 'o', 'task', 'event')
      })

      test('task to done', async ({ page }) => {
        await testTypeChange(page, '.', 'x', 'task', 'done')
      })

      test('task to cancelled', async ({ page }) => {
        await testTypeChange(page, '.', '~', 'task', 'cancelled')
      })

      test('task to question', async ({ page }) => {
        await testTypeChange(page, '.', '?', 'task', 'question')
      })

      test('task to migrated', async ({ page }) => {
        await testTypeChange(page, '.', '>', 'task', 'migrated')
      })

      // Note to other types
      test('note to task', async ({ page }) => {
        await testTypeChange(page, '-', '.', 'note', 'task')
      })

      test('note to event', async ({ page }) => {
        await testTypeChange(page, '-', 'o', 'note', 'event')
      })

      test('note to done', async ({ page }) => {
        await testTypeChange(page, '-', 'x', 'note', 'done')
      })

      test('note to cancelled', async ({ page }) => {
        await testTypeChange(page, '-', '~', 'note', 'cancelled')
      })

      test('note to question', async ({ page }) => {
        await testTypeChange(page, '-', '?', 'note', 'question')
      })

      test('note to migrated', async ({ page }) => {
        await testTypeChange(page, '-', '>', 'note', 'migrated')
      })

      // Event to other types
      test('event to task', async ({ page }) => {
        await testTypeChange(page, 'o', '.', 'event', 'task')
      })

      test('event to note', async ({ page }) => {
        await testTypeChange(page, 'o', '-', 'event', 'note')
      })

      test('event to done', async ({ page }) => {
        await testTypeChange(page, 'o', 'x', 'event', 'done')
      })

      test('event to cancelled', async ({ page }) => {
        await testTypeChange(page, 'o', '~', 'event', 'cancelled')
      })

      test('event to question', async ({ page }) => {
        await testTypeChange(page, 'o', '?', 'event', 'question')
      })

      test('event to migrated', async ({ page }) => {
        await testTypeChange(page, 'o', '>', 'event', 'migrated')
      })

      // Done to other types
      test('done to task', async ({ page }) => {
        await testTypeChange(page, 'x', '.', 'done', 'task')
      })

      test('done to note', async ({ page }) => {
        await testTypeChange(page, 'x', '-', 'done', 'note')
      })

      test('done to event', async ({ page }) => {
        await testTypeChange(page, 'x', 'o', 'done', 'event')
      })

      test('done to cancelled', async ({ page }) => {
        await testTypeChange(page, 'x', '~', 'done', 'cancelled')
      })

      test('done to question', async ({ page }) => {
        await testTypeChange(page, 'x', '?', 'done', 'question')
      })

      test('done to migrated', async ({ page }) => {
        await testTypeChange(page, 'x', '>', 'done', 'migrated')
      })

      // Cancelled to other types
      test('cancelled to task', async ({ page }) => {
        await testTypeChange(page, '~', '.', 'cancelled', 'task')
      })

      test('cancelled to note', async ({ page }) => {
        await testTypeChange(page, '~', '-', 'cancelled', 'note')
      })

      test('cancelled to event', async ({ page }) => {
        await testTypeChange(page, '~', 'o', 'cancelled', 'event')
      })

      test('cancelled to done', async ({ page }) => {
        await testTypeChange(page, '~', 'x', 'cancelled', 'done')
      })

      test('cancelled to question', async ({ page }) => {
        await testTypeChange(page, '~', '?', 'cancelled', 'question')
      })

      test('cancelled to migrated', async ({ page }) => {
        await testTypeChange(page, '~', '>', 'cancelled', 'migrated')
      })

      // Question to other types
      test('question to task', async ({ page }) => {
        await testTypeChange(page, '?', '.', 'question', 'task')
      })

      test('question to note', async ({ page }) => {
        await testTypeChange(page, '?', '-', 'question', 'note')
      })

      test('question to event', async ({ page }) => {
        await testTypeChange(page, '?', 'o', 'question', 'event')
      })

      test('question to done', async ({ page }) => {
        await testTypeChange(page, '?', 'x', 'question', 'done')
      })

      test('question to cancelled', async ({ page }) => {
        await testTypeChange(page, '?', '~', 'question', 'cancelled')
      })

      test('question to migrated', async ({ page }) => {
        await testTypeChange(page, '?', '>', 'question', 'migrated')
      })

      // Migrated to other types
      test('migrated to task', async ({ page }) => {
        await testTypeChange(page, '>', '.', 'migrated', 'task')
      })

      test('migrated to note', async ({ page }) => {
        await testTypeChange(page, '>', '-', 'migrated', 'note')
      })

      test('migrated to event', async ({ page }) => {
        await testTypeChange(page, '>', 'o', 'migrated', 'event')
      })

      test('migrated to done', async ({ page }) => {
        await testTypeChange(page, '>', 'x', 'migrated', 'done')
      })

      test('migrated to cancelled', async ({ page }) => {
        await testTypeChange(page, '>', '~', 'migrated', 'cancelled')
      })

      test('migrated to question', async ({ page }) => {
        await testTypeChange(page, '>', '?', 'migrated', 'question')
      })
    })

    test.skip('E3: change priority - adding !!! shows priority label 1', async ({ page }) => {
      // TODO: Implement priority labels in CodeMirror decorations
      await seedEntry(page, '. Task for priority')

      const editor = page.locator('.cm-editor')
      await editor.click()

      await page.keyboard.press('End')
      await page.keyboard.type(' !!!')

      await expect(page.locator('.priority-label-1')).toBeVisible()

      await page.keyboard.press('Meta+s')

      await page.reload()
      await page.locator('.cm-editor').waitFor({ state: 'visible' })

      await expect(page.locator('.priority-label-1')).toBeVisible()
    })

    test('E4: create new entry - typing new line creates entry', async ({ page }) => {
      // Use fill() since keyboard.type() doesn't work reliably with CodeMirror
      await focusEditor(page)
      await setEditorContent(page, '. New task created by E2E test')

      await page.keyboard.press('Meta+s')
      await page.waitForTimeout(300)

      await page.reload()
      await page.locator('.cm-editor').waitFor({ state: 'visible' })

      await expect(page.locator('.cm-editor')).toContainText('New task created by E2E test')
    })
  })

  test.describe('E5-E7: Deletion Flow', () => {
    test.skip('E5: delete entry - deleted line shows in save dialog', async ({ page }) => {
      // TODO: Implement delete confirmation dialog
      await seedEntry(page, '. Entry to delete')

      const editor = page.locator('.cm-editor')
      await editor.click()

      await page.keyboard.press('Meta+Shift+k')

      await page.keyboard.press('Meta+s')

      const dialog = page.getByRole('dialog')
      await expect(dialog).toBeVisible()
      await expect(dialog).toContainText('delete')

      await page.getByRole('button', { name: /save/i }).click()

      await page.reload()
      await page.locator('.cm-editor').waitFor({ state: 'visible' })

      const content = await page.locator('.cm-editor').textContent()
      expect(content).not.toContain('Entry to delete')
    })

    test.skip('E6: migrate entry - using >[date] moves entry to target date', async ({ page }) => {
      // TODO: Implement migration date preview
      await seedEntry(page, '. Task to migrate')

      const editor = page.locator('.cm-editor')
      await editor.click()

      await page.keyboard.press('Home')
      await page.keyboard.press('Delete')
      await page.keyboard.type('>[tomorrow]')

      const preview = page.locator('.migration-date-preview')
      await expect(preview).toBeVisible()

      await page.keyboard.press('Meta+s')
    })

    test.skip('E7: restore deleted - unchecking in dialog restores entry', async ({ page }) => {
      // TODO: Implement delete confirmation dialog with checkboxes
      await seedEntry(page, '. Entry to restore')

      const editor = page.locator('.cm-editor')
      await editor.click()

      await page.keyboard.press('Meta+Shift+k')

      await page.keyboard.press('Meta+s')

      const dialog = page.getByRole('dialog')
      await expect(dialog).toBeVisible()

      await dialog.getByRole('checkbox').first().uncheck()
      await page.getByRole('button', { name: /save/i }).click()

      await expect(page.locator('.cm-editor')).toContainText('Entry to restore')
    })
  })

  test.describe('E8: Hierarchy', () => {
    test.skip('E8: indent/outdent - Tab indents entry and shows visual guide', async ({ page }) => {
      // TODO: Implement indent guides in CodeMirror
      await seedEntry(page, '. Parent task')

      const editor = page.locator('.cm-editor')
      await editor.click()

      await page.keyboard.press('End')
      await page.keyboard.press('Enter')
      await page.keyboard.type('. Child task')
      await page.keyboard.press('Tab')

      await expect(page.locator('.cm-indent-guide')).toBeVisible()

      await page.keyboard.press('Meta+s')

      await page.reload()
      await page.locator('.cm-editor').waitFor({ state: 'visible' })

      await expect(page.locator('.cm-editor')).toContainText('  . Child task')
    })
  })

  test.describe('E9-E11: Validation', () => {
    test.skip('E9: invalid syntax - unknown symbol shows red highlight and blocks save', async ({
      page,
    }) => {
      // TODO: Implement syntax validation and line error highlighting
      const editor = page.locator('.cm-editor')
      await editor.click()

      await page.keyboard.type('^ Invalid symbol line')

      await expect(page.locator('.cm-line-error')).toBeVisible()

      await page.keyboard.press('Meta+s')

      await expect(page.locator('.validation-error')).toBeVisible()
    })

    test.skip('E10: quick-fix - clicking suggestion corrects invalid line', async ({ page }) => {
      // TODO: Implement quick-fix suggestions
      const editor = page.locator('.cm-editor')
      await editor.click()

      await page.keyboard.type('^ Invalid symbol')

      const quickFix = page.locator('.quick-fix-button').first()
      await expect(quickFix).toBeVisible()
      await quickFix.click()

      await expect(page.locator('.cm-line-error')).not.toBeVisible()
    })

    test.skip('E11: warnings - past migration date shows warning but allows save', async ({
      page,
    }) => {
      // TODO: Implement line warning highlighting
      const editor = page.locator('.cm-editor')
      await editor.click()

      await page.keyboard.type('>[2020-01-01] Past migration')

      await expect(page.locator('.cm-line-warning')).toBeVisible()

      await page.keyboard.press('Meta+s')

      await expect(page.getByText(/saved/i)).toBeVisible()
    })
  })

  test.describe('E12-E13: Feedback', () => {
    test.skip('E12: unsaved indicator - editing shows dot, saving removes it', async ({ page }) => {
      // TODO: Implement unsaved indicator
      await expect(page.getByTestId('unsaved-indicator')).not.toBeVisible()

      const editor = page.locator('.cm-editor')
      await editor.click()
      await page.keyboard.type('. Test task')

      await expect(page.getByTestId('unsaved-indicator')).toBeVisible()

      await page.keyboard.press('Meta+s')

      await expect(page.getByTestId('unsaved-indicator')).not.toBeVisible()
    })

    test.skip('E13: save feedback - successful save shows confirmation in status bar', async ({
      page,
    }) => {
      // TODO: Implement save feedback in status bar
      const editor = page.locator('.cm-editor')
      await editor.click()
      await page.keyboard.type('. Test task')

      await page.keyboard.press('Meta+s')

      await expect(page.getByText(/saved/i)).toBeVisible()
    })
  })

  test.describe('E14: Crash Recovery', () => {
    test.skip('E14: crash recovery - localStorage draft restored after reload', async ({ page }) => {
      // TODO: Implement crash recovery UI
      const editor = page.locator('.cm-editor')
      await editor.click()
      await page.keyboard.type('. Unsaved crash test content')

      await page.waitForTimeout(600)

      await page.evaluate(() => {
        const keys = Object.keys(localStorage).filter((k) => k.startsWith('bujo.draft.'))
        if (keys.length === 0) {
          const today = new Date().toISOString().split('T')[0]
          localStorage.setItem(
            `bujo.draft.${today}`,
            JSON.stringify({
              document: '. Unsaved crash test content',
              deletedIds: [],
              timestamp: Date.now(),
            })
          )
        }
      })

      await page.reload()
      await page.locator('.cm-editor').waitFor({ state: 'visible' })

      const restorePrompt = page.getByText(/unsaved changes found/i)
      await expect(restorePrompt).toBeVisible()

      await page.getByRole('button', { name: /restore/i }).click()

      await expect(page.locator('.cm-editor')).toContainText('Unsaved crash test content')
    })
  })

  test.describe('E15: Event Sourcing', () => {
    test.skip('E15: event sourcing - deleted entry can be restored via CLI', async ({ page }) => {
      // TODO: This test requires CLI integration
      await seedEntry(page, '. Entry for event sourcing test')

      const editor = page.locator('.cm-editor')
      await editor.click()

      const contentBefore = await editor.textContent()
      const firstLine = contentBefore?.split('\n')[0] || ''

      await page.keyboard.press('Meta+Shift+k')
      await page.keyboard.press('Meta+s')

      const dialog = page.getByRole('dialog')
      if (await dialog.isVisible()) {
        await page.getByRole('button', { name: /save/i }).click()
      }

      await page.reload()
      await page.locator('.cm-editor').waitFor({ state: 'visible' })

      const contentAfter = await page.locator('.cm-editor').textContent()
      expect(contentAfter).not.toContain(firstLine)
    })
  })

  test.describe('E16: Help/Syntax Reference', () => {
    test.skip('E16: help - keyboard shortcuts popup shows syntax reference', async ({ page }) => {
      // TODO: Implement help dialog
      await page.keyboard.press('Meta+/')

      const popup = page.getByRole('dialog')
      await expect(popup).toBeVisible()

      await expect(popup).toContainText('.')
      await expect(popup).toContainText('task')
      await expect(popup).toContainText('-')
      await expect(popup).toContainText('note')
      await expect(popup).toContainText('>')
      await expect(popup).toContainText('migrate')
    })
  })

  test.describe('E17: File Import', () => {
    test.skip('E17: file import - Cmd+I opens file dialog and inserts content', async ({ page }) => {
      // TODO: Implement file import
      const editor = page.locator('.cm-editor')
      await editor.click()

      const [fileChooser] = await Promise.all([
        page.waitForEvent('filechooser'),
        page.keyboard.press('Meta+i'),
      ])

      await fileChooser.setFiles({
        name: 'test-import.txt',
        mimeType: 'text/plain',
        buffer: Buffer.from('. Imported task from file\n- Imported note'),
      })

      await expect(page.locator('.cm-editor')).toContainText('Imported task from file')
      await expect(page.locator('.cm-editor')).toContainText('Imported note')

      await page.keyboard.press('Meta+s')

      await page.reload()
      await page.locator('.cm-editor').waitFor({ state: 'visible' })

      await expect(page.locator('.cm-editor')).toContainText('Imported task from file')
    })
  })

  test.describe('Theme Support', () => {
    async function setTheme(page: Page, theme: 'light' | 'dark'): Promise<void> {
      // Navigate to settings
      await page.getByRole('button', { name: 'Settings' }).click()
      await page.waitForTimeout(200)

      // Click the theme button - labels are capitalized ('Light', 'Dark')
      const themeLabel = theme === 'light' ? 'Light' : 'Dark'
      await page.getByRole('button', { name: themeLabel, exact: true }).click()
      await page.waitForTimeout(200)

      // Navigate back to journal
      await page.getByRole('button', { name: 'Edit Journal' }).click()
      await page.locator('.cm-editor').waitFor({ state: 'visible' })
    }

    function parseRgb(rgbString: string): { r: number; g: number; b: number } {
      const match = rgbString.match(/rgb\((\d+),\s*(\d+),\s*(\d+)\)/)
      if (!match) {
        throw new Error(`Invalid RGB string: ${rgbString}`)
      }
      return {
        r: parseInt(match[1], 10),
        g: parseInt(match[2], 10),
        b: parseInt(match[3], 10),
      }
    }

    function isLightColor(rgb: { r: number; g: number; b: number }): boolean {
      // Calculate relative luminance - light colors have luminance > 0.5
      const luminance = (0.299 * rgb.r + 0.587 * rgb.g + 0.114 * rgb.b) / 255
      return luminance > 0.5
    }

    function isDarkColor(rgb: { r: number; g: number; b: number }): boolean {
      const luminance = (0.299 * rgb.r + 0.587 * rgb.g + 0.114 * rgb.b) / 255
      return luminance < 0.3
    }

    test('light mode - editor has light background and dark text', async ({ page }) => {
      await setTheme(page, 'light')

      const editor = page.locator('.cm-editor')
      const bgColor = await editor.evaluate((el) => getComputedStyle(el).backgroundColor)
      const textColor = await editor.evaluate((el) => getComputedStyle(el).color)

      const bgRgb = parseRgb(bgColor)
      const textRgb = parseRgb(textColor)

      expect(isLightColor(bgRgb)).toBe(true)
      expect(isDarkColor(textRgb)).toBe(true)
    })

    test('dark mode - editor has dark background and light text', async ({ page }) => {
      await setTheme(page, 'dark')

      const editor = page.locator('.cm-editor')
      const bgColor = await editor.evaluate((el) => getComputedStyle(el).backgroundColor)
      const textColor = await editor.evaluate((el) => getComputedStyle(el).color)

      const bgRgb = parseRgb(bgColor)
      const textRgb = parseRgb(textColor)

      expect(isDarkColor(bgRgb)).toBe(true)
      expect(isLightColor(textRgb)).toBe(true)
    })
  })
})
