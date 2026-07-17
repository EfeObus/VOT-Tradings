import { Navigate, Route, Routes } from 'react-router-dom'
import { AppLayout } from './components/layout/AppLayout'
import { PortfolioProvider } from './context/PortfolioContext'
import { Dashboard } from './pages/Dashboard'
import { Forecasts } from './pages/Forecasts'
import { Funds } from './pages/Funds'
import { Market } from './pages/Market'
import { Profile } from './pages/Profile'
import { Reports } from './pages/Reports'
import { Settings } from './pages/Settings'
import { Trade } from './pages/Trade'

function App() {
  return (
    <PortfolioProvider>
      <Routes>
        <Route element={<AppLayout />}>
          <Route index element={<Navigate to="/dashboard" replace />} />
          <Route path="profile" element={<Profile />} />
          <Route path="dashboard" element={<Dashboard />} />
          <Route path="funds" element={<Funds />} />
          <Route path="market/:symbol" element={<Market />} />
          <Route path="trade" element={<Trade />} />
          <Route path="forecasts" element={<Forecasts />} />
          <Route path="reports" element={<Reports />} />
          <Route path="settings" element={<Settings />} />
        </Route>
      </Routes>
    </PortfolioProvider>
  )
}

export default App
