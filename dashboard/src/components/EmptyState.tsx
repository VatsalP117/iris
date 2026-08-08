interface EmptyStateStep {
    title: string;
    description: string;
}

interface Props {
    eyebrow: string;
    title: string;
    description: string;
    code?: string;
    steps?: EmptyStateStep[];
    compact?: boolean;
}

export function EmptyState({
    eyebrow,
    title,
    description,
    code,
    steps = [],
    compact = false,
}: Props) {
    return (
        <section className={`editorial-empty ${compact ? "is-compact" : ""}`}>
            <div className="empty-copy">
                <span className="eyebrow">{eyebrow}</span>
                <h2>{title}</h2>
                <p>{description}</p>
            </div>

            {code && (
                <div className="empty-code-row">
                    <span>$</span>
                    <code>{code}</code>
                    <button onClick={() => navigator.clipboard?.writeText(code)}>Copy</button>
                </div>
            )}

            {steps.length > 0 && (
                <ol className="empty-steps">
                    {steps.map((step, index) => (
                        <li key={step.title}>
                            <span>{String(index + 1).padStart(2, "0")}</span>
                            <div><strong>{step.title}</strong><p>{step.description}</p></div>
                        </li>
                    ))}
                </ol>
            )}
        </section>
    );
}
