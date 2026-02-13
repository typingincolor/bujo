import { LogDebug, LogInfo, LogWarning, LogError } from '@/wailsjs/runtime/runtime'

export const log = {
  debug: (message: string, ...args: unknown[]) => {
    const formatted = args.length > 0 ? `${message} ${JSON.stringify(args)}` : message
    LogDebug(formatted)
  },
  info: (message: string, ...args: unknown[]) => {
    const formatted = args.length > 0 ? `${message} ${JSON.stringify(args)}` : message
    LogInfo(formatted)
  },
  warn: (message: string, ...args: unknown[]) => {
    const formatted = args.length > 0 ? `${message} ${JSON.stringify(args)}` : message
    LogWarning(formatted)
  },
  error: (message: string, ...args: unknown[]) => {
    const formatted = args.length > 0 ? `${message} ${JSON.stringify(args)}` : message
    LogError(formatted)
  },
}
