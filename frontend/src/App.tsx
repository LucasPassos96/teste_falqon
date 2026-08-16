import { Route, Routes } from 'react-router-dom'

import { RequireAuth } from './auth/RequireAuth'
import { AppLayout } from './components/AppLayout'
import FormBuilderPage from './pages/FormBuilderPage'
import FormListPage from './pages/FormListPage'
import LoginPage from './pages/LoginPage'
import PublicFormPage from './pages/PublicFormPage'
import SubmissionsPage from './pages/SubmissionsPage'

/** Envolve a rota no layout administrativo e no guard de sessão. */
function Privada({ children }: { children: React.ReactNode }) {
  return (
    <RequireAuth>
      <AppLayout>{children}</AppLayout>
    </RequireAuth>
  )
}

export default function App() {
  return (
    <Routes>
      {/* Públicas */}
      <Route path="/login" element={<LoginPage />} />
      {/* A rota do visitante fica fora do layout administrativo de propósito:
          ela não tem barra de navegação nem nome de usuário. */}
      <Route path="/f/:slug" element={<PublicFormPage />} />

      {/* Administrativas */}
      <Route
        path="/"
        element={
          <Privada>
            <FormListPage />
          </Privada>
        }
      />
      <Route
        path="/forms/:formId"
        element={
          <Privada>
            <FormBuilderPage />
          </Privada>
        }
      />
      <Route
        path="/forms/:formId/submissions"
        element={
          <Privada>
            <SubmissionsPage />
          </Privada>
        }
      />

      <Route path="*" element={<LoginPage />} />
    </Routes>
  )
}
