import { Link } from "react-router";
import { LostIllo } from "../components/brand/illustrations";
import { Logo } from "../components/brand/logo";
import { Button } from "../components/ui/button";

export function NotFoundPage({ inApp }: { inApp?: boolean }) {
  const body = (
    <div className="flex flex-col items-center px-6 text-center animate-rise">
      <LostIllo />
      <p className="mt-3 text-sm font-medium tracking-wide text-fg-4">404</p>
      <h1 className="mt-1 text-xl font-semibold tracking-[-0.02em] text-fg">This page wandered off</h1>
      <p className="mt-1.5 max-w-[320px] text-sm text-fg-3">The link may be old, or the page may have moved. Your tasks are safe.</p>
      <Button asChild variant="primary" className="mt-6">
        <Link to={inApp ? "/app/today" : "/"}>{inApp ? "Back to Today" : "Go home"}</Link>
      </Button>
    </div>
  );
  if (inApp) return <div className="flex flex-1 items-center justify-center py-20">{body}</div>;
  return (
    <div className="flex min-h-dvh flex-col bg-sheet px-6 py-6">
      <Link to="/" aria-label="Todo home">
        <Logo />
      </Link>
      <div className="flex flex-1 items-center justify-center">{body}</div>
    </div>
  );
}
