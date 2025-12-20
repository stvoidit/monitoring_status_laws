import { type ConfigEnv, loadEnv } from "vite";
import { defineConfig } from "vite";
import { URL, fileURLToPath } from "node:url";
import vue from "@vitejs/plugin-vue";
import Components from "unplugin-vue-components/vite";
import { ElementPlusResolver } from "unplugin-vue-components/resolvers";
import ElementPlus from "unplugin-element-plus/vite";
// import vueDevTools from "vite-plugin-vue-devtools";
import Unfonts from "unplugin-fonts/vite";

/** @see https://vitejs.dev/config/ */
export default ({ mode }: ConfigEnv) => {
    const env = loadEnv(mode, ".");
    const isDev = mode === "development";
    return defineConfig({
        plugins: [
            vue({
                features: {
                    optionsAPI: false,
                    customElement: false,
                    propsDestructure: true,
                },
            }),
            Unfonts({
                custom: {
                    prefetch: true,
                    families: [
                        {
                            name: "TT Moscow Economy",
                            src: "src/assets/fonts/*.ttf",
                        },
                    ],
                },
            }),
            // vueDevTools(),
            Components({
                sourcemap: isDev,
                dts: false,
                resolvers: [
                    ElementPlusResolver({
                        importStyle: "css",
                    }),
                ],
            }),
            ElementPlus({
                defaultLocale: "ru",
                format: "esm",
            }),
        ],
        server: {
            proxy: {
                "/api": {
                    target: env.VITE_PROXY_TARGET || "http://localhost:8080",
                    changeOrigin: true,
                },
            },
            host: env.VITE_HOST || "0.0.0.0",
            port: parseInt(env.VITE_PORT),
        },
        css: {
            transformer: "lightningcss",
            lightningcss: {
                nonStandard: {
                    deepSelectorCombinator: true,
                },
            },
        },
        build: {
            cssMinify: "lightningcss",
            target: "esnext",
            outDir: "dist",
            manifest: false,
            minify: "esbuild",
            emptyOutDir: true,
            sourcemap: isDev,
            cssCodeSplit: true,
            chunkSizeWarningLimit: 2 << 19,
        },
        resolve: {
            alias:
            {
                "@": fileURLToPath(new URL("./src", import.meta.url)),
            },
        },
    });
};
