interface Token {
  value: string
  start: number
}

/** Разбор по пробелам с кавычками. Обратный слэш экранирует только кавычки и сам слэш: \n остаётся как есть. */
export function tokenize(input: string): Token[] {
  const tokens: Token[] = []
  let i = 0
  while (i < input.length) {
    while (i < input.length && /\s/.test(input[i]!)) i++
    if (i >= input.length) break

    const start = i
    let value = ''
    let quote: string | null = null
    while (i < input.length) {
      const ch = input[i]!
      const next = input[i + 1]
      if (ch === '\\' && next !== undefined && (next === '"' || next === "'" || next === '\\')) {
        value += next
        i += 2
        continue
      }
      if (quote) {
        if (ch === quote) quote = null
        else value += ch
        i++
        continue
      }
      if (ch === '"' || ch === "'") {
        quote = ch
        i++
        continue
      }
      if (/\s/.test(ch)) break
      value += ch
      i++
    }
    tokens.push({ value, start })
  }
  return tokens
}

export function splitCommand(line: string) {
  const trimmed = line.trim()
  if (!trimmed) return null
  const slash = trimmed.startsWith('/')
  const body = slash ? trimmed.slice(1) : trimmed
  const m = body.match(/^(\S+)\s*(.*)$/s)
  return m ? { slash, name: m[1]!.toLowerCase(), argsRaw: m[2]! } : null
}

export function parseArgs(argsRaw: string, spec: { argc?: number; rest?: boolean } = {}): string[] {
  const tokens = tokenize(argsRaw)
  if (spec.rest && spec.argc && tokens.length >= spec.argc) {
    const head = tokens.slice(0, spec.argc - 1).map((t) => t.value)
    return [...head, argsRaw.slice(tokens[spec.argc - 1]!.start).trim()]
  }
  return tokens.map((t) => t.value)
}
