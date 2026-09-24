import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { App } from "./app/App";
import { applyTheme } from "./lib/theme";
import "./styles/index.css";

applyTheme();

async function boot() {
  // Mock mode (vite --mode mock) answers the API in memory. In a production
  // build MODE is "production": the branch and its import are dropped.
  if (import.meta.env.MODE === "mock") {
    const { installMock } = await import("./mock/install");
    installMock();
  }
  const root = document.getElementById("root");
  if (!root) throw new Error("no #root element");
  createRoot(root).render(
    <StrictMode>
      <App />
    </StrictMode>,
  );
}

void boot();
