import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import path from 'path'
import fs from 'fs'
import yaml from 'js-yaml'

const rulesYamlPath = '../game/rules.yaml'
const rulesPlugin = {
    name: 'rules-plugin',
    resolveId(id) {
        if (id === 'virtual:rules') {
            return '\0virtual:rules'
        }
        return null
    },
    load(id) {
        if (id === '\0virtual:rules') {
            try {
                const yamlContent = fs.readFileSync(path.resolve(__dirname, rulesYamlPath), 'utf-8')
                const data = yaml.load(yamlContent)
                return `export default ${JSON.stringify(data)}`
            } catch (error) {
                console.error('Error loading rules.yaml:', error)
                return 'export default {}'
            }
        }
        return null
    }
}

export default defineConfig(({ command }) => {
    const GAME_API_URL = process.env.API_URL ?? "http://localhost:8086"

    return {
        plugins: [
            svelte(),
            rulesPlugin
        ],
        define: {
            'GAME_API_URL': JSON.stringify(GAME_API_URL)
        },
        base: './',
        build: {
            outDir: 'dist',
            assetsDir: 'assets',
            rollupOptions: {
                input: './index.html',
                output: {
                    manualChunks: undefined,
                    entryFileNames: 'bundle.js',
                    assetFileNames: 'bundle.[ext]'
                }
            }
        }
    }
})