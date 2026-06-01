// 前端日志上报工具（admin）
// 对应后端 D1: POST /api/v1/internal/clog/batch
// 设计目标：
//   - 内存缓冲，满 20 条或 5 秒 flush；
//   - beforeunload 时 sendBeacon 同步发送；
//   - 自身错误静默，绝不影响主流程；
//   - 敏感字段（password/token/secret）自动脱敏。
const MSG_MAX = 4 * 1024;
const STACK_MAX = 16 * 1024;
const EXTRA_MAX = 8 * 1024;
const BATCH_SIZE = 20;
const FLUSH_INTERVAL_MS = 5000;
const SENSITIVE_RE = /("?)(password|token|secret|authorization|csrf)("?\s*[:=]\s*"?)([^",}\s]+)/gi;
function safeRelease() {
    try {
        // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
        return typeof __APP_VERSION__ === 'string' ? __APP_VERSION__ : 'dev';
    }
    catch {
        return 'dev';
    }
}
function truncate(s, max) {
    if (!s)
        return s;
    if (s.length <= max)
        return s;
    return s.slice(0, max) + '...[truncated]';
}
function sanitize(s) {
    if (!s)
        return s;
    try {
        return s.replace(SENSITIVE_RE, '$1$2$3***');
    }
    catch {
        return s;
    }
}
function sanitizeExtra(extra) {
    if (!extra)
        return undefined;
    try {
        let json = JSON.stringify(extra);
        json = sanitize(json);
        if (json.length > EXTRA_MAX) {
            json = json.slice(0, EXTRA_MAX) + '"...[truncated]"';
            try {
                return JSON.parse(json);
            }
            catch {
                return { _truncated: true, _raw: json.slice(0, EXTRA_MAX) };
            }
        }
        return JSON.parse(json);
    }
    catch {
        return undefined;
    }
}
export class ClogReporter {
    queue = [];
    timer = null;
    endpoint;
    getAdminId;
    fetchImpl;
    beaconImpl;
    constructor(opts = {}) {
        this.endpoint = opts.endpoint ?? '/api/v1/internal/clog/batch';
        this.getAdminId = opts.getAdminId ?? (() => undefined);
        this.fetchImpl = opts.fetchImpl ?? ((...args) => fetch(...args));
        this.beaconImpl = opts.beaconImpl;
    }
    report(level, message, opts = {}) {
        try {
            const entry = {
                source: 'admin',
                level,
                message: truncate(sanitize(String(message ?? '')), MSG_MAX),
                release: safeRelease(),
            };
            if (opts.stack)
                entry.stack = truncate(sanitize(opts.stack), STACK_MAX);
            const ex = sanitizeExtra(opts.extra);
            if (ex)
                entry.extra = ex;
            if (typeof window !== 'undefined') {
                entry.url = window.location?.href;
                entry.user_agent = window.navigator?.userAgent;
            }
            const adminId = this.getAdminId();
            if (adminId)
                entry.admin_id = adminId;
            this.queue.push(entry);
            if (this.queue.length >= BATCH_SIZE) {
                void this.flush();
            }
            else {
                this.scheduleFlush();
            }
        }
        catch {
            // 静默
        }
    }
    scheduleFlush() {
        if (this.timer)
            return;
        this.timer = setTimeout(() => {
            this.timer = null;
            void this.flush();
        }, FLUSH_INTERVAL_MS);
    }
    async flush() {
        if (this.timer) {
            clearTimeout(this.timer);
            this.timer = null;
        }
        if (this.queue.length === 0)
            return;
        const batch = this.queue.splice(0, BATCH_SIZE);
        const body = JSON.stringify({ logs: batch });
        try {
            await this.send(body);
        }
        catch {
            // 重试一次
            try {
                await this.send(body);
            }
            catch {
                // 仍失败，丢弃，避免循环
            }
        }
    }
    async send(body) {
        const res = await this.fetchImpl(this.endpoint, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body,
            credentials: 'same-origin',
            keepalive: true,
        });
        if (!res.ok)
            throw new Error(`clog upload failed: ${res.status}`);
    }
    flushSync() {
        try {
            if (this.timer) {
                clearTimeout(this.timer);
                this.timer = null;
            }
            if (this.queue.length === 0)
                return;
            const batch = this.queue.splice(0, BATCH_SIZE);
            const body = JSON.stringify({ logs: batch });
            const beacon = this.beaconImpl ??
                (typeof navigator !== 'undefined' && typeof navigator.sendBeacon === 'function'
                    ? (url, data) => navigator.sendBeacon(url, data)
                    : undefined);
            if (beacon) {
                const blob = new Blob([body], { type: 'application/json' });
                const ok = beacon(this.endpoint, blob);
                if (ok)
                    return;
            }
            // 降级：fetch keepalive，无 await
            void this.fetchImpl(this.endpoint, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body,
                credentials: 'same-origin',
                keepalive: true,
            }).catch(() => undefined);
        }
        catch {
            // 静默
        }
    }
    // 测试辅助
    _peekQueue() {
        return this.queue;
    }
}
let singleton = null;
let adminIdGetter = () => undefined;
export function getReporter() {
    if (!singleton) {
        singleton = new ClogReporter({
            getAdminId: () => {
                try {
                    return adminIdGetter();
                }
                catch {
                    return undefined;
                }
            },
        });
    }
    return singleton;
}
export function setReporter(r) {
    singleton = r;
}
export function configureAdminIdGetter(fn) {
    adminIdGetter = fn;
}
export function report(level, message, opts = {}) {
    getReporter().report(level, message, opts);
}
export function installClogLifecycle() {
    if (typeof window === 'undefined')
        return;
    window.addEventListener('beforeunload', () => {
        getReporter().flushSync();
    });
}
