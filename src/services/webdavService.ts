/**
 * WebDAV 同步服务
 * 支持将存档上传到 WebDAV 服务器（如坚果云、Nextcloud、NAS等）
 * 注意：WebDAV 配置存储在 gameStore.webdavConfig 中，与用户数据一起持久化
 */

import type { WebDAVConfig } from '@/types/game'

// 重新导出类型以保持向后兼容
export type { WebDAVConfig }

export interface WebDAVFile {
  name: string
  path: string
  size: number
  lastModified: Date
  isDirectory: boolean
}

// WebDAV 消息 key（用于 i18n）
export const WebDAVMessageKey = {
  // 连接相关
  connectionSuccess: 'webdav.connectionSuccess',
  connectionSuccessDirectoryCreated: 'webdav.connectionSuccessDirectoryCreated',
  authFailed: 'webdav.authFailed',
  directoryNotExist: 'webdav.directoryNotExist',
  connectionFailedHttp: 'webdav.connectionFailedHttp',
  networkError: 'webdav.networkError',
  connectionError: 'webdav.connectionError',

  // 上传相关
  uploadSuccess: 'webdav.uploadSuccess',
  noWritePermission: 'webdav.noWritePermission',
  insufficientStorage: 'webdav.insufficientStorage',
  uploadFailedHttp: 'webdav.uploadFailedHttp',
  uploadError: 'webdav.uploadError',

  // 下载相关
  fileNotExist: 'webdav.fileNotExist',
  downloadFailedHttp: 'webdav.downloadFailedHttp',
  downloadError: 'webdav.downloadError',

  // 列表相关
  listFailedHttp: 'webdav.listFailedHttp',
  listError: 'webdav.listError',

  // 删除相关
  deleteFailedHttp: 'webdav.deleteFailedHttp',
  deleteError: 'webdav.deleteError'
} as const

export type WebDAVMessageKeyType = (typeof WebDAVMessageKey)[keyof typeof WebDAVMessageKey]

export interface WebDAVResult {
  success: boolean
  messageKey: WebDAVMessageKeyType
  messageParams?: Record<string, string | number>
  fileName?: string
  data?: string
  files?: WebDAVFile[]
}

// 构建 Authorization header
const buildAuthHeader = (config: WebDAVConfig): string => {
  const credentials = btoa(`${config.username}:${config.password}`)
  return `Basic ${credentials}`
}

// 规范化 URL 路径
const normalizePath = (serverUrl: string, basePath: string, fileName?: string): string => {
  let url = serverUrl.replace(/\/+$/, '')
  let path = basePath.replace(/^\/+/, '').replace(/\/+$/, '')

  if (path) {
    url = `${url}/${path}`
  }

  if (fileName) {
    url = `${url}/${fileName}`
  }

  return url
}

// 测试 WebDAV 连接
export const testWebDAVConnection = async (config: WebDAVConfig): Promise<WebDAVResult> => {
  try {
    const url = normalizePath(config.serverUrl, config.basePath)

    const response = await fetch(url, {
      method: 'PROPFIND',
      headers: {
        Authorization: buildAuthHeader(config),
        Depth: '0',
        'Content-Type': 'application/xml'
      }
    })

    if (response.ok || response.status === 207) {
      return { success: true, messageKey: WebDAVMessageKey.connectionSuccess }
    }

    if (response.status === 401) {
      return { success: false, messageKey: WebDAVMessageKey.authFailed }
    }

    if (response.status === 404) {
      // 尝试创建目录
      const createResult = await createDirectory(config, config.basePath)
      if (createResult) {
        return { success: true, messageKey: WebDAVMessageKey.connectionSuccessDirectoryCreated }
      }
      return { success: false, messageKey: WebDAVMessageKey.directoryNotExist }
    }

    return { success: false, messageKey: WebDAVMessageKey.connectionFailedHttp, messageParams: { status: response.status } }
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : String(error)
    // CORS 错误的处理
    if (errorMessage.includes('Failed to fetch') || errorMessage.includes('NetworkError')) {
      return { success: false, messageKey: WebDAVMessageKey.networkError }
    }
    return { success: false, messageKey: WebDAVMessageKey.connectionError, messageParams: { error: errorMessage } }
  }
}

// 创建目录
const createDirectory = async (config: WebDAVConfig, path: string): Promise<boolean> => {
  try {
    const url = normalizePath(config.serverUrl, path)

    const response = await fetch(url, {
      method: 'MKCOL',
      headers: {
        Authorization: buildAuthHeader(config)
      }
    })

    return response.ok || response.status === 201
  } catch (error) {
    console.error('Failed to create directory:', error)
    return false
  }
}

// 上传存档到 WebDAV
export const uploadToWebDAV = async (config: WebDAVConfig, data: string, fileName?: string): Promise<WebDAVResult> => {
  try {
    // 生成带时间戳的文件名
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19)
    const actualFileName = fileName || `ogame-save-${timestamp}.json`
    const url = normalizePath(config.serverUrl, config.basePath, actualFileName)

    const response = await fetch(url, {
      method: 'PUT',
      headers: {
        Authorization: buildAuthHeader(config),
        'Content-Type': 'application/json'
      },
      body: data
    })

    if (response.ok || response.status === 201 || response.status === 204) {
      return { success: true, messageKey: WebDAVMessageKey.uploadSuccess, fileName: actualFileName }
    }

    if (response.status === 401) {
      return { success: false, messageKey: WebDAVMessageKey.authFailed }
    }

    if (response.status === 403) {
      return { success: false, messageKey: WebDAVMessageKey.noWritePermission }
    }

    if (response.status === 507) {
      return { success: false, messageKey: WebDAVMessageKey.insufficientStorage }
    }

    return { success: false, messageKey: WebDAVMessageKey.uploadFailedHttp, messageParams: { status: response.status } }
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : String(error)
    return { success: false, messageKey: WebDAVMessageKey.uploadError, messageParams: { error: errorMessage } }
  }
}

