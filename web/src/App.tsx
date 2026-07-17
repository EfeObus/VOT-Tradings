import { Navigate, Route, Routes } from 'react-router-dom'
import { AppLayout } from './components/layout/AppLayout'
import { PortfolioProvider } from './context/PortfolioContext'
import { Dashboard } from './pages/Dashboard'
import { Intelligence } from './pages/Intelligence'
import { Market } from './pages/Market'
import { Settings } from './pages/Settings'
import { Trade } from './pages/Trade'

function App() {
  return (
    <PortfolioProvider>
      <Routes>
        <Route element={<AppLayout />}>
          <Route index element={<Navigate to="/dashboard" replace />} />
          <Route path="dashboard" element={<Dashboard />} />
          <Route path="market/:symbol" element={<Market />} />
          <Route path="intelligence" element={<Intelligence />} />
          <Route path="trade" element={<Trade />} />
          <Route path="settings" element={<Settings />} />
        </Route>
      </Routes>
    </PortfolioProvider>
  )
}

export default App
