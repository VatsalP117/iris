import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import './index.css'
import App, { Home } from './App.tsx'
import Docs from './Docs.tsx'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      <Routes>
        <Route element={<App />}>
          <Route index element={<Home />} />
          <Route path="docs" element={<Docs />} />
        </Route>
      </Routes>
    </BrowserRouter>
  </StrictMode>,
)
