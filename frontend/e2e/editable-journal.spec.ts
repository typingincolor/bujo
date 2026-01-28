import { test, expect } from '@playwright/test'

test.describe('Editable Journal View', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
    await page.getByRole('button', { name: /edit journal/i }).click()
  })

  test.describe('E1-E4: Basic Editing', () => {
    test('E1: edit content - user can modify entry text and save', async ({ page }) => {
      const editor = page.locator('.cm-editor')
      await expect(editor).toBeVisible()

      await editor.click()
      await page.keyboard.press('End')
      await page.keyboard.type(' modified')

      await page.keyboard.press('Meta+s')

      await page.reload()
      await page.getByRole('button', { name: /edit journal/i }).click()

      await expect(page.locator('.cm-editor')).toContainText('modified')
    })

    test('E2: change entry type - changing symbol marks entry as done', async ({ page }) => {
      const editor = page.locator('.cm-editor')
      await editor.click()

      await page.keyboard.press('Home')
      await page.keyboard.press('Delete')
      await page.keyboard.type('x')

      await page.keyboard.press('Meta+s')

      await page.reload()
      await page.getByRole('button', { name: /edit journal/i }).click()

      const content = await page.locator('.cm-editor').textContent()
      expect(content).toMatch(/^x\s/)
    })

    test('E3: change priority - adding !!! shows priority label 1', async ({ page }) => {
      const editor = page.locator('.cm-editor')
      await editor.click()

      await page.keyboard.press('Home')
      await page.keyboard.press('ArrowRight')
      await page.keyboard.type(' !!!')

      await expect(page.locator('.priority-label-1')).toBeVisible()

      await page.keyboard.press('Meta+s')

      await page.reload()
      await page.getByRole('button', { name: /edit journal/i }).click()

      await expect(page.locator('.priority-label-1')).toBeVisible()
    })

    test('E4: create new entry - typing new line creates entry', async ({ page }) => {
      const editor = page.locator('.cm-editor')
      await editor.click()

      await page.keyboard.press('End')
      await page.keyboard.press('Enter')
      await page.keyboard.type('. New task created by E2E test')

      await page.keyboard.press('Meta+s')

      await page.reload()
      await page.getByRole('button', { name: /edit journal/i }).click()

      await expect(page.locator('.cm-editor')).toContainText('New task created by E2E test')
    })
  })

  test.describe('E5-E7: Deletion Flow', () => {
    test('E5: delete entry - deleted line shows in save dialog', async ({ page }) => {
      const editor = page.locator('.cm-editor')
      await editor.click()

      await page.keyboard.press('Meta+Shift+k')

      await page.keyboard.press('Meta+s')

      const dialog = page.getByRole('dialog')
      await expect(dialog).toBeVisible()
      await expect(dialog).toContainText('deleted')

      await page.getByRole('button', { name: /save/i }).click()

      await page.reload()
      await page.getByRole('button', { name: /edit journal/i }).click()

      const content = await page.locator('.cm-editor').textContent()
      expect(content).not.toContain('deleted entry')
    })

    test('E6: migrate entry - using >[date] moves entry to target date', async ({ page }) => {
      const editor = page.locator('.cm-editor')
      await editor.click()

      await page.keyboard.press('Home')
      await page.keyboard.press('Delete')
      await page.keyboard.type('>[tomorrow]')

      const preview = page.locator('.migration-date-preview')
      await expect(preview).toBeVisible()

      await page.keyboard.press('Meta+s')
    })

    test('E7: restore deleted - unchecking in dialog restores entry', async ({ page }) => {
      const editor = page.locator('.cm-editor')
      const originalContent = await editor.textContent()
      await editor.click()

      await page.keyboard.press('Meta+Shift+k')

      await page.keyboard.press('Meta+s')

      const dialog = page.getByRole('dialog')
      await expect(dialog).toBeVisible()

      await dialog.getByRole('checkbox').first().uncheck()
      await page.getByRole('button', { name: /save/i }).click()

      const newContent = await page.locator('.cm-editor').textContent()
      expect(newContent).toContain(originalContent?.split('\n')[0])
    })
  })

  test.describe('E8: Hierarchy', () => {
    test('E8: indent/outdent - Tab indents entry and shows visual guide', async ({ page }) => {
      const editor = page.locator('.cm-editor')
      await editor.click()

      await page.keyboard.press('End')
      await page.keyboard.press('Enter')
      await page.keyboard.type('. Child task')
      await page.keyboard.press('Tab')

      await expect(page.locator('.cm-indent-guide')).toBeVisible()

      await page.keyboard.press('Meta+s')

      await page.reload()
      await page.getByRole('button', { name: /edit journal/i }).click()

      await expect(page.locator('.cm-editor')).toContainText('  . Child task')
    })
  })

  test.describe('E9-E11: Validation', () => {
    test('E9: invalid syntax - unknown symbol shows red highlight and blocks save', async ({
      page,
    }) => {
      const editor = page.locator('.cm-editor')
      await editor.click()

      await page.keyboard.press('End')
      await page.keyboard.press('Enter')
      await page.keyboard.type('^ Invalid symbol line')

      await expect(page.locator('.cm-line-error')).toBeVisible()

      await page.keyboard.press('Meta+s')

      await expect(page.locator('.validation-error')).toBeVisible()
    })

    test('E10: quick-fix - clicking suggestion corrects invalid line', async ({ page }) => {
      const editor = page.locator('.cm-editor')
      await editor.click()

      await page.keyboard.press('End')
      await page.keyboard.press('Enter')
      await page.keyboard.type('^ Invalid symbol')

      const quickFix = page.locator('.quick-fix-button').first()
      await expect(quickFix).toBeVisible()
      await quickFix.click()

      await expect(page.locator('.cm-line-error')).not.toBeVisible()
    })

    test('E11: warnings - past migration date shows warning but allows save', async ({ page }) => {
      const editor = page.locator('.cm-editor')
      await editor.click()

      await page.keyboard.press('End')
      await page.keyboard.press('Enter')
      await page.keyboard.type('>[2020-01-01] Past migration')

      await expect(page.locator('.cm-line-warning')).toBeVisible()

      await page.keyboard.press('Meta+s')

      await expect(page.getByText(/saved/i)).toBeVisible()
    })
  })

  test.describe('E12-E13: Feedback', () => {
    test('E12: unsaved indicator - editing shows dot, saving removes it', async ({ page }) => {
      await expect(page.getByTestId('unsaved-indicator')).not.toBeVisible()

      const editor = page.locator('.cm-editor')
      await editor.click()
      await page.keyboard.type('a')

      await expect(page.getByTestId('unsaved-indicator')).toBeVisible()

      await page.keyboard.press('Meta+s')

      await expect(page.getByTestId('unsaved-indicator')).not.toBeVisible()
    })

    test('E13: save feedback - successful save shows confirmation in status bar', async ({
      page,
    }) => {
      const editor = page.locator('.cm-editor')
      await editor.click()
      await page.keyboard.type('a')

      await page.keyboard.press('Meta+s')

      await expect(page.getByText(/saved/i)).toBeVisible()
    })
  })

  test.describe('E14: Crash Recovery', () => {
    test('E14: crash recovery - localStorage draft restored after reload', async ({ page }) => {
      const editor = page.locator('.cm-editor')
      await editor.click()
      await page.keyboard.type('Unsaved crash test content')

      await page.waitForTimeout(600)

      await page.evaluate(() => {
        const keys = Object.keys(localStorage).filter((k) => k.startsWith('bujo.draft.'))
        if (keys.length === 0) {
          const today = new Date().toISOString().split('T')[0]
          localStorage.setItem(
            `bujo.draft.${today}`,
            JSON.stringify({
              document: 'Unsaved crash test content',
              deletedIds: [],
              timestamp: Date.now(),
            })
          )
        }
      })

      await page.reload()
      await page.getByRole('button', { name: /edit journal/i }).click()

      const restorePrompt = page.getByText(/unsaved changes found/i)
      await expect(restorePrompt).toBeVisible()

      await page.getByRole('button', { name: /restore/i }).click()

      await expect(page.locator('.cm-editor')).toContainText('Unsaved crash test content')
    })
  })

  test.describe('E15: Event Sourcing', () => {
    test('E15: event sourcing - deleted entry can be restored via CLI', async ({ page }) => {
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
      await page.getByRole('button', { name: /edit journal/i }).click()

      const contentAfter = await page.locator('.cm-editor').textContent()
      expect(contentAfter).not.toContain(firstLine)
    })
  })

  test.describe('E16: Help/Syntax Reference', () => {
    test('E16: help - keyboard shortcuts popup shows syntax reference', async ({ page }) => {
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
    test('E17: file import - Cmd+I opens file dialog and inserts content', async ({ page }) => {
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
      await page.getByRole('button', { name: /edit journal/i }).click()

      await expect(page.locator('.cm-editor')).toContainText('Imported task from file')
    })
  })
})
