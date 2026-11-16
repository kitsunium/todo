/// <reference types="vitest/config" />
// The todo web app. `npm run build` writes ./dist, which web.go embeds
// (//go:embed dist) and the Go binary serves at "/".
//
//   npm run dev        http://localhost:5173, /api and /_app proxied to the app (APP_URL)
//   npm run dev:mock   the whole HTTP contract in memory: no backend needed

import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

/** The running app `npm run dev` proxies to. */
const APP_URL = process.env.APP_URL ?? "http://localhost:4000";

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
            "/api": { target: APP_URL, changeOrigin: true },
            "/_app": { target: APP_URL, changeOrigin: true },
          },
  },
  test: {
    include: ["tests/**/*.test.ts"],
    environment: "node",
  },
}));
