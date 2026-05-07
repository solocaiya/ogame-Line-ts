import type { CapacitorConfig } from '@capacitor/cli'

const config: CapacitorConfig = {
  appId: 'games.wenzi.ogame',
  appName: 'OGame Vue Ts',
  webDir: 'docs',
  server: { androidScheme: 'https', cacheControl: 'max-age=31536000' },
  android: {
    buildOptions: { keystorePath: undefined, keystoreAlias: undefined },
    webContentsDebuggingEnabled: false,
    allowMixedContent: false,
    hardwareAcceleration: true
  },
  // 禁用键盘自动调整视口
  plugins: { Keyboard: { resize: 'none' } }
}

export default config
