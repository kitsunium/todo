import type { ReactNode } from "react";
import { Link, Outlet, useLocation } from "react-router";
import { AuthArt } from "../../components/brand/auth-art";
import { Logo } from "../../components/brand/logo";

/** Split screen: the form on the left, the brand on the right (hidden on small screens). */
export function AuthLayout() {
  const { pathname } = useLocation();
  const onSignup = pathname.startsWith("/signup");
  return (
    <div className="grid min-h-dvh bg-sheet lg:grid-cols-[minmax(0,1fr)_minmax(0,1.08fr)]">
      <div className="flex min-h-dvh flex-col px-5 py-5 sm:px-10 sm:py-7">
        <header className="flex items-center justify-between">
          <Link to="/" className="rounded-md focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-ring" aria-label="Todo home">
            <Logo />
          </Link>
          <p className="text-sm text-fg-3">
            {onSignup ? "Have an account?" : "New to Todo?"}{" "}
            <Link
              to={onSignup ? "/login" : "/signup"}
              className="font-medium text-fg underline-offset-4 hover:underline focus-visible:outline-2 focus-visible:outline-ring"
            >
              {onSignup ? "Sign in" : "Create an account"}
            </Link>
          </p>
        </header>
        <main className="flex flex-1 items-center justify-center py-10">
          <div className="w-full max-w-[400px] animate-rise">
            <Outlet />
          </div>
        </main>
        <footer className="flex items-center justify-between text-xs text-fg-4">
          <span>© {new Date().getFullYear()} Todo</span>
          <span>Built with kit</span>
        </footer>
      </div>
      <aside className="relative m-2 hidden overflow-hidden rounded-[20px] lg:block" aria-hidden="false">
        <BrandPanel />
      </aside>
    </div>
  );
}

function BrandPanel() {
  return (
    <div className="relative flex h-full min-h-[640px] flex-col justify-between overflow-hidden bg-[#f26b1d] p-10 xl:p-14">
      <div
        className="pointer-events-none absolute inset-0"
        style={{
          background:
            "radial-gradient(120% 80% at 0% 0%, #ffc48f 0%, rgb(255 196 143 / 0) 55%)," +
            "radial-gradient(80% 70% at 100% 100%, #d9412f 0%, rgb(217 65 47 / 0) 65%)," +
            "radial-gradient(60% 50% at 85% 10%, #ff9a52 0%, rgb(255 154 82 / 0) 70%)," +
            "linear-gradient(160deg, #ff8a45 0%, #f26b1d 48%, #e4532a 100%)",
        }}
      />
      <svg className="pointer-events-none absolute inset-0 h-full w-full opacity-[0.16] mix-blend-overlay" aria-hidden="true">
        <filter id="grain">
          <feTurbulence type="fractalNoise" baseFrequency="0.9" numOctaves="2" stitchTiles="stitch" />
          <feColorMatrix type="saturate" values="0" />
        </filter>
        <rect width="100%" height="100%" filter="url(#grain)" />
      </svg>
      <div
        className="pointer-events-none absolute inset-0 opacity-[0.12]"
        style={{
          backgroundImage: "radial-gradient(rgb(255 255 255 / 0.9) 1px, transparent 1px)",
          backgroundSize: "22px 22px",
          maskImage: "radial-gradient(70% 60% at 50% 45%, black, transparent)",
        }}
      />
      <div className="pointer-events-none absolute inset-0 hidden bg-[#0e0e10]/25 dark:block" />
      <div className="relative flex items-center gap-2 text-sm font-medium text-white/85">
        <span className="inline-flex size-1.5 rounded-full bg-white" />
        Tasks, shared the calm way
      </div>
      <AuthArt className="relative mx-auto my-4 w-full max-w-[540px]" />
      <div className="relative max-w-[440px]">
        <p className="text-[30px] leading-[36px] font-semibold tracking-[-0.025em] text-balance text-white">
          One calm list for everything you and your people need to get done.
        </p>
        <p className="mt-3 text-[15px] leading-6 text-white/80">
          Priorities, due dates, shared tasks and groups — and nothing you don’t need.
        </p>
      </div>
    </div>
  );
}

export function AuthHeading({ title, children }: { title: string; children?: ReactNode }) {
  return (
    <div className="mb-7">
      <h1 className="text-2xl font-semibold tracking-[-0.025em] text-fg">{title}</h1>
      {children ? <p className="mt-2 text-[15px] leading-[22px] text-fg-3">{children}</p> : null}
    </div>
  );
}
