import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react-swc";
import path from "path";
import { componentTagger } from "lovable-tagger";

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {

  // Load environment variables from the .env file
  // The third parameter '' makes it load all env vars regardless of the VITE_ prefix
  const env = loadEnv(mode, process.cwd(), '');

  // Get the port from the environment variable, with a fallback
  const PORT = env.VITE_PORT ? Number(env.VITE_PORT) : 5173;

  return {
    server: {
      host: "::",
      port: PORT,
      hmr: {
        overlay: false,
      },
    },
    plugins: [react(), mode === "development" && componentTagger()].filter(Boolean),
    resolve: {
      alias: {
        "@": path.resolve(__dirname, "./src"),
      },
    },
    build: {
      terserOptions: {
        compress: {
          drop_console: true
        }
      }
    }
  }
});
