import { defineConfig } from "vite";
import react from "@vitejs/plugin-react-swc";
import path from "path";

export default defineConfig(({ mode }) => ({
  // App na raiz do domínio em produção
  base: "/",

  plugins: [react()], // 🔥 tiramos componentTagger()

  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },

  build: {
    outDir: "dist",
    assetsDir: "assets",
    sourcemap: false,
    target: "es2018",
    minify: "esbuild",
  },

  server: {
    host: true,
    port: 8081,
    strictPort: true,
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
        rewrite: (p) => p.replace(/^\/api/, ""),
      },
    },
  },
}));