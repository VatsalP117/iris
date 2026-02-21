import type { ReferrerStat } from "../api";

interface Props {
    referrers: ReferrerStat[];
    loading: boolean;
}

function cleanReferrer(ref: string): string {
    try {
        const url = new URL(ref);
        return url.hostname.replace(/^www\./, "");
    } catch {
        return ref;
    }
}

export function TopReferrers({ referrers, loading }: Props) {
    const max = referrers[0]?.visitors ?? 1;

    return (
        <div className="card">
            <div className="card-header">
                <span className="card-title">Top Referrers</span>
                <span style={{ fontSize: 12, color: "var(--text-muted)" }}>
                    {referrers.length} sources
                </span>
            </div>
            <div className="card-body">
                {loading ? (
                    <div className="state-center">
                        <div className="spinner" />
                        Loading…
                    </div>
                ) : referrers.length === 0 ? (
                    <div className="state-center">No referrer data</div>
                ) : (
                    <table className="data-table">
                        <thead>
                            <tr>
                                <th>Source</th>
                                <th className="col-num">Visitors</th>
                                <th className="col-bar"></th>
                            </tr>
                        </thead>
                        <tbody>
                            {referrers.map((r) => (
                                <tr key={r.referrer}>
                                    <td className="col-url">{cleanReferrer(r.referrer)}</td>
                                    <td className="col-num">{r.visitors.toLocaleString()}</td>
                                    <td className="col-bar">
                                        <div className="bar-bg">
                                            <div
                                                className="bar-fill"
                                                style={{
                                                    width: `${(r.visitors / max) * 100}%`,
                                                    background: "var(--green)",
                                                }}
                                            />
                                        </div>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                )}
            </div>
        </div>
    );
}
