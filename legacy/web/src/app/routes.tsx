import { lazy, Suspense } from "react";
import { Navigate, Route, Routes } from "react-router";
import { useMe } from "../api/queries";
import { NotFoundPage } from "../features/NotFoundPage";
import { Splash, useSessionGuard } from "../features/shell/session";

// Two chunks load on demand: the signed-out pages, and the app itself.
const auth = () => import("../features/auth/pages");
const AuthLayout = lazy(() => auth().then((m) => ({ default: m.AuthLayout })));
const LoginPage = lazy(() => auth().then((m) => ({ default: m.LoginPage })));
const SignupPage = lazy(() => auth().then((m) => ({ default: m.SignupPage })));
const CheckInboxPage = lazy(() => auth().then((m) => ({ default: m.CheckInboxPage })));
const VerifyPage = lazy(() => auth().then((m) => ({ default: m.VerifyPage })));
const ForgotPage = lazy(() => auth().then((m) => ({ default: m.ForgotPage })));
const ResetPage = lazy(() => auth().then((m) => ({ default: m.ResetPage })));
const AppGate = lazy(() => import("../features/shell/Shell").then((m) => ({ default: m.AppGate })));

/** "/": the app when signed in, the sign-in page otherwise. */
function Root() {
  const me = useMe();
  if (me.isPending) return <Splash />;
  return <Navigate to={me.isSuccess ? "/app/today" : "/login"} replace />;
}

export function AppRoutes() {
  useSessionGuard();
  return (
    <Suspense fallback={<Splash />}>
      <Routes>
        <Route path="/" element={<Root />} />
        <Route element={<AuthLayout />}>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/signup" element={<SignupPage />} />
          <Route path="/check-inbox" element={<CheckInboxPage />} />
          <Route path="/verify" element={<VerifyPage />} />
          <Route path="/forgot" element={<ForgotPage />} />
          <Route path="/reset" element={<ResetPage />} />
        </Route>
        <Route path="/app/*" element={<AppGate />} />
        <Route path="*" element={<NotFoundPage />} />
      </Routes>
    </Suspense>
  );
}
