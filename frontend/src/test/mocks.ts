import { service, domain } from '@/wailsjs/go/models'

export function createMockEntry(overrides: Partial<{
  ID: number
  EntityID: string
  Type: string
  Content: string
  Priority: string
  ParentID: number | null
  Depth: number
  CreatedAt: string
}> = {}): domain.Entry {
  return domain.Entry.createFrom({
    ID: overrides.ID ?? 1,
    EntityID: overrides.EntityID ?? 'e1',
    Type: overrides.Type ?? 'Task',
    Content: overrides.Content ?? 'Test entry',
    Priority: overrides.Priority ?? '',
    ParentID: overrides.ParentID ?? null,
    Depth: overrides.Depth ?? 0,
    CreatedAt: overrides.CreatedAt ?? '2026-01-17T10:00:00Z',
  })
}

export function createMockDayEntries(overrides: Partial<{
  Date: string
  Location: string
  Mood: string
  Weather: string
  Entries: domain.Entry[]
}> = {}): service.DayEntries {
  return service.DayEntries.createFrom({
    Date: overrides.Date ?? '2026-01-17T00:00:00Z',
    Location: overrides.Location ?? '',
    Mood: overrides.Mood ?? '',
    Weather: overrides.Weather ?? '',
    Entries: overrides.Entries ?? [],
  })
}

export function createMockAgenda(overrides: Partial<{
  Overdue: domain.Entry[]
  Days: service.DayEntries[]
}> = {}): service.MultiDayAgenda {
  return service.MultiDayAgenda.createFrom({
    Overdue: overrides.Overdue ?? [],
    Days: overrides.Days ?? [createMockDayEntries()],
  })
}
