/// <reference types="vitest/config" />
// The todo web app. `npm run build` writes ./dist, which web.go embeds
// (//go:embed dist) and the Go binary serves at "/".
//
//   npm run dev        http://localhost:5173, /api and /_kit proxied to the app (KIT_APP)
//   npm run dev:mock   the whole HTTP contract in memory: no backend needed

import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

/** The running kit app `npm run dev` proxies to. */
const KIT_APP = process.env.KIT_APP ?? "http://localhost:4000";

export default defineConfig(({ mode }) => ({
  base: "/",
  plugins: [react(), tailwindcss()],
  build: {
    outDir: "dist",
    emptyOutDir: true,
    target: "es2022",
    sourcemap: false,
    // The frontend CSP has no data: in font-src: never inline an asset.
    assetsInlineLimit: 0,
    rolldownOptions: {
      output: {
        // //go:embed drops files whose name starts with "_" or ".": never emit one.
        sanitizeFileName: (name: string) =>
          name.replace(/[\0?*]/g, "").replace(/(^|\/)[_.]+/g, "$1"),
      },
    },
  },
  server: {
    port: 5173,
    strictPort: false,
    proxy:
      mode === "mock"
        ? undefined
        : {
            "/api": { target: KIT_APP, changeOrigin: true },
            "/_kit": { target: KIT_APP, changeOrigin: true },
          },
  },
  test: {
    include: ["tests/**/*.test.ts"],
    environment: "node",
  },
}));
