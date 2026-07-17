import { Navigate, Route, Routes } from 'react-router-dom'
import { AppLayout } from './components/layout/AppLayout'
import { ProtectedRoute } from './components/layout/ProtectedRoute'
import { AuthProvider } from './context/AuthContext'
import { PortfolioProvider } from './context/PortfolioContext'
import { Dashboard } from './pages/Dashboard'
import { Forecasts } from './pages/Forecasts'
import { Funds } from './pages/Funds'
import { Login } from './pages/Login'
import { Market } from './pages/Market'
import { Profile } from './pages/Profile'
import { Register } from './pages/Register'
import { Reports } from './pages/Reports'
import { Settings } from './pages/Settings'
import { Trade } from './pages/Trade'

// Balance/health polling only makes sense once a session exists, so
// PortfolioProvider mounts inside the authenticated layout, not around the
// whole app (which would spam 401s on the login/register screens).
function AuthenticatedLayout() {
  return (
    <PortfolioProvider>
      <AppLayout />
    </PortfolioProvider>
  )
}

function App() {
  return (
    <AuthProvider>
      <Routes>
        <Route path="login" element={<Login />} />
        <Route path="register" element={<Register />} />

        <Route element={<ProtectedRoute />}>
          <Route element={<AuthenticatedLayout />}>
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
        </Route>
      </Routes>
    </AuthProvider>
  )
}

export default App
