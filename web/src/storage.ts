const VID_KEY = "iris_vid";
const VID_DAY_KEY = "iris_vid_day";
const VID_LOCK_NAME = "iris_visitor_id";
const SID_KEY = "iris_sid";
const SID_LAST_ACTIVITY_KEY = "iris_sid_last_activity";
const SESSION_INACTIVITY_MS = 30 * 60 * 1000;

const memoryVisitors = new Map<string, { id: string; day: string }>();
const memorySessions = new Map<string, { id: string; lastActivity: number }>();

export function generateId(): string {
    if (typeof crypto !== "undefined" && crypto.randomUUID) {
        return crypto.randomUUID();
    }
    // Fallback for older environments
    return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, (c) => {
        const r = (Math.random() * 16) | 0;
        return (c === "x" ? r : (r & 0x3) | 0x8).toString(16);
    });
}

function storageKey(base: string, siteId: string): string {
    return `${base}:${siteId}`;
}

function currentDateKey(timezone: string): string {
    const parts = new Intl.DateTimeFormat("en-US", {
        timeZone: timezone,
        year: "numeric",
        month: "2-digit",
        day: "2-digit",
    }).formatToParts(new Date());
    const values = Object.fromEntries(parts.map((part) => [part.type, part.value]));
    return `${values.year}-${values.month}-${values.day}`;
}

/**
 * Returns an anonymous visitor ID that rotates at midnight in the site's
 * configured timezone and is isolated from other Iris sites on the origin.
 */
function readVisitorId(siteId: string, timezone: string): string {
    const today = currentDateKey(timezone);
    const idKey = storageKey(VID_KEY, siteId);
    const dayKey = storageKey(VID_DAY_KEY, siteId);

    try {
        const savedDay = localStorage.getItem(dayKey);
        let vid = localStorage.getItem(idKey);
        if (!vid || savedDay !== today) {
            vid = generateId();
            localStorage.setItem(idKey, vid);
            localStorage.setItem(dayKey, today);
        }
        return vid;
    } catch {
        const current = memoryVisitors.get(siteId);
        if (!current || current.day !== today) {
            const created = { id: generateId(), day: today };
            memoryVisitors.set(siteId, created);
            return created.id;
        }
        return current.id;
    }
}

export interface VisitorIdentity {
    current: string;
    ready: Promise<string>;
}

export function getVisitorIdentity(siteId: string, timezone: string): VisitorIdentity {
    const read = () => readVisitorId(siteId, timezone);
    const identity: VisitorIdentity = {
        current: read(),
        ready: Promise.resolve(""),
    };
    const coordinatedId =
        typeof navigator !== "undefined" && navigator.locks
            ? navigator.locks.request(storageKey(VID_LOCK_NAME, siteId), read)
            : Promise.resolve(identity.current);
    identity.ready = coordinatedId
        .then((vid) => {
            identity.current = vid;
            return vid;
        })
        .catch(() => identity.current);
    return identity;
}

/**
 * Returns an anonymous session ID shared by same-origin tabs for this site. The
 * session rolls after 30 minutes without tracked activity.
 */
export function getSessionId(siteId: string): string {
    const now = Date.now();
    const idKey = storageKey(SID_KEY, siteId);
    const activityKey = storageKey(SID_LAST_ACTIVITY_KEY, siteId);
    try {
        let sid = localStorage.getItem(idKey);
        const lastActivity = Number(localStorage.getItem(activityKey) || "0");
        if (!sid || now - lastActivity > SESSION_INACTIVITY_MS) {
            sid = generateId();
            localStorage.setItem(idKey, sid);
        }
        localStorage.setItem(activityKey, String(now));
        return sid;
    } catch {
        const current = memorySessions.get(siteId);
        if (!current || now - current.lastActivity > SESSION_INACTIVITY_MS) {
            const created = { id: generateId(), lastActivity: now };
            memorySessions.set(siteId, created);
            return created.id;
        }
        current.lastActivity = now;
        return current.id;
    }
}
