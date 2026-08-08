import { useState } from "react";
import { motion } from "framer-motion";
import {
    Activity,
    ArrowRight,
    BarChart3,
    Check,
    ChevronRight,
    Container,
    Copy,
    Database,
    Globe2,
    LockKeyhole,
    Menu,
    MousePointer2,
    ShieldCheck,
    X,
    Zap,
} from "lucide-react";
import { Link, Outlet } from "react-router-dom";

function GitHubIcon({ size = 17 }: { size?: number }) {
    return (
        <svg
            width={size}
            height={size}
            viewBox="0 0 24 24"
            fill="currentColor"
            aria-hidden="true"
        >
            <path d="M12 .7A11.5 11.5 0 0 0 8.36 23.1c.58.1.79-.25.79-.56v-2.02c-3.22.7-3.9-1.37-3.9-1.37-.53-1.34-1.29-1.7-1.29-1.7-1.05-.72.08-.71.08-.71 1.17.08 1.78 1.2 1.78 1.2 1.04 1.77 2.72 1.26 3.38.96.1-.75.4-1.26.74-1.55-2.57-.29-5.27-1.28-5.27-5.68 0-1.26.45-2.28 1.19-3.09-.12-.29-.52-1.46.11-3.05 0 0 .97-.31 3.16 1.18a10.9 10.9 0 0 1 5.76 0c2.2-1.49 3.16-1.18 3.16-1.18.63 1.59.23 2.76.11 3.05.74.81 1.19 1.83 1.19 3.09 0 4.41-2.71 5.38-5.29 5.67.42.36.79 1.07.79 2.16v3.04c0 .31.21.67.8.56A11.5 11.5 0 0 0 12 .7Z" />
        </svg>
    );
}

const rise = {
    initial: { opacity: 0, y: 24 },
    whileInView: { opacity: 1, y: 0 },
    viewport: { once: true, margin: "-80px" },
    transition: { duration: 0.55, ease: [0.22, 1, 0.36, 1] as const },
};

function IrisMark({ compact = false }: { compact?: boolean }) {
    return (
        <span className={`iris-mark ${compact ? "iris-mark--compact" : ""}`} aria-hidden="true">
            <span />
            <span />
            <span />
            <span />
            <span />
            <span />
        </span>
    );
}

function Nav() {
    const [isOpen, setIsOpen] = useState(false);

    return (
        <nav className="site-nav">
            <div className="nav-inner">
                <Link to="/" className="brand" aria-label="Iris Analytics home">
                    <IrisMark />
                    <span>Iris</span>
                </Link>

                <div className="nav-links">
                    <a href="/#product">Product</a>
                    <a href="/#privacy">Privacy</a>
                    <a href="/#developers">Developers</a>
                    <Link to="/docs">Docs</Link>
                </div>

                <div className="nav-actions">
                    <a href="/docs#self-hosting" className="button button--dark button--small">
                        Deploy Iris
                        <ArrowRight size={15} />
                    </a>
                </div>

                <button
                    type="button"
                    className="nav-toggle"
                    onClick={() => setIsOpen((open) => !open)}
                    aria-label="Toggle navigation"
                    aria-expanded={isOpen}
                >
                    {isOpen ? <X size={22} /> : <Menu size={22} />}
                </button>
            </div>

            {isOpen && (
                <div className="mobile-nav">
                    <a href="/#product" onClick={() => setIsOpen(false)}>Product</a>
                    <a href="/#privacy" onClick={() => setIsOpen(false)}>Privacy</a>
                    <a href="/#developers" onClick={() => setIsOpen(false)}>Developers</a>
                    <Link to="/docs" onClick={() => setIsOpen(false)}>Docs</Link>
                    <a href="/docs#self-hosting" className="button button--dark" onClick={() => setIsOpen(false)}>
                        Deploy Iris
                    </a>
                </div>
            )}
        </nav>
    );
}

