import { motion } from 'framer-motion';
import {
  ShieldCheck,
  Container,
  Zap,
  Database,
  ArrowRight,
  MousePointer2,
  Lock,
  Menu,
  X,
  ChevronRight
} from 'lucide-react';

const GitHubIcon = ({ className }: { className?: string }) => (
  <svg viewBox="0 0 24 24" fill="currentColor" className={className}>
    <path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0 0 24 12c0-6.63-5.37-12-12-12z" />
  </svg>
);
import { useState } from 'react';
import { Link, Outlet, useLocation } from 'react-router-dom';
// test comment to trigger deploy
const Nav = () => {
  const [isOpen, setIsOpen] = useState(false);
  const location = useLocation();
  const isDocsPage = location.pathname === '/docs';

  return (
    <nav className="fixed top-0 w-full z-50 border-b border-subtle bg-background/80 backdrop-blur-sm">
      <div className="max-w-6xl mx-auto px-6">
        <div className="flex justify-between h-16 items-center">
          <Link to="/" className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-lg bg-foreground text-background flex items-center justify-center font-bold text-sm">
              IR
            </div>
            <span className="text-base font-medium tracking-tight">Iris</span>
          </Link>

          <div className="hidden md:flex items-center gap-8 text-sm">
            {!isDocsPage && (
              <a href="#features" className="text-muted hover:text-foreground transition-colors">
                Features
              </a>
            )}
            <Link to="/docs" className="text-muted hover:text-foreground transition-colors">
              Docs
            </Link>
            <a
              href="https://github.com/VatsalP117/iris"
              className="text-muted hover:text-foreground transition-colors flex items-center gap-2"
            >
              <GitHubIcon className="w-4 h-4" />
              GitHub
            </a>
            <Link
              to="/docs"
              className="px-4 py-2 rounded-lg bg-foreground text-background font-medium text-sm hover:bg-white/90 transition-colors"
            >
              Get Started
            </Link>
          </div>

          <button
            onClick={() => setIsOpen(!isOpen)}
            className="md:hidden p-2 text-muted hover:text-foreground"
          >
            {isOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
          </button>
        </div>
      </div>

      {isOpen && (
        <div className="md:hidden border-t border-subtle bg-background px-6 py-4 space-y-3">
          {!isDocsPage && (
            <a href="#features" className="block text-muted py-2" onClick={() => setIsOpen(false)}>
              Features
            </a>
          )}
          <Link to="/docs" className="block text-muted py-2" onClick={() => setIsOpen(false)}>
            Docs
          </Link>
          <a href="https://github.com/VatsalP117/iris" className="block text-muted py-2" onClick={() => setIsOpen(false)}>
            GitHub
          </a>
          <Link
            to="/docs"
            className="block w-full px-4 py-2.5 rounded-lg bg-foreground text-background font-medium text-center"
            onClick={() => setIsOpen(false)}
          >
            Get Started
          </Link>
        </div>
      )}
    </nav>
  );
};

const FeatureCard = ({
  icon: Icon,
  title,
  description,
  delay = 0
}: {
  icon: any;
  title: string;
  description: string;
  delay?: number
}) => (
  <motion.div
    initial={{ opacity: 0, y: 20 }}
    whileInView={{ opacity: 1, y: 0 }}
    viewport={{ once: true }}
    transition={{ duration: 0.5, delay }}
    className="p-6 rounded-xl border border-subtle glass-hover transition-colors group"
  >
    <div className="w-10 h-10 rounded-lg bg-foreground/5 flex items-center justify-center mb-4 group-hover:bg-foreground/10 transition-colors">
      <Icon className="w-5 h-5 text-muted-foreground" />
    </div>
    <h3 className="text-base font-medium mb-2">{title}</h3>
    <p className="text-sm text-muted leading-relaxed">{description}</p>
  </motion.div>
);

export function Home() {
  return (
    <main className="pt-32 pb-20">
      {/* Hero */}
      <section className="max-w-6xl mx-auto px-6 text-center py-20">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6 }}
        >
          <span className="inline-flex items-center gap-2 px-3 py-1 rounded-full text-xs font-medium bg-foreground/5 text-muted-foreground border border-subtle mb-8">
            <span className="w-1.5 h-1.5 rounded-full bg-foreground" />
            v0.1.0 — Public Alpha
          </span>

          <h1 className="text-5xl md:text-7xl font-semibold tracking-tight mb-6">
            Own your analytics.
          </h1>
          <h1 className="text-5xl md:text-7xl font-semibold tracking-tight text-muted mb-8">
            Keep your privacy.
          </h1>

          <p className="text-lg text-muted max-w-xl mx-auto mb-10 leading-relaxed">
            The lightweight, self-hosted web analytics platform built for developers.
            Single Docker container. 100% data ownership. Zero cookies.
          </p>

          <div className="flex flex-col sm:flex-row items-center justify-center gap-3">
            <Link
              to="/docs"
              className="w-full sm:w-auto px-6 py-3 rounded-lg bg-foreground text-background font-medium text-sm hover:bg-white/90 transition-colors inline-flex items-center justify-center gap-2"
            >
              Deploy in 5 minutes
              <ArrowRight className="w-4 h-4" />
            </Link>
            <a
              href="https://github.com/VatsalP117/iris"
              className="w-full sm:w-auto px-6 py-3 rounded-lg border border-subtle text-sm font-medium hover:bg-foreground/5 transition-colors inline-flex items-center justify-center gap-2"
            >
              <GitHubIcon className="w-4 h-4" />
              View on GitHub
            </a>
          </div>
        </motion.div>

        {/* Terminal */}
        <motion.div
          initial={{ opacity: 0, y: 40 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.3, duration: 0.8 }}
          className="mt-20 max-w-2xl mx-auto"
        >
          <div className="rounded-xl border border-subtle overflow-hidden">
            <div className="px-4 py-2.5 border-b border-subtle flex items-center gap-2">
              <div className="w-2.5 h-2.5 rounded-full bg-foreground/20" />
              <div className="w-2.5 h-2.5 rounded-full bg-foreground/20" />
              <div className="w-2.5 h-2.5 rounded-full bg-foreground/20" />
              <span className="text-xs text-muted ml-2 font-mono">iris-deploy.sh</span>
            </div>
            <div className="p-5 text-left font-mono text-sm bg-bg-secondary">
              <p><span className="text-muted">$</span> docker run -d -p 8080:8080 \</p>
              <p className="pl-4">-v ./data:/app/data \</p>
              <p className="pl-4">-e SITE_ID="my-site" \</p>
              <p className="pl-4">ghcr.io/iris/iris:latest</p>
              <br />
              <p className="text-muted">✓ Server listening on :8080</p>
              <p className="text-muted">✓ Database initialized</p>
              <p className="text-muted">✓ Dashboard ready</p>
            </div>
          </div>
        </motion.div>
      </section>

      {/* Features */}
      <section id="features" className="max-w-6xl mx-auto px-6 py-24">
        <div className="text-center mb-12">
          <h2 className="text-2xl md:text-3xl font-semibold tracking-tight mb-4">
            Everything you need
          </h2>
          <p className="text-muted max-w-lg mx-auto">
            Track your visitors without compromising their privacy or your control.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          <FeatureCard
            icon={ShieldCheck}
            title="Privacy-First"
            description="Fully anonymous, cookieless tracking. No GDPR consent banners required."
            delay={0}
          />
          <FeatureCard
            icon={Container}
            title="Single Binary"
            description="Go backend, SQLite, and dashboard in one Docker image. No complex setup."
            delay={0.1}
          />
          <FeatureCard
            icon={Zap}
            title="Lightning Fast"
            description="Built with Go. Uses navigator.sendBeacon for zero performance impact."
            delay={0.2}
          />
          <FeatureCard
            icon={Database}
            title="Own Your Data"
            description="Your data never leaves your server. No third-party scripts or tracking."
            delay={0.3}
          />
          <FeatureCard
            icon={MousePointer2}
            title="Auto-capture"
            description="Automatically track clicks and pageviews without manual event handlers."
            delay={0.4}
          />
          <FeatureCard
            icon={Lock}
            title="Self-Hosted"
            description="Full control over your infrastructure. Deploy anywhere in seconds."
            delay={0.5}
          />
        </div>
      </section>

      {/* Integration */}
      <section className="max-w-6xl mx-auto px-6 py-24 border-t border-subtle">
        <div className="flex flex-col lg:flex-row items-center gap-12">
          <div className="flex-1">
            <h2 className="text-2xl md:text-3xl font-semibold mb-4 tracking-tight">
              Integrate in seconds
            </h2>
            <p className="text-muted leading-relaxed mb-6">
              Add a script tag or use our npm package. Everything is optimized for modern web development.
            </p>
            <ul className="space-y-3">
              {[
                "No configuration required",
                "TypeScript SDK",
                "Supports Next.js, Remix, SPAs",
                "Automatic Web Vitals tracking"
              ].map((item, i) => (
                <li key={i} className="flex items-center gap-3 text-sm">
                  <ChevronRight className="w-4 h-4 text-muted" />
                  <span className="text-muted-foreground">{item}</span>
                </li>
              ))}
            </ul>
          </div>

          <div className="flex-1 w-full">
            <div className="rounded-xl border border-subtle overflow-hidden">
              <div className="px-4 py-2.5 border-b border-subtle flex items-center gap-2">
                <span className="text-xs font-medium text-muted-foreground">SDK</span>
                <span className="text-xs text-muted font-mono">index.ts</span>
              </div>
              <div className="p-5 font-mono text-sm bg-bg-secondary">
                <p className="text-muted">// Install</p>
                <p className="mb-3"><span className="text-foreground">import</span> &#123; Iris &#125; <span className="text-foreground">from</span> <span className="text-muted">'iris-analytics'</span>;</p>

                <p className="text-muted">// Initialize</p>
                <p className="text-foreground">const <span className="text-foreground">iris</span> = new <span className="text-foreground">Iris</span>(&#123;</p>
                <p className="pl-4">host: <span className="text-muted">'https://your-iris.com'</span>,</p>
                <p className="pl-4">siteId: <span className="text-muted">'my-site'</span></p>
                <p className="text-foreground">&#125;);</p>
                <br />
                <p className="text-foreground">iris.<span className="text-foreground">start</span>();</p>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="max-w-4xl mx-auto px-6 py-24 text-center">
        <div className="p-12 rounded-2xl border border-subtle">
          <h2 className="text-2xl md:text-3xl font-semibold mb-4">
            Ready to own your data?
          </h2>
          <p className="text-muted mb-8 max-w-md mx-auto">
            Join developers who have switched to a more private, faster analytics tool.
          </p>
          <div className="flex flex-col sm:flex-row items-center justify-center gap-3">
            <Link
              to="/docs"
              className="px-6 py-3 rounded-lg bg-foreground text-background font-medium text-sm hover:bg-white/90 transition-colors"
            >
              Get Started
            </Link>
            <Link
              to="/docs"
              className="px-6 py-3 rounded-lg border border-subtle text-sm font-medium hover:bg-foreground/5 transition-colors"
            >
              Read the Docs
            </Link>
          </div>
        </div>
      </section>
    </main>
  );
}

export default function App() {
  return (
    <div className="min-h-screen bg-background">
      <div className="fixed inset-0 bg-grid pointer-events-none" />

      <Nav />
      <Outlet />

      <footer className="max-w-6xl mx-auto px-6 py-12 border-t border-subtle">
        <div className="flex flex-col md:flex-row justify-between items-center gap-6">
          <Link to="/" className="flex items-center gap-2 text-sm font-medium">
            <div className="w-5 h-5 rounded bg-foreground text-background flex items-center justify-center text-[8px] font-bold">
              IR
            </div>
            Iris Analytics
          </Link>
          <div className="flex items-center gap-6 text-sm text-muted">
            <a href="https://github.com/VatsalP117/iris" className="hover:text-foreground transition-colors">
              GitHub
            </a>
            <a href="#" className="hover:text-foreground transition-colors">
              Discord
            </a>
          </div>
          <p className="text-xs text-muted">© 2026 Iris Analytics</p>
        </div>
      </footer>
    </div>
  );
}
