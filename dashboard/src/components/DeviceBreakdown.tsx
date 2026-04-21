import type { DeviceStat } from "../api";

interface Props {
    devices: DeviceStat[];
    loading: boolean;
}

const DEVICE_LABELS: Record<string, string> = {
    Desktop: "Desktop",
    Tablet: "Tablet",
    Mobile: "Mobile",
};

export function DeviceBreakdown({ devices, loading }: Props) {
    const total = devices.reduce((sum, d) => sum + d.count, 0) || 1;

    return (
        <div className="card">
            <div className="card-header">
                <span className="card-title">Devices</span>
                <span className="card-meta">By screen width</span>
            </div>
            <div className="card-body">
                {loading ? (
                    <div className="state-center" style={{ height: 120 }}>
                        <div className="spinner" />
                    </div>
                ) : devices.length === 0 ? (
                    <div className="state-center" style={{ height: 120 }}>No data yet</div>
                ) : (
                    <div className="device-list">
                        {devices.map((d) => {
                            const pct = Math.round((d.count / total) * 100);
                            return (
                                <div className="device-row" key={d.device}>
                                    <div className="device-row-header">
                                        <span className="device-name">{DEVICE_LABELS[d.device] ?? d.device}</span>
                                        <span className="device-stats">
                                            <span className="device-pct" style={{ color: "var(--text-primary)" }}>
                                                {pct}%
                                            </span>
                                            <span className="device-count">{d.count.toLocaleString()}</span>
                                        </span>
                                    </div>
                                    <div className="bar-bg">
                                        <div
                                            className="bar-fill"
                                            style={{ width: `${pct}%`, opacity: 0.5 }}
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
