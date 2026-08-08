import { useState } from "react";

import { DATE_PRESETS, PresetKey } from "../App";
import type { SiteStat } from "../api";
import { Icon, IconName } from "./Icon";

export type DashboardView = "dashboard" | "sites" | "events" | "vitals";

interface Props {
    children: React.ReactNode;
    view: DashboardView;
    title: string;
    sites: SiteStat[];
    selectedSiteId: string;
    preset: PresetKey;
    loading: boolean;
    onNavigate: (view: DashboardView) => void;
    onSiteChange: (siteId: string) => void;
    onPresetChange: (preset: PresetKey) => void;
    onRefresh: () => void;
}

const NAV_ITEMS: { view: DashboardView; label: string; icon: IconName }[] = [
    { view: "dashboard", label: "Overview", icon: "dashboard" },
    { view: "sites", label: "Sites", icon: "globe" },
    { view: "events", label: "Custom Events", icon: "calendar" },
    { view: "vitals", label: "Web Vitals", icon: "speed" },
];

export function DashboardShell({
    children,
    view,
    title,
    sites,
    selectedSiteId,
    preset,
    loading,
    onNavigate,
    onSiteChange,
    onPresetChange,
    onRefresh,
}: Props) {
    const [mobileNavigationOpen, setMobileNavigationOpen] = useState(false);

    function navigate(nextView: DashboardView) {
        onNavigate(nextView);
        setMobileNavigationOpen(false);
    }

    return (
        <div className="app-shell">
            <aside className={`sidebar ${mobileNavigationOpen ? "is-open" : ""}`}>
                <div className="brand">
                    <div className="brand-mark">
                        {Array.from({ length: 6 }, (_, index) => <span key={index} />)}
                    </div>
                    <div>
                        <strong>Iris</strong>
                        <small>Self-hosted analytics</small>
                    </div>
                </div>

                <nav className="sidebar-nav" aria-label="Primary navigation">
                    {NAV_ITEMS.map((item, index) => (
                        <button
                            className={view === item.view ? "active" : ""}
                            key={item.view}
                            onClick={() => navigate(item.view)}
                        >
                            <b>{String(index + 1).padStart(2, "0")}</b>
                            <Icon name={item.icon} size={18} />
                            <span>{item.label}</span>
                        </button>
                    ))}
                </nav>

                <div className="sidebar-footer">
                    <button className="sidebar-settings">
                        <Icon name="settings" size={18} />
                        <span>Settings</span>
                    </button>
                    <div className="profile">
                        <div className="avatar">VP</div>
                        <div>
                            <strong>Admin User</strong>
                            <small>Administrator</small>
                        </div>
                    </div>
                </div>
            </aside>

            {mobileNavigationOpen && (
                <button
                    aria-label="Close navigation"
                    className="sidebar-scrim"
                    onClick={() => setMobileNavigationOpen(false)}
                />
            )}

            <div className="shell-main">
                <header className="app-bar">
                    <div className="app-bar-title">
                        <button
                            aria-label="Open navigation"
                            className="icon-button mobile-menu"
                            onClick={() => setMobileNavigationOpen(true)}
                        >
                            <Icon name="menu" />
                        </button>
                        <div>
                            <span className="app-bar-eyebrow">Analytics workspace</span>
                            <h1>{title}.</h1>
                        </div>
                    </div>

                    <div className="app-bar-actions">
                        {sites.length > 0 && view !== "sites" && (
                            <label className="select-shell site-select">
                                <span className="sr-only">Site</span>
                                <Icon name="globe" size={16} />
                                <select value={selectedSiteId} onChange={(event) => onSiteChange(event.target.value)}>
                                    {sites.map((site) => (
                                        <option key={site.site_id} value={site.site_id}>
                                            {site.domain || site.site_id}
                                        </option>
                                    ))}
                                </select>
                                <Icon name="chevron-down" size={14} />
                            </label>
                        )}

                        {view !== "sites" && sites.length > 0 && (
                            <label className="select-shell period-select">
                                <span className="sr-only">Date range</span>
                                <Icon name="calendar" size={15} />
                                <select value={preset} onChange={(event) => onPresetChange(event.target.value as PresetKey)}>
                                    {DATE_PRESETS.map((item) => (
                                        <option key={item.key} value={item.key}>{item.label}</option>
                                    ))}
                                </select>
                                <Icon name="chevron-down" size={14} />
                            </label>
                        )}

                        <button
                            aria-label="Refresh analytics"
                            className={`icon-button ${loading ? "is-spinning" : ""}`}
                            disabled={loading}
                            onClick={onRefresh}
                        >
                            <Icon name="refresh" size={18} />
                        </button>
                        <button aria-label="Notifications" className="icon-button notification-button">
                            <Icon name="bell" size={18} />
                            <span />
                        </button>
                        <button aria-label="Help" className="icon-button">
                            <Icon name="help" size={18} />
                        </button>
                    </div>
                </header>

                <main className="page-canvas">
                    <div className="page-container">{children}</div>
                </main>
            </div>
        </div>
    );
}
