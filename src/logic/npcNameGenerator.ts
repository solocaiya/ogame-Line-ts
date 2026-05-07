/**
 * NPC 名字生成器
 * 为 NPC 生成有代入感的随机名字，支持多语言
 */

export type SupportedLocale = 'zh-CN' | 'zh-TW' | 'en' | 'de' | 'ru' | 'ko' | 'ja'

// 各语言的名字数据
const nameData: Record<
  SupportedLocale,
  {
    prefixes: string[]
    firstNames: string[] // 对于中文是"姓"
    lastNames: string[] // 对于中文是"名"
  }
> = {
  // 简体中文 - 采用"姓+名"格式，可选头衔
  'zh-CN': {
    prefixes: [
      '将军',
      '统帅',
      '元帅',
      '司令',
      '总督',
      '领主',
      '公爵',
      '帝王',
      '大师',
      '长老',
      '圣者',
      '先知',
      '星主',
      '舰长',
      '指挥官',
      '督军'
    ],
    firstNames: [
      // 中文姓氏
      '云',
      '星',
      '龙',
      '凤',
      '墨',
      '风',
      '雷',
      '玄',
      '青',
      '白',
      '朱',
      '银',
      '暗',
      '烈',
      '寒',
      '钢',
      '铁',
      '战',
      '天',
      '皇',
      '帝',
      '王',
      '霸',
      '炎',
      '冰',
      '雪',
      '影',
      '光',
      '夜',
      '苍'
    ],
    lastNames: [
      // 中文名
      '天翔',
      '云龙',
      '星河',
      '凌风',
      '剑心',
      '寒渊',
      '霆威',
      '狼牙',
      '武玄',
      '龙吟',
      '虎啸',
      '鹏飞',
      '辰耀',
      '焰灵',
      '霜羽',
      '电鸣',
      '涛声',
      '壁坚',
      '魂铸',
      '晓破',
      '暮辉',
      '极光',
      '彗尾',
      '傲天',
      '战神',
      '无双',
      '霸天',
      '九霄',
      '破军',
      '七杀'
    ]
  },

  // 繁体中文 - 採用"姓+名"格式，可選頭銜
  'zh-TW': {
    prefixes: [
      '將軍',
      '統帥',
      '元帥',
      '司令',
      '總督',
      '領主',
      '公爵',
      '帝王',
      '大師',
      '長老',
      '聖者',
      '先知',
      '星主',
      '艦長',
      '指揮官',
      '督軍'
    ],
    firstNames: [
      // 中文姓氏
      '雲',
      '星',
      '龍',
      '鳳',
      '墨',
      '風',
      '雷',
      '玄',
      '青',
      '白',
      '朱',
      '銀',
      '暗',
      '烈',
      '寒',
      '鋼',
      '鐵',
      '戰',
      '天',
      '皇',
      '帝',
      '王',
      '霸',
      '炎',
      '冰',
      '雪',
      '影',
      '光',
      '夜',
      '蒼'
    ],
    lastNames: [
      // 中文名
      '天翔',
      '雲龍',
      '星河',
      '凌風',
      '劍心',
      '寒淵',
      '霆威',
      '狼牙',
      '武玄',
      '龍吟',
      '虎嘯',
      '鵬飛',
      '辰耀',
      '焰靈',
      '霜羽',
      '電鳴',
      '濤聲',
      '壁堅',
      '魂鑄',
      '曉破',
      '暮輝',
      '極光',
      '彗尾',
      '傲天',
      '戰神',
      '無雙',
      '霸天',
      '九霄',
      '破軍',
      '七殺'
    ]
  },

  // 英语 - 科幻/西方风格
  en: {
    prefixes: [
      'Admiral',
      'Commander',
      'Captain',
      'General',
      'Marshal',
      'Warlord',
      'Lord',
      'Duke',
      'Baron',
      'Emperor',
      'King',
      'Prince',
      'Dr.',
      'Master',
      'Elder',
      'Oracle',
      'Archon',
      'Overseer',
      'Sentinel',
      'Guardian'
    ],
    firstNames: [
      'Maximus',
      'Aurelius',
      'Tiberius',
      'Corvus',
      'Theron',
      'Darius',
      'Cyrus',
      'Orion',
      'Atlas',
      'Phoenix',
      'Ragnar',
      'Magnus',
      'Zarak',
      'Kael',
      'Vorn',
      'Xander',
      'Drax',
      'Raven',
      'Storm',
      'Nova',
      'Vector',
      'Reaper',
      'Vortex',
      'Aldaris',
      'Zeratul',
      'Artanis',
      'Arcturus',
      'Valerian',
      'Raynor',
      'Fenix'
    ],
    lastNames: [
      'Darkblade',
      'Ironforge',
      'Stormwind',
      'Blackwood',
      'Stargazer',
      'Voidwalker',
      'Skybreaker',
      'Nebula',
      'Ironside',
      'Steelheart',
      'the Wise',
      'the Bold',
      'the Merciless',
      'the Conqueror',
      'the Destroyer',
      'the Eternal'
    ]
  },

  // 德语 - 日耳曼/科幻风格
  de: {
    prefixes: [
      'Admiral',
      'Kommandant',
      'Kapitän',
      'General',
      'Marschall',
      'Kriegsherr',
      'Fürst',
      'Herzog',
      'Baron',
      'Kaiser',
      'König',
      'Prinz',
      'Meister',
      'Ältester',
      'Hüter',
      'Wächter',
      'Richter',
      'Vollstrecker'
    ],
    firstNames: [
      'Wolfgang',
      'Friedrich',
      'Heinrich',
      'Siegfried',
      'Dietrich',
      'Adalbert',
      'Konrad',
      'Gerhard',
      'Reinhard',
      'Lothar',
      'Gunther',
      'Alaric',
      'Baldur',
      'Ragnar',
      'Thorvald',
      'Sigurd',
      'Harald',
      'Bjorn',
      'Fenris',
      'Grimwald',
      'Steinhart',
      'Eisenherz',
      'Donner',
      'Blitz',
      'Sturm',
      'Schatten',
      'Flamme',
      'Frost'
    ],
    lastNames: [
      'von Kriegstein',
      'von Donnerfels',
      'von Eisenberg',
      'von Schwarzwald',
      'Sturmbrecher',
      'Sternenkrieger',
      'Schattenläufer',
      'der Weise',
      'der Tapfere',
      'der Gnadenlose',
      'der Eroberer',
      'der Zerstörer',
      'der Ewige',
      'der Furchtlose',
      'der Mächtige'
    ]
  },

  // 俄语 - 斯拉夫/科幻风格
  ru: {
    prefixes: [
      'Адмирал',
      'Командир',
      'Капитан',
      'Генерал',
      'Маршал',
      'Воевода',
      'Князь',
      'Герцог',
      'Барон',
      'Царь',
      'Император',
      'Владыка',
      'Мастер',
      'Старейшина',
      'Страж',
      'Хранитель',
      'Судья',
      'Каратель'
    ],
    firstNames: [
      'Владимир',
      'Александр',
      'Дмитрий',
      'Николай',
      'Сергей',
      'Андрей',
      'Михаил',
      'Игорь',
      'Ярослав',
      'Святослав',
      'Борис',
      'Олег',
      'Руслан',
      'Богдан',
      'Вадим',
      'Громовой',
      'Молния',
      'Буря',
      'Тень',
      'Пламя',
      'Мороз',
      'Сталь',
      'Железо',
      'Космос',
      'Звезда',
      'Комета',
      'Туман',
      'Вихрь'
    ],
    lastNames: [
      'Громобой',
      'Железнов',
      'Стальной',
      'Черный',
      'Звездочет',
      'Пустоход',
      'Небоход',
      'Мудрый',
      'Храбрый',
      'Беспощадный',
      'Завоеватель',
      'Разрушитель',
      'Вечный',
      'Бесстрашный',
      'Могучий',
      'Непобедимый'
    ]
  },

  // 韩语 - 韩国风格，采用"姓+名"格式
  ko: {
    prefixes: ['제독', '사령관', '함장', '장군', '원수', '군주', '영주', '황제', '왕', '장로', '현자', '수호자', '지휘관', '총독'],
    firstNames: [
      // 韩国姓氏
      '김',
      '이',
      '박',
      '최',
      '정',
      '강',
      '조',
      '윤',
      '장',
      '임',
      '한',
      '오',
      '서',
      '신',
      '권',
      '황',
      '안',
      '송',
      '류',
      '홍'
    ],
    lastNames: [
      // 韩国名字（带有科幻/古风感）
      '천룡',
      '성하',
      '용현',
      '검심',
      '풍화',
      '뇌정',
      '천랑',
      '현무',
      '은하',
      '성진',
      '광년',
      '열염',
      '뇌전',
      '광풍',
      '철벽',
      '강혼',
      '극광',
      '태양',
      '별빛',
      '우주'
    ]
  },

  // 日语 - 日本风格，采用"姓+名"格式
  ja: {
    prefixes: ['提督', '司令官', '艦長', '将軍', '元帥', '領主', '皇帝', '王', '師匠', '長老', '賢者', '守護者', '指揮官', '総督'],
    firstNames: [
      // 日本姓氏
      '山本',
      '田中',
      '佐藤',
      '鈴木',
      '高橋',
      '伊藤',
      '渡辺',
      '中村',
      '小林',
      '加藤',
      '吉田',
      '山田',
      '松本',
      '井上',
      '木村',
      '林',
      '斎藤',
      '清水',
      '山口',
      '阿部'
    ],
    lastNames: [
      // 日本名字（带有科幻/古风感）
      '龍馬',
      '星河',
      '雷神',
      '風神',
      '武蔵',
      '銀河',
      '彗星',
      '流星',
      '閃光',
      '烈火',
      '氷雪',
      '雷電',
      '疾風',
      '鋼鉄',
      '黎明',
      '黄昏',
      '光輝',
      '暁',
      '蒼天',
      '紅蓮'
    ]
  }
}

