import type { PageStat } from "../api";

interface Props {
    pages: PageStat[];
    loading: boolean;
}

export function TopPages({ pages, loading }: Props) {
    const max = pages[0]?.pageviews ?? 1;

    return (
        <div className="card">
            <div className="card-header">
                <span className="card-title">Top Pages</span>
                <span className="card-meta">
                    {pages.length} pages
                </span>
            </div>
            <div className="card-body">
                {loading ? (
                    <div className="state-center">
                        <div className="spinner" />
                    </div>
                ) : pages.length === 0 ? (
                    <div className="state-center">No data yet</div>
                ) : (
                    <table className="data-table">
                        <thead>
                            <tr>
                                <th>Page</th>
                                <th className="col-num">Views</th>
                                <th className="col-bar"></th>
                            </tr>
                        </thead>
                        <tbody>
                            {pages.map((p, i) => (
                                <tr key={p.url} style={{ animationDelay: `${i * 50}ms` }}>
                                    <td className="col-url" title={p.url}>
                                        {p.url.replace(/^https?:\/\/[^/]+/, "") || "/"}
                                    </td>
                                    <td className="col-num">{p.pageviews.toLocaleString()}</td>
                                    <td className="col-bar">
                                        <div className="bar-bg">
                                            <div
                                                className="bar-fill"
                                                style={{ width: `${(p.pageviews / max) * 100}%` }}
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
