import { useState } from 'react'
import { Greet } from './wailsjs/go/wails/App'
import './App.css'

function App() {
  const [greeting, setGreeting] = useState('')
  const [name, setName] = useState('')

  const greet = async () => {
    const result = await Greet(name)
    setGreeting(result)
  }

  return (
    <div className="container">
      <h1>Bujo Desktop</h1>
      <div className="input-box">
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Enter your name"
        />
        <button onClick={greet}>Greet</button>
      </div>
      {greeting && <p className="greeting">{greeting}</p>}
    </div>
  )
}

export default App
