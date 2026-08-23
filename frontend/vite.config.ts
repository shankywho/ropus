import { defineConfig } from "@lovable.dev/vite-tanstack-config";

const backendTarget = process.env.BACKEND_INTERNAL_URL || "http://localhost:8080";

export default defineConfig({
  tanstackStart: {
    server: { entry: "server" },
  },
  vite: {
    server: {
      host: "0.0.0.0",
      port: 3000,
      proxy: {
        "/v1": {
          target: backendTarget,
          changeOrigin: true,
        },
      },
    },
  },
});
