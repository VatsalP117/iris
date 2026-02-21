import type { DeviceStat } from "../api";

interface Props {
    devices: DeviceStat[];
    loading: boolean;
}

const DEVICE_COLORS: Record<string, string> = {
    Desktop: "var(--accent)",
    Tablet: "var(--yellow)",
    Mobile: "var(--green)",
};

export function DeviceBreakdown({ devices, loading }: Props) {
    const total = devices.reduce((sum, d) => sum + d.count, 0) || 1;

    return (
        <div className="card">
            <div className="card-header">
                <span className="card-title">Devices</span>
                <span style={{ fontSize: 12, color: "var(--text-muted)" }}>
                    By screen width
                </span>
            </div>
            <div className="card-body">
                {loading ? (
                    <div className="state-center">
                        <div className="spinner" />
                        Loading…
                    </div>
                ) : devices.length === 0 ? (
                    <div className="state-center">No data yet</div>
                ) : (
                    <div className="device-list">
                        {devices.map((d) => {
                            const pct = Math.round((d.count / total) * 100);
                            const color = DEVICE_COLORS[d.device] ?? "var(--accent)";
                            return (
                                <div className="device-row" key={d.device}>
                                    <div className="device-row-header">
                                        <span className="device-name">{d.device}</span>
                                        <span className="device-count">
                                            {pct}%{" "}
                                            <span style={{ color: "var(--text-muted)", fontWeight: 400 }}>
                                                ({d.count.toLocaleString()})
                                            </span>
                                        </span>
                                    </div>
                                    <div className="bar-bg">
                                        <div
                                            className="bar-fill"
                                            style={{ width: `${pct}%`, background: color }}
                                        />
                                    </div>
                                </div>
                            );
                        })}
                    </div>
                )}
            </div>
        </div>
    );
}
