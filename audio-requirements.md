# 🎵 音效与音乐需求清单

> 太空策略游戏（OGame 类型），Vue 3 前端，Web Audio API 播放。
> 当前使用程序化合成音效作为临时方案，需要替换为专业音频素材。

---

## 一、背景音乐（BGM）

格式要求：`MP3 或 OGG`，循环播放，建议 128kbps，文件体积 < 2MB/首

| # | 场景 | 文件名 | 风格描述 | 时长 | 备注 |
|---|------|--------|----------|------|------|
| B1 | 主界面/总览 | `bgm_overview.mp3` | 太空氛围，宁静深邃，缓慢弦乐 + 电子合成垫音，有科幻感 | 2-3 分钟循环 | 默认播放，最重要 |
| B2 | 建筑页面 | `bgm_building.mp3` | 工业感，轻微机械节奏，金属质感音效点缀，沉稳 | 2 分钟循环 | 建造星球的感觉 |
| B3 | 研究页面 | `bgm_research.mp3` | 科技感，电子脉冲，数据流动感，神秘探索氛围 | 2 分钟循环 | 实验室/科技树 |
| B4 | 舰队/银河 | `bgm_fleet.mp3` | 史诗感，管弦乐 + 电子，太空远征的壮阔感 | 2-3 分钟循环 | 舰队管理/银河地图 |
| B5 | 战斗/警报 | `bgm_battle.mp3` | 紧张激烈，鼓点密集，低音铜管，危机感 | 1.5-2 分钟循环 | 战斗模拟/被攻击时 |
| B6 | 交易/外交 | `bgm_trade.mp3` | 轻松愉快，异域风情，太空集市感 | 2 分钟循环 | 交易所/外交页面 |
| B7 | 登录界面 | `bgm_login.mp3` | 简洁大气，太空主题，引人进入世界 | 1-2 分钟循环 | 可选 |

---

## 二、音效（SFX）

格式要求：`WAV 或 OGG`，短音效 < 2 秒，体积 < 100KB/个

### 2.1 游戏事件音效

| # | 事件 | 文件名 | 描述 | 时长 | 对应 SoundType |
|---|------|--------|------|------|----------------|
| S1 | 建筑升级完成 | `sfx_building_complete.wav` | 叮～上升音阶，成就感，金属质感完成音 | 1-1.5s | `BuildingComplete` |
| S2 | 研究完成 | `sfx_research_complete.wav` | 科技解锁感，电子脉冲 + 闪光音效，未来感 | 1.5-2s | `ResearchComplete` |
| S3 | 舰队出发 | `sfx_fleet_dispatch.wav` | 引擎启动 + 推进器加速，太空飞船离港 | 2-3s | `FleetDispatch` |
| S4 | 舰队返回 | `sfx_fleet_return.wav` | 飞船降落/停靠音，减速 + 着陆感 | 1.5-2s | `FleetReturn` |
| S5 | 战斗胜利 | `sfx_battle_victory.wav` | 凯旋号角 + 爆炸余波，史诗胜利感 | 2-3s | `BattleVictory` |
| S6 | 战斗失败 | `sfx_battle_defeat.wav` | 低沉号角 + 爆炸残响，挫败感 | 2-3s | `BattleDefeat` |
| S7 | 签到/打卡 | `sfx_checkin.wav` | 清脆叮当，奖励获得感，愉悦短促 | 0.5-1s | `CheckIn` |
| S8 | 成就解锁 | `sfx_achievement.wav` | 华丽上升音阶 + 闪光，史诗解锁感 | 2-3s | `AchievementUnlock` |
| S9 | 交易完成 | `sfx_trade_complete.wav` | 金币/资源交换音，轻快确认感 | 0.5-1s | `TradeComplete` |
| S10 | 通用通知 | `sfx_notification.wav` | 柔和提示音，不刺耳，双音交替 | 0.3-0.5s | `Notification` |

### 2.2 UI 交互音效