function AnalyticsPreview() {
    return (
        <div className="product-stage" aria-label="Iris analytics dashboard preview">
            <div className="stage-orbit stage-orbit--one" />
            <div className="stage-orbit stage-orbit--two" />

            <motion.div
                className="dashboard-window dashboard-window--editorial"
                initial={{ opacity: 0, y: 30, rotateX: 3, rotateZ: 1.25 }}
                animate={{ opacity: 1, y: 0, rotateX: 0, rotateZ: 1.25 }}
                transition={{ delay: 0.25, duration: 0.8, ease: [0.22, 1, 0.36, 1] }}
            >
                <div className="dashboard-topbar">
                    <strong>Iris / Overview</strong>
                    <span className="dashboard-live"><i /> Live</span>
                </div>

                <div className="dashboard-body">
                    <div className="metric-row">
                        <div className="metric">
                            <span>Visitors</span>
                            <strong>8,492</strong>
                        </div>
                        <div className="metric">
                            <span>Pageviews</span>
                            <strong>21.6k</strong>
                        </div>
                        <div className="metric">
                            <span>Growth</span>
                            <strong>+18.4%</strong>
                        </div>
                    </div>

                    <div className="chart-panel">
                        <svg viewBox="0 0 620 185" role="img" aria-label="Rising visitor traffic chart">
                            <defs>
                                <linearGradient id="chartFill" x1="0" x2="0" y1="0" y2="1">
                                    <stop offset="0%" stopColor="#6f5cff" stopOpacity=".22" />
                                    <stop offset="100%" stopColor="#6f5cff" stopOpacity="0" />
                                </linearGradient>
                            </defs>
                            <g className="chart-grid">
                                <line x1="0" y1="35" x2="620" y2="35" />
                                <line x1="0" y1="85" x2="620" y2="85" />
                                <line x1="0" y1="135" x2="620" y2="135" />
                            </g>
                            <path
                                className="chart-area"
                                d="M0,150 C45,142 55,117 98,123 C145,130 154,82 202,96 C246,109 271,66 310,76 C354,88 370,50 414,58 C459,68 476,29 518,43 C561,55 579,17 620,24 L620,185 L0,185 Z"
                            />
                            <path
                                className="chart-line"
                                d="M0,150 C45,142 55,117 98,123 C145,130 154,82 202,96 C246,109 271,66 310,76 C354,88 370,50 414,58 C459,68 476,29 518,43 C561,55 579,17 620,24"
                            />
                            <circle cx="518" cy="43" r="5" />
                        </svg>
                    </div>
                </div>
            </motion.div>
        </div>
    );
}

function SectionLabel({ children }: { children: React.ReactNode }) {
    return (
        <div className="section-label">
            <span />
            {children}
        </div>
    );
}

const highlights = [
    {
        icon: ShieldCheck,
        number: "01",
        title: "Private by default",
        text: "No cookies. No fingerprints. No personal profiles. Just the signals you need to improve your product.",
    },
    {
        icon: Container,
        number: "02",
        title: "One container",
        text: "The Go server, SQLite database, and dashboard ship together. Deploy once and own the whole stack.",
    },
    {
        icon: Zap,
        number: "03",
        title: "Invisible to users",
        text: "Events travel with sendBeacon, off the critical path, so your analytics never become your performance problem.",
    },
];

function ProductSection() {
    return (
        <section id="product" className="section product-section">
            <motion.div className="section-heading" {...rise}>
                <div>
                    <SectionLabel>01 / One clear view</SectionLabel>
                    <h2>Every useful signal.<br /><em>None of the noise.</em></h2>
                </div>
                <p>
                    Iris turns pageviews, events, sources, and Web Vitals into one focused view.
                    No maze of reports. No analyst required.
                </p>
            </motion.div>

            <div className="bento-grid">
                <motion.article className="bento-card bento-card--wide" {...rise}>
                    <div className="card-copy">
                        <span className="card-icon"><BarChart3 size={18} /></span>
                        <h3>Understand the whole journey</h3>
                        <p>Follow traffic from source to page to action, without following the person.</p>
                    </div>
                    <div className="journey-visual">
                        <div className="journey-row">
                            <span><Globe2 size={15} /> google.com</span>
                            <ChevronRight size={15} />
                            <span>/launch</span>
                            <ChevronRight size={15} />
                            <span className="journey-success"><Check size={14} /> signup</span>
                        </div>
                        <div className="journey-bars">
                            <i style={{ width: "88%" }} />
                            <i style={{ width: "62%" }} />
                            <i style={{ width: "38%" }} />
                        </div>
                    </div>
                </motion.article>

                <motion.article className="bento-card bento-card--events" {...rise}>
                    <div className="card-copy">
                        <span className="card-icon"><MousePointer2 size={18} /></span>
                        <h3>Events that explain “why”</h3>
                        <p>Auto-capture the interactions that reveal what people actually use.</p>
                    </div>
                    <div className="event-stream">
                        <div><span className="event-dot event-dot--violet" /> page_view <em>now</em></div>
                        <div><span className="event-dot event-dot--green" /> signup_click <em>3s</em></div>
                        <div><span className="event-dot event-dot--orange" /> docs_search <em>8s</em></div>
                        <div><span className="event-dot" /> outbound_link <em>12s</em></div>
                    </div>
                </motion.article>

                <motion.article className="bento-card bento-card--vitals" {...rise}>
                    <div className="card-copy">
                        <span className="card-icon"><Activity size={18} /></span>
                        <h3>Performance in context</h3>
                        <p>See real-user Web Vitals alongside the pages and releases that shaped them.</p>
                    </div>
                    <div className="vital-score">
                        <div className="score-ring"><span>96</span></div>
                        <div>
                            <strong>Excellent</strong>
                            <span>Real-user score</span>
                        </div>
                    </div>
                </motion.article>
            </div>
        </section>
    );
}

