import { useEffect, useState } from 'react'
import { GetAgenda, GetHabits, GetLists, GetGoals } from './wailsjs/go/wails/App'
import './App.css'

type View = 'today' | 'habits' | 'lists' | 'goals'

function App() {
  const [view, setView] = useState<View>('today')
  const [agenda, setAgenda] = useState<any>(null)
  const [habits, setHabits] = useState<any>(null)
  const [lists, setLists] = useState<any>(null)
  const [goals, setGoals] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    setLoading(true)
    setError(null)
    try {
      const now = new Date()
      const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())
      const weekLater = new Date(today.getTime() + 7 * 24 * 60 * 60 * 1000)
      const monthStart = new Date(now.getFullYear(), now.getMonth(), 1)

      const [agendaData, habitsData, listsData, goalsData] = await Promise.all([
        GetAgenda(today.toISOString(), weekLater.toISOString()),
        GetHabits(7),
        GetLists(),
        GetGoals(monthStart.toISOString()),
      ])

      setAgenda(agendaData)
      setHabits(habitsData)
      setLists(listsData)
      setGoals(goalsData)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load data')
    } finally {
      setLoading(false)
    }
  }

  const getEntrySymbol = (type: string) => {
    switch (type) {
      case 'task': return '•'
      case 'done': return '×'
      case 'event': return '○'
      case 'note': return '–'
      case 'migrated': return '>'
      default: return '•'
    }
  }

  if (loading) {
    return <div className="container"><h1>Loading...</h1></div>
  }

  if (error) {
    return (
      <div className="container">
        <h1>Error</h1>
        <p className="error">{error}</p>
        <button onClick={loadData}>Retry</button>
      </div>
    )
  }

  return (
    <div className="app">
      <nav className="sidebar">
        <h2>Bujo</h2>
        <ul>
          <li className={view === 'today' ? 'active' : ''} onClick={() => setView('today')}>Today</li>
          <li className={view === 'habits' ? 'active' : ''} onClick={() => setView('habits')}>Habits</li>
          <li className={view === 'lists' ? 'active' : ''} onClick={() => setView('lists')}>Lists</li>
          <li className={view === 'goals' ? 'active' : ''} onClick={() => setView('goals')}>Goals</li>
        </ul>
        <button className="refresh" onClick={loadData}>↻ Refresh</button>
      </nav>

      <main className="content">
        {view === 'today' && (
          <div className="view">
            <h1>Today</h1>
            {agenda?.Overdue?.length > 0 && (
              <section>
                <h3>Overdue</h3>
                <ul className="entries">
                  {agenda.Overdue.map((entry: any, i: number) => (
                    <li key={i} className={`entry ${entry.Type}`}>
                      <span className="symbol">{getEntrySymbol(entry.Type)}</span>
                      <span className="content">{entry.Content}</span>
                    </li>
                  ))}
                </ul>
              </section>
            )}
            {agenda?.Days?.map((day: any, i: number) => (
              <section key={i}>
                <h3>{new Date(day.Date).toLocaleDateString('en-US', { weekday: 'long', month: 'short', day: 'numeric' })}</h3>
                {day.Entries?.length > 0 ? (
                  <ul className="entries">
                    {day.Entries.map((entry: any, j: number) => (
                      <li key={j} className={`entry ${entry.Type}`}>
                        <span className="symbol">{getEntrySymbol(entry.Type)}</span>
                        <span className="content">{entry.Content}</span>
                      </li>
                    ))}
                  </ul>
                ) : (
                  <p className="empty">No entries</p>
                )}
              </section>
            ))}
          </div>
        )}

        {view === 'habits' && (
          <div className="view">
            <h1>Habits</h1>
            {habits?.Habits?.length > 0 ? (
              <ul className="habits">
                {habits.Habits.map((habit: any, i: number) => (
                  <li key={i} className="habit">
                    <div className="habit-header">
                      <span className="name">{habit.Name}</span>
                      <span className="streak">🔥 {habit.CurrentStreak}</span>
                    </div>
                    <div className="habit-history">
                      {habit.DayHistory?.slice(0, 7).reverse().map((day: any, j: number) => (
                        <span key={j} className={`day ${day.Completed ? 'done' : ''}`}>
                          {day.Completed ? '●' : '○'}
                        </span>
                      ))}
                    </div>
                  </li>
                ))}
              </ul>
            ) : (
              <p className="empty">No habits tracked yet</p>
            )}
          </div>
        )}

        {view === 'lists' && (
          <div className="view">
            <h1>Lists</h1>
            {lists?.length > 0 ? (
              lists.map((list: any, i: number) => (
                <section key={i}>
                  <h3>{list.Name}</h3>
                  {list.Items?.length > 0 ? (
                    <ul className="entries">
                      {list.Items.map((item: any, j: number) => (
                        <li key={j} className={`entry ${item.Type}`}>
                          <span className="symbol">{item.Type === 'done' ? '×' : '•'}</span>
                          <span className="content">{item.Content}</span>
                        </li>
                      ))}
                    </ul>
                  ) : (
                    <p className="empty">Empty list</p>
                  )}
                </section>
              ))
            ) : (
              <p className="empty">No lists yet</p>
            )}
          </div>
        )}

        {view === 'goals' && (
          <div className="view">
            <h1>Goals - {new Date().toLocaleDateString('en-US', { month: 'long', year: 'numeric' })}</h1>
            {goals?.length > 0 ? (
              <ul className="goals">
                {goals.map((goal: any, i: number) => (
                  <li key={i} className={`goal ${goal.Status}`}>
                    <span className="symbol">{goal.Status === 'done' ? '✓' : '○'}</span>
                    <span className="content">{goal.Content}</span>
                  </li>
                ))}
              </ul>
            ) : (
              <p className="empty">No goals for this month</p>
            )}
          </div>
        )}
      </main>
    </div>
  )
}

export default App