| # | 事件 | 文件名 | 描述 | 时长 | 对应 SoundType |
|---|------|--------|------|------|----------------|
| S11 | 按钮点击 | `sfx_click.wav` | 清脆短促的点击音，不扰人 | 0.05-0.1s | `Click` |
| S12 | 操作错误 | `sfx_error.wav` | 低沉嗡嗡 + 短促警告，表示操作失败 | 0.3-0.5s | `Error` |

### 2.3 补充音效（当前代码未定义但建议添加）

| # | 事件 | 文件名 | 描述 | 时长 | 建议 SoundType |
|---|------|--------|------|------|----------------|
| S13 | 建筑开始建造 | `sfx_build_start.wav` | 施工启动音，锤子/机械启动 | 0.5s | `BuildStart` |
| S14 | 研究开始 | `sfx_research_start.wav` | 实验室启动音，电子嗡鸣 | 0.5s | `ResearchStart` |
| S15 | 星球切换 | `sfx_planet_switch.wav` | 太空穿梭音，嗖～ | 0.3-0.5s | `PlanetSwitch` |
| S16 | 资源不足 | `sfx_insufficient.wav` | 低沉否定音，嗡～ | 0.3s | `Insufficient` |
| S17 | 消息收到 | `sfx_message.wav` | 通讯接收音，滴滴滴 | 0.5s | `Message` |
| S18 | 殖民成功 | `sfx_colonize.wav` | 着陆 + 旗帜插上感 | 1.5s | `Colonize` |
| S19 | 远征返回 | `sfx_expedition.wav` | 神秘探索音 + 返回确认 | 1-1.5s | `Expedition` |
| S20 | 回收完成 | `sfx_recycle.wav` | 金属回收入库音 | 0.5-1s | `Recycle` |
| S21 | 月球生成 | `sfx_moon_form.wav` | 天体形成音，深沉共鸣 + 碎片碰撞 | 2-3s | `MoonForm` |
| S22 | 页面切换 | `sfx_page_turn.wav` | 轻柔翻页/滑动音 | 0.1-0.2s | `PageTurn` |
| S23 | 弹窗打开 | `sfx_popup_open.wav` | 弹出音，轻快 | 0.1-0.2s | `PopupOpen` |
| S24 | 弹窗关闭 | `sfx_popup_close.wav` | 收起音，干脆 | 0.1s | `PopupClose` |

---

## 三、技术规范

### 文件放置
```
src/assets/audio/
├── bgm/
│   ├── bgm_overview.mp3
│   ├── bgm_building.mp3
│   └── ...
└── sfx/
    ├── sfx_building_complete.wav
    ├── sfx_click.wav
    └── ...
```

### 格式要求
- **BGM**: MP3/OGG, 128kbps, 循环无缝（loop point 首尾平滑）
- **SFX**: WAV(PCM 16bit 44.1kHz) 或 OGG Vorbis
- 所有文件命名用小写 + 下划线，与上表一致
- 每个文件体积尽量小：BGM < 2MB, SFX < 100KB

### 风格统一要求
- 整体风格：**太空科幻**，参考《星际争霸》《EVE Online》《OGame》原版
- 色调偏冷，电子 + 管弦混合
- 音效不能刺耳，要柔和但清晰可辨
- BGM 不能抢注意力，适合长时间后台播放

---

## 四、优先级

**P0（必须）：**
- B1 主界面 BGM
- S1-S6 核心游戏事件音效
- S11 按钮点击
- S12 错误提示

**P1（重要）：**
- B2-B4 场景 BGM
- S7-S10 辅助事件音效
- S13-S17 补充音效

**P2（锦上添花）：**
- B5-B7 特殊场景 BGM
- S18-S24 细节音效

---

## 五、接入方式

音频文件放到 `src/assets/audio/` 后，告诉我路径，我会：
1. 在 `soundManager.ts` 中替换合成音效为真实音频播放
2. BGM 用 `HTMLAudioElement` 播放（支持循环）
3. SFX 用 `AudioBuffer` + Web Audio API 播放（低延迟）
4. 保留现有的音量控制、开关、解锁逻辑不变