// 使用 id 作为种子的伪随机数生成器
const seededRandom = (seed: string): (() => number) => {
  let hash = 0
  for (let i = 0; i < seed.length; i++) {
    const char = seed.charCodeAt(i)
    hash = (hash << 5) - hash + char
    hash = hash & hash
  }

  return () => {
    hash = (hash * 1103515245 + 12345) & 0x7fffffff
    return hash / 0x7fffffff
  }
}

/**
 * 根据语言代码获取对应的名字数据
 */
const getNameDataForLocale = (locale: string): (typeof nameData)[SupportedLocale] => {
  // 标准化语言代码
  const normalizedLocale = locale.toLowerCase()

  if (normalizedLocale.startsWith('zh-tw') || normalizedLocale.startsWith('zh-hant')) {
    return nameData['zh-TW']
  }
  if (normalizedLocale.startsWith('zh')) {
    return nameData['zh-CN']
  }
  if (normalizedLocale.startsWith('de')) {
    return nameData['de']
  }
  if (normalizedLocale.startsWith('ru')) {
    return nameData['ru']
  }
  if (normalizedLocale.startsWith('ko')) {
    return nameData['ko']
  }
  if (normalizedLocale.startsWith('ja')) {
    return nameData['ja']
  }

  // 默认英语
  return nameData['en']
}

