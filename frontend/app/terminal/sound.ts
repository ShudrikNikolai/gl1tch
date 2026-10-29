export type Blip = 'key' | 'ok' | 'error'

const KEY = 'gl1tch.sound'
let enabled = false
let ctx: AudioContext | null = null

const PRESETS: Record<Blip, { freq: number; dur: number; vol: number; type: OscillatorType }> = {
  key: { freq: 1400, dur: 0.015, vol: 0.015, type: 'square' },
  ok: { freq: 880, dur: 0.09, vol: 0.04, type: 'sine' },
  error: { freq: 140, dur: 0.2, vol: 0.06, type: 'sawtooth' },
}

export function loadSound() {
  try {
    enabled = localStorage.getItem(KEY) === '1'
  } catch {
    enabled = false
  }
}

export const isSoundEnabled = () => enabled

export function setSound(on: boolean) {
  enabled = on
  try {
    localStorage.setItem(KEY, on ? '1' : '0')
  } catch {
    // ignore
  }
}

/** AudioContext создаётся лениво на первом звуке, то есть уже после действия пользователя. */
export function blip(kind: Blip) {
  if (!enabled) return
  try {
    ctx ??= new AudioContext()
    if (ctx.state === 'suspended') void ctx.resume()
    const p = PRESETS[kind]
    const osc = ctx.createOscillator()
    const gain = ctx.createGain()
    osc.type = p.type
    osc.frequency.value = p.freq
    gain.gain.setValueAtTime(p.vol, ctx.currentTime)
    gain.gain.exponentialRampToValueAtTime(0.0001, ctx.currentTime + p.dur)
    osc.connect(gain).connect(ctx.destination)
    osc.start()
    osc.stop(ctx.currentTime + p.dur)
  } catch {
    // звук недоступен
  }
}
