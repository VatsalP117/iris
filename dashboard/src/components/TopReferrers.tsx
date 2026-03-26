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
                <span className="card-meta">
                    {referrers.length} sources
                </span>
            </div>
            <div className="card-body">
                {loading ? (
                    <div className="state-center">
                        <div className="spinner" />
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
                            {referrers.map((r, i) => (
                                <tr key={r.referrer} style={{ animationDelay: `${i * 50}ms` }}>
                                    <td className="col-url">{cleanReferrer(r.referrer)}</td>
                                    <td className="col-num">{r.visitors.toLocaleString()}</td>
                                    <td className="col-bar">
                                        <div className="bar-bg">
                                            <div
                                                className="bar-fill"
                                                style={{
                                                    width: `${(r.visitors / max) * 100}%`,
                                                    background: "linear-gradient(90deg, #60a5fa, #93c5fd)",
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
