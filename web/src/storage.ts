const VID_KEY = "iris_vid";
const VID_DAY_KEY = "iris_vid_day";
const SID_KEY = "iris_sid";

let memoryVID = "";
let memoryVIDDay = "";
let memorySID = "";

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

function currentUTCDateKey(): string {
    return new Date().toISOString().slice(0, 10);
}

/**
 * Returns an anonymous visitor ID that rotates once per UTC day.
 * Stays stable for a given browser/profile within the same UTC day.
 */
export function getVisitorId(): string {
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
        if (!memorySID) {
            memorySID = generateId();
        }
        return memorySID;
    }
}
