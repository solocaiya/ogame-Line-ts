import pkg from '../../package.json'

export interface VersionInfo {
  version: string
  releaseNotes: string
  downloadUrl: string
}

// 检查GitHub最新版本
export const checkLatestVersion = async (lastCheckTime: number, updateCheckTime: (time: number) => void): Promise<VersionInfo | null> => {
  const now = Date.now()
  const oneHour = 60 * 60 * 1000 // 1小时

  // 如果距离上次检查不到1小时，跳过
  if (now - lastCheckTime < oneHour) {
    return null
  }

  try {
    const response = await fetch(`https://api.github.com/repos/${pkg.author.name}/${pkg.name}/releases/latest`)

    if (!response.ok) {
      console.error('Failed to fetch latest version:', response.status)
      // 更新检查时间，避免频繁请求失败的API
      updateCheckTime(now)
      throw new Error(`Failed to fetch version: ${response.status}`)
    }

    const data = await response.json()
    const githubVersion = data.tag_name?.replace(/^v/, '') // 移除开头的 'v' (如 v1.2.0 -> 1.2.0)

    // 更新最后检查时间
    updateCheckTime(now)

    // 比较版本号
    if (githubVersion && githubVersion !== pkg.version) {
      return {
        version: githubVersion,
        releaseNotes: data.body || '',
        downloadUrl: `https://github.com/${pkg.author.name}/${pkg.name}/releases/latest`
      }
    }
    return null
  } catch (error) {
    console.error('Error checking version:', error)
    // 更新检查时间，避免频繁请求失败的API
    updateCheckTime(now)
    throw error
  }
}

// 检查是否可以进行版本检查（距离上次检查是否超过1小时）
export const canCheckVersion = (lastCheckTime: number): boolean => {
  const now = Date.now()
  const oneHour = 60 * 60 * 1000 // 1小时
  return now - lastCheckTime >= oneHour
}
