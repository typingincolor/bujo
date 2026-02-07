import { test, expect, Page } from '@playwright/test'
import * as path from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const SCREENSHOT_DIR = path.resolve(__dirname, '../../docs/images')

function mockWailsBindings() {
  const today = '2026-02-06T00:00:00Z'

  function daysAgoStr(n: number) {
    const d = new Date(today)
    d.setDate(d.getDate() - n)
    return d.toISOString()
  }

  const todayEntries = {
    Date: today,
    Location: 'Home Office',
    Mood: 'Focused',
    Weather: 'Sunny',
    Entries: [
      { ID: 1, EntityID: 'e1', Type: 'Done', Content: 'Review PR for auth module', Priority: 'None', Depth: 0, CreatedAt: today, SortOrder: 1 },
      { ID: 2, EntityID: 'e2', Type: 'Done', Content: 'Write unit tests for payment service', Priority: 'None', Depth: 0, CreatedAt: today, SortOrder: 2 },
      { ID: 3, EntityID: 'e3', Type: 'Task', Content: 'Deploy staging environment', Priority: 'High', Depth: 0, CreatedAt: today, SortOrder: 3 },
      { ID: 4, EntityID: 'e4', Type: 'Task', Content: 'Update API documentation', Priority: 'Medium', Depth: 0, CreatedAt: today, SortOrder: 4 },
      { ID: 5, EntityID: 'e5', Type: 'Task', Content: 'Fix CSS layout bug on mobile', Priority: 'None', Depth: 0, CreatedAt: today, SortOrder: 5 },
      { ID: 6, EntityID: 'e6', Type: 'Note', Content: 'Team standup: discussed migration timeline, agreed on Q2 deadline', Priority: 'None', Depth: 0, CreatedAt: today, SortOrder: 6 },
      { ID: 7, EntityID: 'e7', Type: 'Event', Content: 'Design review @ 2pm', Priority: 'None', Depth: 0, CreatedAt: today, SortOrder: 7 },
      { ID: 8, EntityID: 'e8', Type: 'Event', Content: 'Lunch with Sarah', Priority: 'None', Depth: 0, CreatedAt: today, SortOrder: 8 },
      { ID: 9, EntityID: 'e9', Type: 'Question', Content: 'Should we use Redis or Memcached for session caching?', Priority: 'None', Depth: 0, CreatedAt: today, SortOrder: 9 },
      { ID: 10, EntityID: 'e10', Type: 'Cancelled', Content: 'Afternoon coffee chat', Priority: 'None', Depth: 0, CreatedAt: today, SortOrder: 10 },
      { ID: 11, EntityID: 'e11', Type: 'Task', Content: 'Refactor database connection pooling', Priority: 'Low', Depth: 0, CreatedAt: today, SortOrder: 11 },
      { ID: 12, EntityID: 'e12', Type: 'Note', Content: 'Architecture decision: moving to event-driven messaging for inter-service communication', Priority: 'None', Depth: 0, CreatedAt: today, SortOrder: 12 },
    ],
  }

  const yesterdayEntries = {
    Date: daysAgoStr(1),
    Location: 'Office',
    Mood: 'Productive',
    Weather: 'Cloudy',
    Entries: [
      { ID: 20, EntityID: 'e20', Type: 'Done', Content: 'Set up CI/CD pipeline', Priority: 'None', Depth: 0, CreatedAt: daysAgoStr(1), SortOrder: 1 },
      { ID: 21, EntityID: 'e21', Type: 'Done', Content: 'Code review for frontend team', Priority: 'None', Depth: 0, CreatedAt: daysAgoStr(1), SortOrder: 2 },
      { ID: 22, EntityID: 'e22', Type: 'Done', Content: 'Database migration script', Priority: 'None', Depth: 0, CreatedAt: daysAgoStr(1), SortOrder: 3 },
      { ID: 23, EntityID: 'e23', Type: 'Task', Content: 'Performance profiling on API', Priority: 'Medium', Depth: 0, CreatedAt: daysAgoStr(1), SortOrder: 4 },
      { ID: 24, EntityID: 'e24', Type: 'Note', Content: 'Sprint planning notes: prioritize security backlog', Priority: 'None', Depth: 0, CreatedAt: daysAgoStr(1), SortOrder: 5 },
      { ID: 25, EntityID: 'e25', Type: 'Event', Content: 'Sprint planning meeting', Priority: 'None', Depth: 0, CreatedAt: daysAgoStr(1), SortOrder: 6 },
    ],
  }

  const twoDaysAgoEntries = {
    Date: daysAgoStr(2),
    Location: 'Home Office',
    Mood: 'Calm',
    Weather: 'Rainy',
    Entries: [
      { ID: 30, EntityID: 'e30', Type: 'Done', Content: 'Write integration tests', Priority: 'None', Depth: 0, CreatedAt: daysAgoStr(2), SortOrder: 1 },
      { ID: 31, EntityID: 'e31', Type: 'Task', Content: 'Review security audit findings', Priority: 'High', Depth: 0, CreatedAt: daysAgoStr(2), SortOrder: 2 },
      { ID: 32, EntityID: 'e32', Type: 'Note', Content: 'Learned about Go generics patterns for repository layer', Priority: 'None', Depth: 0, CreatedAt: daysAgoStr(2), SortOrder: 3 },
    ],
  }

  const threeDaysAgoEntries = {
    Date: daysAgoStr(3),
    Location: 'Office',
    Mood: '',
    Weather: 'Partly Cloudy',
    Entries: [
      { ID: 40, EntityID: 'e40', Type: 'Done', Content: 'Fix authentication token refresh', Priority: 'None', Depth: 0, CreatedAt: daysAgoStr(3), SortOrder: 1 },
      { ID: 41, EntityID: 'e41', Type: 'Done', Content: 'Update dependencies', Priority: 'None', Depth: 0, CreatedAt: daysAgoStr(3), SortOrder: 2 },
      { ID: 42, EntityID: 'e42', Type: 'Task', Content: 'Implement rate limiting middleware', Priority: 'High', Depth: 0, CreatedAt: daysAgoStr(3), SortOrder: 3 },
      { ID: 43, EntityID: 'e43', Type: 'Question', Content: 'What monitoring tool should we adopt?', Priority: 'None', Depth: 0, CreatedAt: daysAgoStr(3), SortOrder: 4 },
    ],
  }

  const fourDaysAgoEntries = {
    Date: daysAgoStr(4),
    Location: 'Home Office',
    Mood: 'Energized',
    Weather: 'Sunny',
    Entries: [
      { ID: 50, EntityID: 'e50', Type: 'Done', Content: 'Set up error tracking with Sentry', Priority: 'None', Depth: 0, CreatedAt: daysAgoStr(4), SortOrder: 1 },
      { ID: 51, EntityID: 'e51', Type: 'Task', Content: 'Write documentation for API endpoints', Priority: 'Low', Depth: 0, CreatedAt: daysAgoStr(4), SortOrder: 2 },
      { ID: 52, EntityID: 'e52', Type: 'Note', Content: 'Team retro highlights: improve PR review process', Priority: 'None', Depth: 0, CreatedAt: daysAgoStr(4), SortOrder: 3 },
    ],
  }

  const fiveDaysAgoEntries = {
    Date: daysAgoStr(5),
    Location: 'Office',
    Mood: 'Focused',
    Weather: 'Cloudy',
    Entries: [
      { ID: 60, EntityID: 'e60', Type: 'Done', Content: 'Implement user profile API', Priority: 'None', Depth: 0, CreatedAt: daysAgoStr(5), SortOrder: 1 },
      { ID: 61, EntityID: 'e61', Type: 'Done', Content: 'Database index optimization', Priority: 'None', Depth: 0, CreatedAt: daysAgoStr(5), SortOrder: 2 },
      { ID: 62, EntityID: 'e62', Type: 'Task', Content: 'Set up staging database replicas', Priority: 'Medium', Depth: 0, CreatedAt: daysAgoStr(5), SortOrder: 3 },
      { ID: 63, EntityID: 'e63', Type: 'Event', Content: 'Architecture review meeting', Priority: 'None', Depth: 0, CreatedAt: daysAgoStr(5), SortOrder: 4 },
    ],
  }

  const sixDaysAgoEntries = {
    Date: daysAgoStr(6),
    Location: 'Office',
    Mood: 'Good',
    Weather: 'Sunny',
    Entries: [
      { ID: 70, EntityID: 'e70', Type: 'Done', Content: 'Set up project scaffolding', Priority: 'None', Depth: 0, CreatedAt: daysAgoStr(6), SortOrder: 1 },
      { ID: 71, EntityID: 'e71', Type: 'Done', Content: 'Configure linting and formatting', Priority: 'None', Depth: 0, CreatedAt: daysAgoStr(6), SortOrder: 2 },
      { ID: 72, EntityID: 'e72', Type: 'Note', Content: 'Kickoff meeting: aligned on project goals and milestones', Priority: 'None', Depth: 0, CreatedAt: daysAgoStr(6), SortOrder: 3 },
    ],
  }

  const allDays = [
    todayEntries, yesterdayEntries, twoDaysAgoEntries, threeDaysAgoEntries,
    fourDaysAgoEntries, fiveDaysAgoEntries, sixDaysAgoEntries,
  ]

  const overdueEntries = [
    { ID: 100, EntityID: 'ov1', Type: 'Task', Content: 'Update SSL certificates before expiry', Priority: 'High', Depth: 0, CreatedAt: daysAgoStr(10), SortOrder: 1 },
    { ID: 101, EntityID: 'ov2', Type: 'Task', Content: 'Migrate legacy user accounts', Priority: 'High', Depth: 0, CreatedAt: daysAgoStr(8), SortOrder: 2 },
    { ID: 102, EntityID: 'ov3', Type: 'Task', Content: 'Fix memory leak in WebSocket handler', Priority: 'Medium', Depth: 0, CreatedAt: daysAgoStr(7), SortOrder: 3 },
    { ID: 103, EntityID: 'ov4', Type: 'Task', Content: 'Review third-party license compliance', Priority: 'Medium', Depth: 0, CreatedAt: daysAgoStr(6), SortOrder: 4 },
    { ID: 104, EntityID: 'ov5', Type: 'Task', Content: 'Update onboarding documentation', Priority: 'Low', Depth: 0, CreatedAt: daysAgoStr(5), SortOrder: 5 },
    { ID: 105, EntityID: 'ov6', Type: 'Task', Content: 'Add pagination to user list endpoint', Priority: 'None', Depth: 0, CreatedAt: daysAgoStr(4), SortOrder: 6 },
    { ID: 106, EntityID: 'ov7', Type: 'Task', Content: 'Write changelog for v2.0 release', Priority: 'Low', Depth: 0, CreatedAt: daysAgoStr(3), SortOrder: 7 },
    { ID: 107, EntityID: 'ov8', Type: 'Task', Content: 'Clean up unused feature flags', Priority: 'None', Depth: 0, CreatedAt: daysAgoStr(2), SortOrder: 8 },
  ]

  function dayStatus(dateStr: string, completed: boolean, count: number) {
    return { Date: dateStr, Completed: completed, Count: count }
  }

  const habits = {
    Habits: [
      {
        ID: 1, Name: 'Exercise', GoalPerDay: 1, GoalPerWeek: 0, GoalPerMonth: 0,
        CurrentStreak: 5, CompletionPercent: 85, WeeklyProgress: 6, MonthlyProgress: 22, TodayCount: 1,
        DayHistory: Array.from({ length: 14 }, (_, i) => dayStatus(daysAgoStr(i), i !== 3 && i !== 10, i !== 3 && i !== 10 ? 1 : 0)),
      },
      {
        ID: 2, Name: 'Read 30 minutes', GoalPerDay: 1, GoalPerWeek: 0, GoalPerMonth: 0,
        CurrentStreak: 3, CompletionPercent: 71, WeeklyProgress: 5, MonthlyProgress: 18, TodayCount: 1,
        DayHistory: Array.from({ length: 14 }, (_, i) => dayStatus(daysAgoStr(i), i % 3 !== 2, i % 3 !== 2 ? 1 : 0)),
      },
      {
        ID: 3, Name: 'Meditate', GoalPerDay: 1, GoalPerWeek: 0, GoalPerMonth: 0,
        CurrentStreak: 2, CompletionPercent: 57, WeeklyProgress: 4, MonthlyProgress: 14, TodayCount: 1,
        DayHistory: Array.from({ length: 14 }, (_, i) => dayStatus(daysAgoStr(i), i % 2 === 0, i % 2 === 0 ? 1 : 0)),
      },
      {
        ID: 4, Name: 'Water (8 glasses)', GoalPerDay: 8, GoalPerWeek: 0, GoalPerMonth: 0,
        CurrentStreak: 7, CompletionPercent: 92, WeeklyProgress: 7, MonthlyProgress: 26, TodayCount: 6,
        DayHistory: Array.from({ length: 14 }, (_, i) => {
          const count = 6 + Math.floor(Math.random() * 3)
          return dayStatus(daysAgoStr(i), count >= 8, count)
        }),
      },
      {
        ID: 5, Name: 'Write journal', GoalPerDay: 1, GoalPerWeek: 0, GoalPerMonth: 0,
        CurrentStreak: 4, CompletionPercent: 64, WeeklyProgress: 5, MonthlyProgress: 16, TodayCount: 1,
        DayHistory: Array.from({ length: 14 }, (_, i) => dayStatus(daysAgoStr(i), i < 4 || (i > 5 && i < 10), (i < 4 || (i > 5 && i < 10)) ? 1 : 0)),
      },
    ],
  }

  const lists = [
    {
      ID: 1, Name: 'Project Ideas',
      Items: [
        { RowID: 1, EntityID: 'li1', Version: 1, ValidFrom: today, OpType: 'INSERT', ListEntityID: 'list1', Type: 'item', Content: 'CLI tool for managing dotfiles', CreatedAt: today },
        { RowID: 2, EntityID: 'li2', Version: 1, ValidFrom: today, OpType: 'INSERT', ListEntityID: 'list1', Type: 'done', Content: 'Personal finance tracker', CreatedAt: today },
        { RowID: 3, EntityID: 'li3', Version: 1, ValidFrom: today, OpType: 'INSERT', ListEntityID: 'list1', Type: 'item', Content: 'Recipe management app with meal planning', CreatedAt: today },
        { RowID: 4, EntityID: 'li4', Version: 1, ValidFrom: today, OpType: 'INSERT', ListEntityID: 'list1', Type: 'item', Content: 'Open source contribution dashboard', CreatedAt: today },
      ],
    },
    {
      ID: 2, Name: 'Books to Read',
      Items: [
        { RowID: 5, EntityID: 'li5', Version: 1, ValidFrom: today, OpType: 'INSERT', ListEntityID: 'list2', Type: 'item', Content: 'Designing Data-Intensive Applications', CreatedAt: today },
        { RowID: 6, EntityID: 'li6', Version: 1, ValidFrom: today, OpType: 'INSERT', ListEntityID: 'list2', Type: 'done', Content: 'The Pragmatic Programmer', CreatedAt: today },
        { RowID: 7, EntityID: 'li7', Version: 1, ValidFrom: today, OpType: 'INSERT', ListEntityID: 'list2', Type: 'done', Content: 'Clean Code', CreatedAt: today },
        { RowID: 8, EntityID: 'li8', Version: 1, ValidFrom: today, OpType: 'INSERT', ListEntityID: 'list2', Type: 'item', Content: 'Staff Engineer by Will Larson', CreatedAt: today },
        { RowID: 9, EntityID: 'li9', Version: 1, ValidFrom: today, OpType: 'INSERT', ListEntityID: 'list2', Type: 'item', Content: 'Building Microservices', CreatedAt: today },
      ],
    },
    {
      ID: 3, Name: 'Groceries',
      Items: [
        { RowID: 10, EntityID: 'li10', Version: 1, ValidFrom: today, OpType: 'INSERT', ListEntityID: 'list3', Type: 'done', Content: 'Eggs', CreatedAt: today },
        { RowID: 11, EntityID: 'li11', Version: 1, ValidFrom: today, OpType: 'INSERT', ListEntityID: 'list3', Type: 'done', Content: 'Milk', CreatedAt: today },
        { RowID: 12, EntityID: 'li12', Version: 1, ValidFrom: today, OpType: 'INSERT', ListEntityID: 'list3', Type: 'item', Content: 'Avocados', CreatedAt: today },
        { RowID: 13, EntityID: 'li13', Version: 1, ValidFrom: today, OpType: 'INSERT', ListEntityID: 'list3', Type: 'item', Content: 'Olive oil', CreatedAt: today },
      ],
    },
  ]

  const monthStr = '2026-02-01T00:00:00Z'
  const goals = [
    { ID: 1, EntityID: 'g1', Content: 'Complete API v2 migration', Month: monthStr, Status: 'active', CreatedAt: monthStr },
    { ID: 2, EntityID: 'g2', Content: 'Achieve 80% test coverage', Month: monthStr, Status: 'active', CreatedAt: monthStr },
    { ID: 3, EntityID: 'g3', Content: 'Launch beta to internal users', Month: monthStr, Status: 'done', CreatedAt: monthStr },
    { ID: 4, EntityID: 'g4', Content: 'Read 2 technical books', Month: monthStr, Status: 'active', CreatedAt: monthStr },
    { ID: 5, EntityID: 'g5', Content: 'Set up automated deployment pipeline', Month: monthStr, Status: 'active', CreatedAt: monthStr },
  ]

  const questions = [
    { ID: 9, EntityID: 'e9', Type: 'Question', Content: 'Should we use Redis or Memcached for session caching?', Priority: 'None', Depth: 0, CreatedAt: today, SortOrder: 1 },
    { ID: 43, EntityID: 'e43', Type: 'Question', Content: 'What monitoring tool should we adopt?', Priority: 'None', Depth: 0, CreatedAt: daysAgoStr(3), SortOrder: 2 },
    { ID: 80, EntityID: 'e80', Type: 'Question', Content: 'When should we schedule the team offsite?', Priority: 'None', Depth: 0, CreatedAt: daysAgoStr(5), SortOrder: 3 },
    { ID: 81, EntityID: 'e81', Type: 'Question', Content: 'Should we migrate to GraphQL or stick with REST?', Priority: 'None', Depth: 0, CreatedAt: daysAgoStr(7), SortOrder: 4 },
  ]

  const insightsDashboard = {
    LatestSummary: {
      ID: 1,
      Type: 'weekly',
      WeekStart: daysAgoStr(6),
      WeekEnd: today,
      Content: '## Weekly Summary\\n\\nThis week focused on infrastructure improvements and API development. Key achievements include setting up the CI/CD pipeline, completing the authentication module review, and making significant progress on the staging deployment.\\n\\n### Highlights\\n- Completed 15 tasks across 6 working days\\n- Fixed critical authentication token refresh bug\\n- Set up error tracking with Sentry\\n- Made architecture decision on event-driven messaging',
      CreatedAt: today,
    },
    ActiveInitiatives: [
      { ID: 1, Name: 'API v2 Migration', Status: 'in_progress', Description: 'Migrating all endpoints to v2 with improved auth' },
      { ID: 2, Name: 'Infrastructure Modernization', Status: 'in_progress', Description: 'Moving to containerized deployments' },
    ],
    HighPriorityActions: [
      { ID: 1, Content: 'Deploy staging environment', Source: 'task', Priority: 'High' },
      { ID: 2, Content: 'Update SSL certificates before expiry', Source: 'overdue', Priority: 'High' },
    ],
    RecentDecisions: [
      { ID: 1, Content: 'Moving to event-driven messaging for inter-service communication', Date: today },
      { ID: 2, Content: 'Adopted Sentry for error tracking', Date: daysAgoStr(4) },
    ],
    DaysSinceLastSummary: 0,
    Status: 'active',
  }

  const journalDocument = [
    'x Review PR for auth module',
    'x Write unit tests for payment service',
    '.!!! Deploy staging environment',
    '.!! Update API documentation',
    '. Fix CSS layout bug on mobile',
    '- Team standup: discussed migration timeline, agreed on Q2 deadline',
    'o Design review @ 2pm',
    'o Lunch with Sarah',
    '? Should we use Redis or Memcached for session caching?',
    '~ Afternoon coffee chat',
    '.! Refactor database connection pooling',
    '- Architecture decision: moving to event-driven messaging for inter-service communication',
  ].join('\n')

  return `
    window.go = {
      wails: {
        App: {
          // Data loading
          GetDayEntries: () => Promise.resolve(${JSON.stringify(allDays)}),
          GetOverdue: () => Promise.resolve(${JSON.stringify(overdueEntries)}),
          GetHabits: () => Promise.resolve(${JSON.stringify(habits)}),
          GetLists: () => Promise.resolve(${JSON.stringify(lists)}),
          GetGoals: () => Promise.resolve(${JSON.stringify(goals)}),
          GetOutstandingQuestions: () => Promise.resolve(${JSON.stringify(questions)}),

          // Journal/Editor
          GetEditableDocumentWithEntries: () => Promise.resolve({
            document: ${JSON.stringify(journalDocument)},
            entries: [],
          }),
          GetEditableDocument: () => Promise.resolve(${JSON.stringify(journalDocument)}),
          ValidateEditableDocument: () => Promise.resolve([]),
          ApplyEditableDocument: () => Promise.resolve({ success: true }),

          // Entry operations (no-ops for screenshots)
          AddEntry: () => Promise.resolve([1]),
          AddChildEntry: () => Promise.resolve(),
          MarkEntryDone: () => Promise.resolve(),
          MarkEntryUndone: () => Promise.resolve(),
          EditEntry: () => Promise.resolve(),
          DeleteEntry: () => Promise.resolve(),
          HasChildren: () => Promise.resolve(false),
          MigrateEntry: () => Promise.resolve(100),
          MoveEntryToList: () => Promise.resolve(),
          MoveEntryToRoot: () => Promise.resolve(),
          CyclePriority: () => Promise.resolve(),
          CancelEntry: () => Promise.resolve(),
          UncancelEntry: () => Promise.resolve(),
          RetypeEntry: () => Promise.resolve(),
          GetEntryContext: () => Promise.resolve([]),
          GetEntry: () => Promise.resolve(null),

          // Search
          Search: () => Promise.resolve([]),

          // Header context
          SetMood: () => Promise.resolve(),
          SetWeather: () => Promise.resolve(),
          SetLocation: () => Promise.resolve(),
          GetLocationHistory: () => Promise.resolve(['Home Office', 'Office', 'Coffee Shop']),

          // Habits
          CreateHabit: () => Promise.resolve(1),
          LogHabit: () => Promise.resolve(),
          DeleteHabit: () => Promise.resolve(),

          // Lists
          CreateList: () => Promise.resolve(1),
          AddListItem: () => Promise.resolve(),
          ToggleListItem: () => Promise.resolve(),
          DeleteList: () => Promise.resolve(),

          // Goals
          AddGoal: () => Promise.resolve(1),
          UpdateGoalStatus: () => Promise.resolve(),
          DeleteGoal: () => Promise.resolve(),

          // Insights
          GetInsightsDashboard: () => Promise.resolve(${JSON.stringify(insightsDashboard)}),
          IsInsightsAvailable: () => Promise.resolve(true),
          GetInsightsSummaryForWeek: () => Promise.resolve(${JSON.stringify(insightsDashboard.LatestSummary)}),
          GetInsightsActionsForWeek: () => Promise.resolve([
            { ID: 1, Content: 'Deploy staging environment', Source: 'task', Priority: 'High', Status: 'open' },
            { ID: 2, Content: 'Review security audit findings', Source: 'task', Priority: 'High', Status: 'open' },
            { ID: 3, Content: 'Update API documentation', Source: 'task', Priority: 'Medium', Status: 'open' },
          ]),
          GetInsightsInitiativePortfolio: () => Promise.resolve([
            { ID: 1, Name: 'API v2 Migration', Status: 'in_progress', Description: 'Migrating all endpoints to v2', Progress: 65 },
            { ID: 2, Name: 'Infrastructure Modernization', Status: 'in_progress', Description: 'Containerized deployments', Progress: 40 },
            { ID: 3, Name: 'Security Hardening', Status: 'planned', Description: 'Comprehensive security review', Progress: 10 },
          ]),
          GetInsightsInitiativeDetail: () => Promise.resolve({
            ID: 1, Name: 'API v2 Migration', Status: 'in_progress',
            Description: 'Migrating all endpoints to v2 with improved authentication',
            Progress: 65, RelatedEntries: [],
          }),
          GetInsightsDistinctTopics: () => Promise.resolve(['infrastructure', 'api', 'testing', 'security', 'documentation']),
          GetInsightsTopicTimeline: () => Promise.resolve([]),
          GetInsightsDecisionLog: () => Promise.resolve([
            { ID: 1, Content: 'Moving to event-driven messaging', Context: 'Inter-service communication', Date: '${today}' },
            { ID: 2, Content: 'Adopted Sentry for error tracking', Context: 'Observability', Date: '${daysAgoStr(4)}' },
            { ID: 3, Content: 'Use JWT with refresh tokens', Context: 'Authentication', Date: '${daysAgoStr(8)}' },
          ]),
          GetInsightsWeeklyReport: () => Promise.resolve({ Content: 'Weekly report content' }),

          // Attention scores
          GetAttentionScores: () => Promise.resolve([]),

          // Settings
          GetVersion: () => Promise.resolve('2.0.0'),
          OpenFileDialog: () => Promise.resolve(''),
          ReadFile: () => Promise.resolve(''),

          // Backup
          CreateBackup: () => Promise.resolve('/path/to/backup'),
          RestoreBackup: () => Promise.resolve(),

          // Answer question
          AnswerQuestion: () => Promise.resolve(),
        },
      },
    };

    // Mock Wails runtime
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

async function waitForApp(page: Page) {
  await page.waitForLoadState('networkidle')
  // Wait for the sidebar to be visible (indicates app has loaded)
  await expect(page.getByRole('button', { name: 'Journal' })).toBeVisible({ timeout: 15000 })
}

async function navigateTo(page: Page, label: string) {
  const button = page.locator(`button:has-text("${label}")`)
  await button.click()
  await page.waitForTimeout(500)
}

async function screenshot(page: Page, name: string) {
  await page.screenshot({
    path: path.join(SCREENSHOT_DIR, `${name}.png`),
    fullPage: false,
  })
}

test.describe('Documentation screenshots', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(mockWailsBindings())
    await page.setViewportSize({ width: 1280, height: 800 })
    await page.goto('/')
    await waitForApp(page)
  })

  test('Journal view', async ({ page }) => {
    // Journal is the default view
    await expect(page.locator('.cm-content')).toBeVisible({ timeout: 10000 })
    await page.waitForTimeout(1000)
    await screenshot(page, 'journal-view')
  })

  test('Pending Tasks view', async ({ page }) => {
    await navigateTo(page, 'Pending Tasks')
    await expect(page.locator('text=Update SSL certificates')).toBeVisible({ timeout: 5000 })
    await screenshot(page, 'pending-tasks-view')
  })

  test('Weekly Review view', async ({ page }) => {
    await navigateTo(page, 'Weekly Review')
    await expect(page.locator('text=Week of')).toBeVisible({ timeout: 5000 })
    await page.waitForTimeout(500)
    await screenshot(page, 'weekly-review-view')
  })

  test('Open Questions view', async ({ page }) => {
    await navigateTo(page, 'Open Questions')
    await expect(page.locator('text=Redis or Memcached')).toBeVisible({ timeout: 5000 })
    await screenshot(page, 'open-questions-view')
  })

  test('Habit Tracker view', async ({ page }) => {
    await navigateTo(page, 'Habit Tracker')
    await expect(page.locator('text=Exercise')).toBeVisible({ timeout: 5000 })
    await screenshot(page, 'habit-tracker-view')
  })

  test('Lists view', async ({ page }) => {
    await navigateTo(page, 'Lists')
    await expect(page.locator('text=Project Ideas')).toBeVisible({ timeout: 5000 })
    await screenshot(page, 'lists-view')
  })

  test('Monthly Goals view', async ({ page }) => {
    await navigateTo(page, 'Monthly Goals')
    await expect(page.locator('text=Complete API v2 migration')).toBeVisible({ timeout: 5000 })
    await screenshot(page, 'monthly-goals-view')
  })

  test('Search view', async ({ page }) => {
    await navigateTo(page, 'Search')
    await expect(page.getByPlaceholder('Search entries...')).toBeVisible({ timeout: 5000 })
    await screenshot(page, 'search-view')
  })

  test('Statistics view', async ({ page }) => {
    await navigateTo(page, 'Statistics')
    await page.waitForTimeout(1000)
    await screenshot(page, 'statistics-view')
  })

  test('Insights view', async ({ page }) => {
    await navigateTo(page, 'Insights')
    await page.waitForTimeout(1500)
    await screenshot(page, 'insights-view')
  })

  test('Settings view', async ({ page }) => {
    await navigateTo(page, 'Settings')
    await expect(page.locator('text=Appearance')).toBeVisible({ timeout: 5000 })
    await screenshot(page, 'settings-view')
  })
})