/**
 * 检查是否是东亚语言（中文、日文、韩文）
 * 这些语言的名字不使用空格分隔
 */
const isEastAsianLocale = (locale: string): boolean => {
  const normalized = locale.toLowerCase()
  return normalized.startsWith('zh') || normalized.startsWith('ja') || normalized.startsWith('ko')
}

/**
 * 生成 NPC 名字
 * @param npcId NPC 的唯一 ID，用作随机种子确保同一 NPC 名字一致
 * @param locale 语言代码，如 'zh-CN', 'en', 'de' 等
 * @param options 可选配置
 * @returns 生成的名字
 */
export const generateNPCName = (
  npcId: string,
  locale: string = 'en',
  options?: {
    includePrefix?: boolean // 是否包含头衔前缀
    includeLastName?: boolean // 是否包含姓氏
    style?: 'simple' | 'full' | 'titled' // 名字风格
  }
): string => {
  const random = seededRandom(npcId)
  const data = getNameDataForLocale(locale)
  const isEastAsian = isEastAsianLocale(locale)

  const style = options?.style ?? 'full'
  const includePrefix = options?.includePrefix ?? style === 'titled'

  // 头衔前缀 (30% 概率，或指定包含)
  let prefix = ''
  if (includePrefix || random() < 0.3) {
    const prefixIndex = Math.floor(random() * data.prefixes.length)
    prefix = data.prefixes[prefixIndex]!
  }

  // 名字主体 (必须) - 对于中文是"姓"
  const firstNameIndex = Math.floor(random() * data.firstNames.length)
  const firstName = data.firstNames[firstNameIndex]!

  // 姓氏/名 (必须包含) - 对于中文是"名"
  const lastNameIndex = Math.floor(random() * data.lastNames.length)
  const lastName = data.lastNames[lastNameIndex]!

  // 根据语言组合名字
  if (isEastAsian) {
    // 东亚语言：姓名不用空格，头衔与名字用空格分隔
    // 格式：[头衔] 姓名  例如：将军 云天翔
    const fullName = firstName + lastName
    return prefix ? `${prefix} ${fullName}` : fullName
  } else {
    // 西方语言：所有部分用空格分隔
    // 格式：[Title] FirstName LastName  例如：Admiral Orion Darkblade
    const parts: string[] = []
    if (prefix) parts.push(prefix)
    parts.push(firstName)
    parts.push(lastName)
    return parts.join(' ')
  }
}

/**
 * 批量生成不重复的 NPC 名字
 * @param count 需要生成的数量
 * @param locale 语言代码
 * @param existingNames 已存在的名字列表（避免重复）
 * @returns 名字数组
 */
