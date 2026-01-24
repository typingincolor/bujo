import '@testing-library/jest-dom'

class ResizeObserverMock {
  observe() {}
  unobserve() {}
  disconnect() {}
}

globalThis.ResizeObserver = ResizeObserverMock

const localStorageMock = (() => {
  let store: Record<string, string> = {}
  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => {
      store[key] = value
    },
    removeItem: (key: string) => {
      delete store[key]
    },
    clear: () => {
      store = {}
    },
  }
})()

Object.defineProperty(window, 'localStorage', {
  value: localStorageMock,
  writable: true,
})

// Mock Wails runtime
const wailsRuntimeMock = {
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
  EventsEmit: () => {},
}

Object.defineProperty(window, 'runtime', {
  value: wailsRuntimeMock,
  writable: true,
})

// Mock Wails go bindings
const wailsAppMock = {
  GetAgenda: () => Promise.resolve({ Days: [], Overdue: [] }),
  GetHabits: () => Promise.resolve({ Habits: [] }),
  GetLists: () => Promise.resolve([]),
  GetGoals: () => Promise.resolve([]),
  GetOutstandingQuestions: () => Promise.resolve([]),
  AddEntry: () => Promise.resolve(),
  AddChildEntry: () => Promise.resolve(),
  MarkEntryDone: () => Promise.resolve(),
  MarkEntryUndone: () => Promise.resolve(),
  EditEntry: () => Promise.resolve(),
  DeleteEntry: () => Promise.resolve(),
  HasChildren: () => Promise.resolve(false),
  MigrateEntry: () => Promise.resolve(),
  MoveEntryToList: () => Promise.resolve(),
  MoveEntryToRoot: () => Promise.resolve(),
  OpenFileDialog: () => Promise.resolve(''),
  CyclePriority: () => Promise.resolve(),
  CancelEntry: () => Promise.resolve(),
  UncancelEntry: () => Promise.resolve(),
  RetypeEntry: () => Promise.resolve(),
}

Object.defineProperty(window, 'go', {
  value: {
    wails: {
      App: wailsAppMock,
    },
  },
  writable: true,
})
