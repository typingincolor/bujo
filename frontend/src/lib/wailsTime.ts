import { time } from '../wailsjs/go/models'

export function toWailsTime(date: Date): time.Time {
  const pad = (n: number) => n.toString().padStart(2, '0')
  const year = date.getFullYear()
  const month = pad(date.getMonth() + 1)
  const day = pad(date.getDate())
  const hours = pad(date.getHours())
  const minutes = pad(date.getMinutes())
  const seconds = pad(date.getSeconds())
  const offset = -date.getTimezoneOffset()
  const sign = offset >= 0 ? '+' : '-'
  const offsetHours = pad(Math.floor(Math.abs(offset) / 60))
  const offsetMins = pad(Math.abs(offset) % 60)
  return `${year}-${month}-${day}T${hours}:${minutes}:${seconds}${sign}${offsetHours}:${offsetMins}` as unknown as time.Time
}