function SignalStorySection() {
    const steps = [
        { number: "01", label: "Source", value: "google.com", detail: "2,931 visits" },
        { number: "02", label: "Page", value: "/docs", detail: "3,891 views" },
        { number: "03", label: "Action", value: "signup_click", detail: "18.4% conversion" },
    ];

    return (
        <section className="signal-story">
            <div className="section signal-story-inner">
                <motion.div className="story-intro" {...rise}>
                    <SectionLabel>02 / The whole journey</SectionLabel>
                    <h2>See what happened.<br /><em>Understand why.</em></h2>
                    <p>
                        Follow the thread from discovery to decision. Iris connects each useful
                        moment without building a profile of the person behind it.
                    </p>
                </motion.div>

                <motion.div className="story-flow" {...rise}>
                    <div className="story-flow-header">
                        <span><i /> Live journey</span>
                        <span>anonymous session · 42s</span>
                    </div>
                    <div className="story-steps">
                        {steps.map((step, index) => (
                            <div className="story-step" key={step.label}>
                                <span className="story-number">{step.number}</span>
                                <div>
                                    <small>{step.label}</small>
                                    <strong>{step.value}</strong>
                                    <span>{step.detail}</span>
                                </div>
                                {index < steps.length - 1 && <ChevronRight size={17} />}
                            </div>
                        ))}
                    </div>
                    <div className="story-caption">
                        <span>One journey, three useful signals</span>
                        <strong>Zero personal identifiers collected</strong>
                    </div>
                </motion.div>
            </div>
        </section>
    );
}

function PrivacySection() {
    return (
        <section id="privacy" className="privacy-section">
            <div className="section privacy-inner">
                <motion.div className="privacy-copy" {...rise}>
                    <SectionLabel>03 / Data sovereignty</SectionLabel>
                    <h2>Your analytics<br />should be yours.</h2>
                    <p>
                        Iris runs on your infrastructure and stores data in your database.
                        There is no third-party data pipeline hiding behind the dashboard.
                    </p>
                    <Link to="/docs" className="text-link">
                        Explore self-hosting <ArrowRight size={17} />
                    </Link>
                </motion.div>

                <motion.div className="ownership-diagram" {...rise}>
                    <div className="diagram-node diagram-node--site">
                        <Globe2 size={20} />
                        <span>Your website</span>
                    </div>
                    <div className="diagram-line"><i /></div>
                    <div className="diagram-core">
                        <IrisMark />
                        <strong>Your Iris</strong>
                        <span>Your server · Your database</span>
                    </div>
                    <div className="diagram-line"><i /></div>
                    <div className="diagram-node">
                        <Database size={20} />
                        <span>SQLite</span>
                    </div>
                    <div className="diagram-shield"><LockKeyhole size={17} /> Nothing leaves your stack</div>
                </motion.div>
            </div>

            <div className="section highlight-grid">
                {highlights.map((item, index) => {
                    const Icon = item.icon;
                    return (
                        <motion.article
                            key={item.title}
                            className="highlight-card"
                            {...rise}
                            transition={{ ...rise.transition, delay: index * 0.08 }}
                        >
                            <div className="highlight-meta">
                                <span>{item.number}</span>
                                <Icon size={20} />
                            </div>
                            <h3>{item.title}</h3>
                            <p>{item.text}</p>
                        </motion.article>
                    );
                })}
            </div>
        </section>
    );
}

