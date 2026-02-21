const VID_KEY = "iris_vid";
const SID_KEY = "iris_sid";

function generateId(): string {
    if (typeof crypto !== "undefined" && crypto.randomUUID) {
        return crypto.randomUUID();
    }
    // Fallback for older environments
    return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, (c) => {
        const r = (Math.random() * 16) | 0;
        return (c === "x" ? r : (r & 0x3) | 0x8).toString(16);
    });
}

/**
 * Returns a stable anonymous visitor ID stored in localStorage.
 * Persists across sessions (same device/browser).
 */
export function getVisitorId(): string {
    try {
        let vid = localStorage.getItem(VID_KEY);
        if (!vid) {
            vid = generateId();
            localStorage.setItem(VID_KEY, vid);
        }
        return vid;
    } catch {
        return generateId();
    }
}

/**
 * Returns a session-scoped ID stored in sessionStorage.
 * Generates a new ID per browser tab / session (cleared when tab closes).
 */
export function getSessionId(): string {
    try {
        let sid = sessionStorage.getItem(SID_KEY);
        if (!sid) {
            sid = generateId();
            sessionStorage.setItem(SID_KEY, sid);
        }
        return sid;
    } catch {
        return generateId();
    }
}
