import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import 'highlight.js/styles/github.css'
import './index.css'
import App from './App.tsx'

// Apply persisted theme before first render to avoid flash.
const stored = localStorage.getItem('snippr-theme')
if (stored) {
  try {
    const { state } = JSON.parse(stored)
    if (state?.theme === 'dark') document.documentElement.classList.add('dark')
  } catch { /* ignore */ }
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
