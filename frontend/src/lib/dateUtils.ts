export function formatDateForInput(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

export function parseDateFromInput(value: string): Date | null {
  if (!value) return null
  const [year, month, day] = value.split('-').map(Number)
  if (isNaN(year) || isNaN(month) || isNaN(day)) return null
  const date = new Date(year, month - 1, day, 0, 0, 0)
  if (isNaN(date.getTime())) return null
  return date
}
