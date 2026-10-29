import type { FxProfile } from './theme'

export type FxLevel = 'off' | 'low' | 'full'
export const FX_LEVELS: FxLevel[] = ['off', 'low', 'full']

const KEY = 'gl1tch.fx'
const MULT: Record<FxLevel, number> = { off: 0, low: 0.5, full: 1 }

export function loadFxLevel(): FxLevel {
  try {
    const v = localStorage.getItem(KEY)
    if (v && (FX_LEVELS as string[]).includes(v)) return v as FxLevel
  } catch {
    // ignore
  }
  return matchMedia('(prefers-reduced-motion: reduce)').matches ? 'off' : 'full'
}

export function saveFxLevel(level: FxLevel) {
  try {
    localStorage.setItem(KEY, level)
  } catch {
    // ignore
  }
}

/** Эффекты живут на корневом элементе .crt: CSS-переменные + класс .glitch для всплесков. */
export function createFx(root: HTMLElement) {
  let timer: ReturnType<typeof setTimeout> | null = null
  let glitchChance = 0

  function burst() {
    if (glitchChance <= 0) return
    root.classList.add('glitch')
    setTimeout(() => root.classList.remove('glitch'), 150 + Math.random() * 200)
  }

  function stop() {
    if (timer) clearTimeout(timer)
    timer = null
  }

  function schedule() {
    stop()
    if (glitchChance <= 0) return
    timer = setTimeout(() => {
      burst()
      schedule()
    }, (3000 + Math.random() * 9000) / glitchChance)
  }

  function apply(profile: FxProfile, level: FxLevel) {
    const k = MULT[level]
    root.dataset.fx = level
    root.style.setProperty('--fx-scan', String(profile.scanlines * k))
    root.style.setProperty('--fx-glow', String(profile.glow * k))
    root.style.setProperty('--fx-flicker', String(profile.flicker * k))
    glitchChance = profile.glitch * k
    schedule()
  }

  return { apply, burst, stop }
}
