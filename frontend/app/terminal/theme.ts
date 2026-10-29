import type { ITheme } from '@xterm/xterm'

export type ThemeName = 'phosphor' | 'amber' | 'ice' | 'redpill' | 'neon'

/** Интенсивности эффектов темы, 0..1. Команда /fx умножает их на выбранный уровень. */
export interface FxProfile {
  scanlines: number
  glow: number
  flicker: number
  glitch: number
}

export interface Theme {
  bg: string
  fg: string
  cursor: string
  accent: string
  /** Индексы 0-7 обычные цвета ANSI, 8-15 яркие. Весь вывод красится через них. */
  ansi: string[]
  fx: FxProfile
}

const CRT: FxProfile = { scanlines: 0.5, glow: 0.6, flicker: 0.3, glitch: 0 }

export const THEMES: Record<ThemeName, Theme> = {
  phosphor: {
    bg: '#020b05', fg: '#33ff66', cursor: '#33ff66', accent: '#33ff66',
    ansi: ['#020b05', '#ff4d4d', '#33ff66', '#d7ff4d', '#2bb673', '#7dffb0', '#4dffd2', '#b6ffcb',
      '#0d3b1d', '#ff7a7a', '#66ff8c', '#eaff7a', '#4ad98f', '#a5ffc9', '#7affe0', '#e6fff0'],
    fx: CRT,
  },
  amber: {
    bg: '#0d0700', fg: '#ffb000', cursor: '#ffb000', accent: '#ffb000',
    ansi: ['#0d0700', '#ff5c33', '#c8e04d', '#ffd24d', '#d98f1f', '#ff9d5c', '#ffd9a0', '#ffe9bf',
      '#3b2600', '#ff8a66', '#ddf06f', '#ffe27a', '#f0a840', '#ffb98a', '#ffe6bf', '#fff4e0'],
    fx: CRT,
  },
  ice: {
    bg: '#040b12', fg: '#7fd4ff', cursor: '#7fd4ff', accent: '#7fd4ff',
    ansi: ['#040b12', '#ff6b8a', '#6bffd0', '#ffe27a', '#4da3ff', '#b58cff', '#7fd4ff', '#d6efff',
      '#12283a', '#ff93aa', '#93ffe0', '#fff0a8', '#7dbcff', '#cdb3ff', '#a8e4ff', '#f2faff'],
    fx: { scanlines: 0.2, glow: 0.3, flicker: 0, glitch: 0 },
  },
  redpill: {
    bg: '#0d0203', fg: '#ff3b3b', cursor: '#ff3b3b', accent: '#ff3b3b',
    ansi: ['#0d0203', '#ff3b3b', '#ff9d5c', '#ffd24d', '#c2185b', '#ff5ca8', '#ff8a8a', '#ffd6d6',
      '#3b0a0e', '#ff6b6b', '#ffb98a', '#ffe27a', '#e0407f', '#ff8ac4', '#ffb3b3', '#fff0f0'],
    fx: { scanlines: 0.5, glow: 0.6, flicker: 0.3, glitch: 0.6 },
  },
  neon: {
    bg: '#0a0614', fg: '#c8f7ff', cursor: '#ff2bd6', accent: '#00f0ff',
    ansi: ['#0a0614', '#ff3860', '#39ff14', '#fcee09', '#3b82ff', '#ff2bd6', '#00f0ff', '#e6e1ff',
      '#2a1d4d', '#ff6b88', '#7dff5c', '#fff56b', '#6ea0ff', '#ff6be3', '#6bf7ff', '#ffffff'],
    fx: { scanlines: 0.7, glow: 1, flicker: 0.4, glitch: 1 },
  },
}

export const THEME_NAMES = Object.keys(THEMES) as ThemeName[]
export const isThemeName = (v: string): v is ThemeName => v in THEMES

const KEY = 'gl1tch.theme'

export function loadThemeName(): ThemeName {
  try {
    const v = localStorage.getItem(KEY)
    if (v && isThemeName(v)) return v
  } catch {
    // ignore
  }
  return 'phosphor'
}

export function saveThemeName(name: ThemeName) {
  try {
    localStorage.setItem(KEY, name)
  } catch {
    // ignore
  }
}

const ANSI_KEYS = [
  'black', 'red', 'green', 'yellow', 'blue', 'magenta', 'cyan', 'white',
  'brightBlack', 'brightRed', 'brightGreen', 'brightYellow',
  'brightBlue', 'brightMagenta', 'brightCyan', 'brightWhite',
] as const

export function toXtermTheme(t: Theme): ITheme {
  const theme: ITheme = {
    background: t.bg,
    foreground: t.fg,
    cursor: t.cursor,
    cursorAccent: t.bg,
    selectionBackground: `${t.accent}44`,
  }
  ANSI_KEYS.forEach((key, i) => {
    theme[key] = t.ansi[i]
  })
  return theme
}

export function applyCssVars(root: HTMLElement, t: Theme) {
  const doc = document.documentElement
  doc.style.setProperty('--bg', t.bg)
  doc.style.setProperty('--fg', t.fg)
  doc.style.setProperty('--accent', t.accent)
  root.style.setProperty('--glow', t.accent)
}
