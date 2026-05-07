import { app, BrowserWindow } from 'electron'
// @ts-ignore
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { dirname } from 'node:path'
import pkg from '../package.json'

app.whenReady().then(() => {
  // @ts-ignore
  const __filename = fileURLToPath(import.meta.url)
  const __dirname = dirname(__filename)
  const win = new BrowserWindow({
    title: pkg.title,
    icon: path.join(__dirname, '../public/favicon.ico'),
    width: 1200,
    height: 800
  })
  win.setMenu(null)

  // You can use `process.env.VITE_DEV_SERVER_URL` when the vite command is called `serve`
  if (process.env.VITE_DEV_SERVER_URL) {
    win.loadURL(process.env.VITE_DEV_SERVER_URL)
  } else {
    // Load your file
    win.loadFile('docs/index.html')
  }
})