export const generateUniqueNPCNames = (count: number, locale: string = 'en', existingNames: string[] = []): string[] => {
  const names: string[] = []
  const usedNames = new Set(existingNames)
  let attempts = 0
  const maxAttempts = count * 10

  while (names.length < count && attempts < maxAttempts) {
    const seed = `npc-${Date.now()}-${attempts}-${Math.random()}`
    const name = generateNPCName(seed, locale)

    if (!usedNames.has(name)) {
      names.push(name)
      usedNames.add(name)
    }
    attempts++
  }

  return names
}

/**
 * 旧版英文名字的特征模式
 * 用于检测是否是旧格式的 NPC 名字
 */
const oldEnglishNamePatterns = [
  // 英文头衔
  /^(Admiral|Commander|Captain|General|Marshal|Warlord|Lord|Duke|Baron|Emperor|King|Prince|Dr\.|Master|Elder|Oracle|Archon|Overseer|Sentinel|Guardian)\s/i,
  // 英文名字特征（常见的科幻名字）
  /\b(Maximus|Aurelius|Tiberius|Corvus|Theron|Darius|Cyrus|Orion|Atlas|Phoenix|Ragnar|Magnus|Zarak|Kael|Vorn|Xander|Drax|Raven|Storm|Nova|Vector|Reaper|Vortex|Aldaris|Zeratul|Artanis|Arcturus|Valerian|Raynor|Fenix)\b/i,
  // 英文姓氏特征
  /\b(Darkblade|Ironforge|Stormwind|Blackwood|Stargazer|Voidwalker|Skybreaker|Nebula|Ironside|Steelheart|the Wise|the Bold|the Merciless|the Conqueror|the Destroyer|the Eternal)\b/i
]

/**
 * 旧版中文名字的特征模式（之前的错误格式）
 * 例如：星辰 征服者、长老 暗影 执法者
 */
const oldChineseNamePatterns = [
  // 旧版后缀（之主、之王等直接作为名字的一部分）
  /\s(之主|之王|天尊|霸主|守护者|征服者|毁灭者|创世者|裁决者|先驱者|开拓者|统治者|复仇者|守望者|执法者|追猎者)$/,
  // 旧版繁体后缀
  /\s(之主|之王|天尊|霸主|守護者|征服者|毀滅者|創世者|裁決者|先驅者|開拓者|統治者|復仇者|守望者|執法者|追獵者)$/
]

/**
 * 检测 NPC 名字是否是原始 ID 格式（如 NPC-npc_10）
 */
const isRawIdFormat = (name: string): boolean => {
  // 匹配 NPC-npc_XX 或 npc_XX 格式
  return /^(NPC-)?npc_\d+$/i.test(name)
}

/**
 * 检测 NPC 名字是否是旧格式
 * @param name NPC 当前的名字
 * @param locale 当前语言
 * @returns 是否是旧格式
 */
export const isOldFormatNPCName = (name: string, locale: string): boolean => {
  // 首先检测是否是原始 ID 格式（所有语言都需要更新）
  if (isRawIdFormat(name)) {
    return true
  }

  const normalizedLocale = locale.toLowerCase()

  // 对于中文用户，检测是否是英文名字或旧版中文格式
  if (normalizedLocale.startsWith('zh')) {
    // 检测英文名字模式
    for (const pattern of oldEnglishNamePatterns) {
      if (pattern.test(name)) {
        return true
      }
    }
    // 检测旧版中文格式
    for (const pattern of oldChineseNamePatterns) {
      if (pattern.test(name)) {
        return true
      }
    }
    return false
  }

  // 对于日语用户，检测是否是英文名字
  if (normalizedLocale.startsWith('ja')) {
    for (const pattern of oldEnglishNamePatterns) {
      if (pattern.test(name)) {
        return true
      }
    }
    return false
  }

  // 对于韩语用户，检测是否是英文名字
  if (normalizedLocale.startsWith('ko')) {
    for (const pattern of oldEnglishNamePatterns) {
      if (pattern.test(name)) {
        return true
      }
    }
    return false
  }

  // 对于德语/俄语等西方语言用户，不需要特别检测
  // 因为英文名字格式与它们类似，不影响代入感
  return false
}

/**
 * 检测 NPC 列表中是否有旧格式的名字
 * @param npcs NPC 列表
 * @param locale 当前语言
 * @returns 需要更新的 NPC 数量
 */
export const countOldFormatNPCs = (npcs: Array<{ id: string; name: string }>, locale: string): number => {
  return npcs.filter(npc => isOldFormatNPCName(npc.name, locale)).length
}

/**
 * 更新 NPC 名字为新格式
 * @param npc NPC 对象
 * @param locale 当前语言
 * @returns 更新后的名字
 */
export const updateNPCName = (npcId: string, locale: string): string => {
  return generateNPCName(npcId, locale)
}

export default {
  generateNPCName,
  generateUniqueNPCNames,
  isOldFormatNPCName,
  countOldFormatNPCs,
  updateNPCName,
  nameData
}