function DevelopersSection() {
    const [copied, setCopied] = useState(false);
    const command = "npm install iris-analytics";

    const handleCopy = async () => {
        await navigator.clipboard.writeText(command);
        setCopied(true);
        window.setTimeout(() => setCopied(false), 1600);
    };

    return (
        <section id="developers" className="section developers-section">
            <motion.div className="developer-copy" {...rise}>
                <SectionLabel>04 / Built for developers</SectionLabel>
                <h2>From zero to signal<br />in a few lines.</h2>
                <p>
                    Add the tiny TypeScript SDK, point it at your Iris instance, and start.
                    Sensible defaults handle the rest.
                </p>
                <ul>
                    <li><Check size={16} /> TypeScript-native SDK</li>
                    <li><Check size={16} /> Automatic pageviews and Web Vitals</li>
                    <li><Check size={16} /> Works with React, Next.js, Remix, and SPAs</li>
                </ul>
                <Link to="/docs" className="button button--dark">
                    Read the docs <ArrowRight size={16} />
                </Link>
            </motion.div>

            <motion.div className="code-window" {...rise}>
                <div className="code-topbar">
                    <div><i /><i /><i /></div>
                    <span>app.ts</span>
                    <button type="button" onClick={handleCopy} aria-label="Copy install command">
                        {copied ? <Check size={16} /> : <Copy size={16} />}
                        {copied ? "Copied" : "Copy"}
                    </button>
                </div>
                <div className="code-body">
                    <div className="code-line"><span>1</span><code className="comment">// Install the SDK</code></div>
                    <div className="code-line"><span>2</span><code>npm install iris-analytics</code></div>
                    <div className="code-line"><span>3</span><code /></div>
                    <div className="code-line"><span>4</span><code className="comment">// Initialize Iris</code></div>
                    <div className="code-line"><span>5</span><code><b>import</b> &#123; Iris &#125; <b>from</b> <q>iris-analytics</q>;</code></div>
                    <div className="code-line"><span>6</span><code /></div>
                    <div className="code-line"><span>7</span><code><b>const</b> iris = <b>new</b> Iris(&#123;</code></div>
                    <div className="code-line"><span>8</span><code>&nbsp;&nbsp;host: <q>https://analytics.yoursite.com</q>,</code></div>
                    <div className="code-line"><span>9</span><code>&nbsp;&nbsp;siteId: <q>my-site</q></code></div>
                    <div className="code-line"><span>10</span><code>&#125;);</code></div>
                    <div className="code-line"><span>11</span><code /></div>
                    <div className="code-line"><span>12</span><code>iris.start();</code></div>
                </div>
                <div className="code-status">
                    <span><i /> Connected</span>
                    <span>event accepted · 204</span>
                </div>
            </motion.div>
        </section>
    );
}

export function Home() {
    return (
        <main className="home-page">
            <section className="hero-section">
                <div className="hero-grid">
                    <motion.div
                        className="hero-copy"
                        initial={{ opacity: 0, y: 24 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ duration: 0.65, ease: [0.22, 1, 0.36, 1] }}
                    >
                        <div className="eyebrow">
                            <span className="status-dot" />
                            Open-source analytics
                            <span className="eyebrow-divider" />
                            v1.0.0
                        </div>
                        <h1>
                            Your traffic,
                            <span>beautifully clear.</span>
                        </h1>
                        <p>
                            A calmer, clearer way to understand what people use, how your site
                            performs, and where growth begins—without giving away your data.
                        </p>
                        <div className="hero-footnote">
                            <span><i /> Cookie-free</span>
                            <span><i /> Self-hosted</span>
                        </div>
                    </motion.div>

                    <AnalyticsPreview />
                </div>

                <div className="proof-strip">
                    <div className="proof-item">
                        <span>01 / Clarity</span>
                        <strong>One calm view of everything that matters.</strong>
                    </div>
                    <div className="proof-item">
                        <span>02 / Privacy</span>
                        <strong>Your data stays yours.</strong>
                    </div>
                    <div className="proof-item">
                        <span>03 / Speed</span>
                        <strong>Built off the critical path.</strong>
                    </div>
                </div>
            </section>

            <ProductSection />
            <SignalStorySection />
            <PrivacySection />
            <DevelopersSection />

            <section className="section final-cta">
                <motion.div {...rise}>
                    <IrisMark />
                    <SectionLabel>Start seeing clearly</SectionLabel>
                    <h2>Good analytics ask less<br />and tell you more.</h2>
                    <p>Deploy Iris on your infrastructure and make your first event count.</p>
                    <div>
                        <Link to="/docs" className="button button--accent">
                            Get started <ArrowRight size={17} />
                        </Link>
                        <a href="https://github.com/VatsalP117/iris" className="button button--ghost">
                            <GitHubIcon size={17} /> Star on GitHub
                        </a>
                    </div>
                </motion.div>
            </section>
        </main>
    );
}

function Footer() {
    return (
        <footer className="site-footer">
            <div className="footer-main">
                <div>
                    <Link to="/" className="brand">
                        <IrisMark />
                        <span>Iris</span>
                    </Link>
                    <p>Private, self-hosted web analytics<br />for teams who own their stack.</p>
                </div>
                <div className="footer-links">
                    <div>
                        <strong>Product</strong>
                        <a href="/#product">Overview</a>
                        <a href="/#privacy">Privacy</a>
                        <a href="/#developers">Developers</a>
                    </div>
                    <div>
                        <strong>Resources</strong>
                        <Link to="/docs">Documentation</Link>
                        <a href="https://www.npmjs.com/package/iris-analytics">npm</a>
                        <a href="https://github.com/VatsalP117/iris">GitHub</a>
                    </div>
                </div>
            </div>
            <div className="footer-bottom">
                <span>© 2026 Iris Analytics</span>
                <span>Open source. Built with intention.</span>
            </div>
        </footer>
    );
}

export default function App() {
    return (
        <div className="min-h-screen">
            <Nav />
            <Outlet />
            <Footer />
        </div>
    );
}
