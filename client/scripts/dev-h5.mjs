import { execFileSync, spawn } from 'node:child_process'

const targetPort = Number.parseInt(process.env.TARO_H5_PORT ?? '10086', 10)

function readLines(command, args) {
  try {
    const output = execFileSync(command, args, {
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'ignore'],
    })

    return output
      .split('\n')
      .map((line) => line.trim())
      .filter(Boolean)
  } catch (error) {
    if (typeof error === 'object' && error && 'status' in error && error.status === 1) {
      return []
    }

    throw error
  }
}

function getListeningPids(port) {
  return readLines('lsof', ['-nP', `-iTCP:${port}`, '-sTCP:LISTEN', '-t'])
}

function getCommand(pid) {
  return readLines('ps', ['-o', 'command=', '-p', pid])[0] ?? ''
}

function isTaroH5Watcher(command) {
  return command.includes('taro build --type h5 --watch')
}

function waitForPortToClear(port, attempts = 20) {
  for (let attempt = 0; attempt < attempts; attempt += 1) {
    if (getListeningPids(port).length === 0) {
      return true
    }

    Atomics.wait(new Int32Array(new SharedArrayBuffer(4)), 0, 0, 200)
  }

  return false
}

const pids = getListeningPids(targetPort)

for (const pid of pids) {
  const command = getCommand(pid)

  if (!isTaroH5Watcher(command)) {
    console.error(
      `[dev:h5] Port ${targetPort} is already in use by a non-Taro process (pid ${pid}): ${command || 'unknown'}`
    )
    process.exit(1)
  }

  console.log(`[dev:h5] Stopping stale Taro H5 watcher on port ${targetPort} (pid ${pid})`)
  process.kill(Number(pid), 'SIGTERM')
}

if (pids.length > 0 && !waitForPortToClear(targetPort)) {
  console.error(`[dev:h5] Port ${targetPort} is still busy after stopping the stale watcher`)
  process.exit(1)
}

const child = spawn('pnpm', ['exec', 'taro', 'build', '--type', 'h5', '--watch'], {
  stdio: 'inherit',
  env: {
    ...process.env,
    TARO_H5_PORT: String(targetPort),
  },
})

child.on('exit', (code, signal) => {
  if (signal) {
    process.kill(process.pid, signal)
  }

  process.exit(code ?? 0)
})
