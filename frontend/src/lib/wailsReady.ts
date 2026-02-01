declare global {
  interface Window {
    go?: {
      wails?: {
        App?: Record<string, unknown>
      }
    }
  }
}

export function isWailsReady(): boolean {
  return !!(window.go?.wails?.App)
}

export function waitForWailsRuntime(timeoutMs = 10000): Promise<void> {
  return new Promise((resolve, reject) => {
    if (isWailsReady()) {
      resolve()
      return
    }

    const startTime = Date.now()
    const checkInterval = 50

    const intervalId = setInterval(() => {
      if (isWailsReady()) {
        clearInterval(intervalId)
        resolve()
        return
      }

      if (Date.now() - startTime >= timeoutMs) {
        clearInterval(intervalId)
        reject(new Error(`Wails runtime not available after ${timeoutMs}ms`))
      }
    }, checkInterval)
  })
}

function delay(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms))
}

export async function withRetry<T>(
  fn: () => Promise<T>,
  maxAttempts = 10,
  initialDelayMs = 100
): Promise<T> {
  let lastError: Error | null = null
  let delayMs = initialDelayMs

  for (let attempt = 1; attempt <= maxAttempts; attempt++) {
    try {
      await waitForWailsRuntime(5000)
      return await fn()
    } catch (err) {
      lastError = err instanceof Error ? err : new Error(String(err))
      if (attempt < maxAttempts) {
        await delay(delayMs)
        delayMs = Math.min(delayMs * 1.5, 2000)
      }
    }
  }

  throw lastError || new Error('Failed after retries')
}
