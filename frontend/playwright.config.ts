import { defineConfig, devices } from '@playwright/test'
import * as path from 'path'
import * as fs from 'fs'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const projectRoot = path.resolve(__dirname, '..')
const testDbPath = path.resolve(__dirname, 'e2e', 'test.db')

console.log('Test DB Path:', testDbPath)

function cleanTestDatabase() {
  console.log('cleanTestDatabase called at:', new Date().toISOString())
  console.log('Test DB Path:', testDbPath)
  console.log('Test DB exists:', fs.existsSync(testDbPath))
  if (fs.existsSync(testDbPath)) {
    const stats = fs.statSync(testDbPath)
    console.log('Test DB size before delete:', stats.size)
    fs.unlinkSync(testDbPath)
    console.log('Test DB deleted')
  }
  const walPath = testDbPath + '-wal'
  const shmPath = testDbPath + '-shm'
  if (fs.existsSync(walPath)) {
    console.log('Deleting WAL file')
    fs.unlinkSync(walPath)
  }
  if (fs.existsSync(shmPath)) {
    console.log('Deleting SHM file')
    fs.unlinkSync(shmPath)
  }
}

console.log('Config loading at:', new Date().toISOString())
// Disabled for manual server testing
// cleanTestDatabase()

export default defineConfig({
  testDir: './e2e',
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  reporter: 'html',
  timeout: 10 * 1000,
  expect: {
    timeout: 3 * 1000,
  },
  use: {
    baseURL: 'http://localhost:34115',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    actionTimeout: 3 * 1000,
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
  webServer: {
    command: 'wails dev',
    cwd: projectRoot,
    url: 'http://localhost:34115',
    reuseExistingServer: true,
    timeout: 120 * 1000,
    env: {
      ...process.env,
      BUJO_DB_PATH: testDbPath,
    },
  },
})