// 解析 PROPFIND XML 响应
const parsePropfindResponse = (xml: string, _basePath: string): WebDAVFile[] => {
  const parser = new DOMParser()
  const doc = parser.parseFromString(xml, 'application/xml')
  const responses = doc.getElementsByTagNameNS('DAV:', 'response')
  const files: WebDAVFile[] = []

  for (let i = 0; i < responses.length; i++) {
    const response = responses[i]
    if (!response) continue
    const href = response.getElementsByTagNameNS('DAV:', 'href')[0]?.textContent || ''
    const displayName = response.getElementsByTagNameNS('DAV:', 'displayname')[0]?.textContent
    const contentLength = response.getElementsByTagNameNS('DAV:', 'getcontentlength')[0]?.textContent
    const lastModified = response.getElementsByTagNameNS('DAV:', 'getlastmodified')[0]?.textContent
    const resourceType = response.getElementsByTagNameNS('DAV:', 'resourcetype')[0]
    const isCollection = resourceType ? resourceType.getElementsByTagNameNS('DAV:', 'collection').length > 0 : false

    // 解码 URL 编码的路径
    const decodedHref = decodeURIComponent(href)
    const fileName = displayName || decodedHref.split('/').filter(Boolean).pop() || ''

    // 跳过目录本身和非 JSON 文件
    if (isCollection) continue
    if (!fileName.endsWith('.json')) continue

    files.push({
      name: fileName,
      path: decodedHref,
      size: parseInt(contentLength || '0', 10),
      lastModified: lastModified ? new Date(lastModified) : new Date(),
      isDirectory: false
    })
  }

  // 按修改时间降序排序（最新的在前）
  return files.sort((a, b) => b.lastModified.getTime() - a.lastModified.getTime())
}

// 列出 WebDAV 目录中的存档文件
export const listWebDAVFiles = async (config: WebDAVConfig): Promise<WebDAVResult> => {
  try {
    const url = normalizePath(config.serverUrl, config.basePath)

    const response = await fetch(url, {
      method: 'PROPFIND',
      headers: {
        Authorization: buildAuthHeader(config),
        Depth: '1',
        'Content-Type': 'application/xml'
      }
    })

    if (!response.ok && response.status !== 207) {
      if (response.status === 401) {
        return { success: false, messageKey: WebDAVMessageKey.authFailed }
      }
      if (response.status === 404) {
        return { success: false, messageKey: WebDAVMessageKey.directoryNotExist }
      }
      return { success: false, messageKey: WebDAVMessageKey.listFailedHttp, messageParams: { status: response.status } }
    }

    const xml = await response.text()
    const files = parsePropfindResponse(xml, config.basePath)

    return { success: true, messageKey: WebDAVMessageKey.connectionSuccess, files }
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : String(error)
    return { success: false, messageKey: WebDAVMessageKey.listError, messageParams: { error: errorMessage } }
  }
}

// 从 WebDAV 下载存档
export const downloadFromWebDAV = async (config: WebDAVConfig, fileName: string): Promise<WebDAVResult> => {
  try {
    const url = normalizePath(config.serverUrl, config.basePath, fileName)

    const response = await fetch(url, {
      method: 'GET',
      headers: {
        Authorization: buildAuthHeader(config)
      }
    })

    if (!response.ok) {
      if (response.status === 401) {
        return { success: false, messageKey: WebDAVMessageKey.authFailed }
      }
      if (response.status === 404) {
        return { success: false, messageKey: WebDAVMessageKey.fileNotExist }
      }
      return { success: false, messageKey: WebDAVMessageKey.downloadFailedHttp, messageParams: { status: response.status } }
    }

    const data = await response.text()
    return { success: true, messageKey: WebDAVMessageKey.connectionSuccess, data }
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : String(error)
    return { success: false, messageKey: WebDAVMessageKey.downloadError, messageParams: { error: errorMessage } }
  }
}

// 删除 WebDAV 文件
export const deleteFromWebDAV = async (config: WebDAVConfig, fileName: string): Promise<WebDAVResult> => {
  try {
    const url = normalizePath(config.serverUrl, config.basePath, fileName)

    const response = await fetch(url, {
      method: 'DELETE',
      headers: {
        Authorization: buildAuthHeader(config)
      }
    })

    if (response.ok || response.status === 204) {
      return { success: true, messageKey: WebDAVMessageKey.connectionSuccess }
    }

    if (response.status === 401) {
      return { success: false, messageKey: WebDAVMessageKey.authFailed }
    }

    if (response.status === 404) {
      return { success: true, messageKey: WebDAVMessageKey.connectionSuccess } // 文件不存在也视为删除成功
    }

    return { success: false, messageKey: WebDAVMessageKey.deleteFailedHttp, messageParams: { status: response.status } }
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : String(error)
    return { success: false, messageKey: WebDAVMessageKey.deleteError, messageParams: { error: errorMessage } }
  }
}
