const base = (process.env.BASE_URL ?? 'http://localhost').replace(/\/$/, '')

function assert(ok, message) {
  if (!ok) throw new Error(message)
}

async function request(path, options = {}) {
  const res = await fetch(`${base}${path}`, options)
  const text = await res.text()
  let body
  try {
    body = text ? JSON.parse(text) : undefined
  } catch {
    body = text
  }
  return { res, body }
}

const main = async () => {
  const page = await request('/')
  assert(page.res.ok, `GET / -> ${page.res.status}`)

  const notifications = await request('/api/notifications/healthz')
  assert(notifications.res.ok, `GET /api/notifications/healthz -> ${notifications.res.status}`)

  const unauthorized = await request('/api/users/me')
  assert(unauthorized.res.status === 401, `GET /api/users/me without token -> ${unauthorized.res.status}`)

  const username = `smoke_${Date.now()}`
  const password = 'smoke-password-123'
  const registered = await request('/api/auth/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
  })
  assert(registered.res.ok, `POST /api/auth/register -> ${registered.res.status}: ${JSON.stringify(registered.body)}`)

  const access = registered.body?.tokens?.accessToken
  assert(access, 'register response has no accessToken')

  const me = await request('/api/users/me', {
    headers: { Authorization: `Bearer ${access}` },
  })
  assert(me.res.ok, `GET /api/users/me -> ${me.res.status}`)
  assert(me.body?.user?.username === username, `unexpected /me user: ${JSON.stringify(me.body)}`)

  const boards = await request('/api/v1/boards', {
    headers: { Authorization: `Bearer ${access}` },
  })
  assert(boards.res.ok, `GET /api/v1/boards -> ${boards.res.status}`)

  console.log('smoke: ok')
}

main().catch((err) => {
  console.error(`smoke: ${err.message}`)
  process.exit(1)
})
