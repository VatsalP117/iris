import React from "react";
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip } from "recharts";
import type { DeviceStat } from "../api";

interface Props {
    devices: DeviceStat[];
    loading: boolean;
}

const DEVICE_COLORS: Record<string, string> = {
    Desktop: "#0070f3",
    Tablet: "#0088ff",
    Mobile: "#0099ff",
};

const DEVICE_ICONS: Record<string, React.ReactNode> = {
    Desktop: (
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <rect x="2" y="3" width="20" height="14" rx="2" ry="2"/>
            <line x1="8" y1="21" x2="16" y2="21"/>
            <line x1="12" y1="17" x2="12" y2="21"/>
        </svg>
    ),
    Tablet: (
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <rect x="4" y="2" width="16" height="20" rx="2" ry="2"/>
            <line x1="12" y1="18" x2="12.01" y2="18"/>
        </svg>
    ),
    Mobile: (
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <rect x="5" y="2" width="14" height="20" rx="2" ry="2"/>
            <line x1="12" y1="18" x2="12.01" y2="18"/>
        </svg>
    ),
};

const CustomTooltip = ({ active, payload }: any) => {
    if (!active || !payload?.length) return null;
    const data = payload[0].payload;
    return (
        <div
            style={{
                background: "var(--bg-secondary)",
                border: "1px solid var(--border)",
                borderRadius: "var(--radius-md)",
                padding: "10px 14px",
                fontSize: 13,
            }}
        >
            <div style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 4 }}>
                <span
                    style={{
                        width: 10,
                        height: 10,
                        borderRadius: "50%",
                        background: data.fill,
                    }}
                />
                <span style={{ color: "var(--text-secondary)", fontWeight: 500 }}>{data.name}</span>
            </div>
            <div style={{ color: "var(--text-primary)", fontWeight: 600, fontSize: 15 }}>
                {data.percent}%
            </div>
            <div style={{ color: "var(--text-tertiary)", fontSize: 12 }}>
                {data.value.toLocaleString()} visitors
            </div>
        </div>
    );
};

export function DeviceBreakdown({ devices, loading }: Props) {
    const total = devices.reduce((sum, d) => sum + d.count, 0) || 1;

    const chartData = devices.map((d) => ({
        name: d.device,
        value: d.count,
        percent: Math.round((d.count / total) * 100),
        fill: DEVICE_COLORS[d.device] ?? "#3b82f6",
    }));

    return (
        <div className="card">
            <div className="card-header">
                <span className="card-title">Devices</span>
                <span className="card-meta">By screen width</span>
            </div>
            <div className="card-body">
                {loading ? (
                    <div className="state-center" style={{ height: 200 }}>
                        <div className="spinner" />
                        Loading…
                    </div>
                ) : devices.length === 0 ? (
                    <div className="state-center" style={{ height: 200 }}>No data yet</div>
                ) : (
                    <div className="donut-container">
                        <div className="donut-chart">
                            <ResponsiveContainer width="100%" height={200}>
                                <PieChart>
                                    <Pie
                                        data={chartData}
                                        cx="50%"
                                        cy="50%"
                                        innerRadius={60}
                                        outerRadius={85}
                                        paddingAngle={3}
                                        dataKey="value"
                                        stroke="none"
                                    >
                                        {chartData.map((entry, index) => (
                                            <Cell key={`cell-${index}`} fill={entry.fill} />
                                        ))}
                                    </Pie>
                                    <Tooltip content={<CustomTooltip />} />
                                </PieChart>
                            </ResponsiveContainer>
                            <div className="donut-center">
                                <span className="donut-center-value">{total.toLocaleString()}</span>
                                <span className="donut-center-label">Total</span>
                            </div>
                        </div>
                        <div className="donut-legend">
                            {chartData.map((item) => (
                                <div key={item.name} className="donut-legend-item">
                                    <div className="donut-legend-icon" style={{ color: item.fill }}>
                                        {DEVICE_ICONS[item.name]}
                                    </div>
                                    <div className="donut-legend-info">
                                        <div className="donut-legend-name">{item.name}</div>
                                        <div className="donut-legend-stats">
                                            <span className="donut-legend-percent" style={{ color: item.fill }}>
                                                {item.percent}%
                                            </span>
                                            <span className="donut-legend-count">
                                                {item.value.toLocaleString()}
                                            </span>
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
}
