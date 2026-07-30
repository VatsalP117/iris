const VID_KEY = "iris_vid";
const VID_DAY_KEY = "iris_vid_day";
const VID_LOCK_NAME = "iris_visitor_id";
const SID_KEY = "iris_sid";

let memoryVID = "";
let memoryVIDDay = "";
let memorySID = "";
let sessionInitialized = false;

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

function currentUTCDateKey(): string {
    return new Date().toISOString().slice(0, 10);
}

/**
 * Returns an anonymous visitor ID that rotates once per UTC day.
 * Stays stable for a given browser/profile within the same UTC day.
 */
function readVisitorId(): string {
    const today = currentUTCDateKey();

    try {
        const savedDay = localStorage.getItem(VID_DAY_KEY);
        let vid = localStorage.getItem(VID_KEY);
        if (!vid || savedDay !== today) {
            vid = generateId();
            localStorage.setItem(VID_KEY, vid);
            localStorage.setItem(VID_DAY_KEY, today);
        }
        return vid;
    } catch {
        // Storage can fail in privacy modes. Keep IDs stable in memory for this page lifecycle.
        if (!memoryVID || memoryVIDDay !== today) {
            memoryVID = generateId();
            memoryVIDDay = today;
        }
        return memoryVID;
    }
}

export interface VisitorIdentity {
    current: string;
    ready: Promise<string>;
}

export function getVisitorIdentity(): VisitorIdentity {
    const identity: VisitorIdentity = {
        current: readVisitorId(),
        ready: Promise.resolve(""),
    };
    const coordinatedId =
        typeof navigator !== "undefined" && navigator.locks
            ? navigator.locks.request(VID_LOCK_NAME, readVisitorId)
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
 * Returns a session-scoped ID stored in sessionStorage.
 * Generates a new ID per browser tab / session (cleared when tab closes).
 */
export function getSessionId(): string {
    try {
        let sid = sessionStorage.getItem(SID_KEY);
        if (!sessionInitialized && window.opener && sid) {
            sid = generateId();
            sessionStorage.setItem(SID_KEY, sid);
        }
        if (!sid) {
            sid = generateId();
            sessionStorage.setItem(SID_KEY, sid);
        }
        sessionInitialized = true;
        return sid;
    } catch {
        if (!memorySID) {
            memorySID = generateId();
        }
        sessionInitialized = true;
        return memorySID;
    }
}
