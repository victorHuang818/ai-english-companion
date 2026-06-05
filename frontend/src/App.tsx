import React from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import { AuthPage } from './pages/AuthPage'
import { DashboardPage } from './pages/DashboardPage'
import { InterviewRoomPage } from './pages/InterviewRoomPage'
import { ReportPage } from './pages/ReportPage'
import { useAuthStore } from './store/authStore'

const ProtectedRoute = ({ children }: { children: React.ReactNode }) => {
  const { token } = useAuthStore()
  if (!token) {
    return <Navigate to="/auth" replace />
  }
  return <>{children}</>
}

function App() {
  return (
    <Routes>
      <Route path="/auth" element={<AuthPage />} />
      <Route 
        path="/" 
        element={
          <ProtectedRoute>
            <DashboardPage />
          </ProtectedRoute>
        } 
      />
      <Route 
        path="/interview/:sessionId" 
        element={
          <ProtectedRoute>
            <InterviewRoomPage />
          </ProtectedRoute>
        } 
      />
      <Route 
        path="/report/:sessionId" 
        element={
          <ProtectedRoute>
            <ReportPage />
          </ProtectedRoute>
        } 
      />
    </Routes>
  )
}


export default App
