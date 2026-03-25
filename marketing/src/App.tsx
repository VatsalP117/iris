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
  X
} from 'lucide-react';

const GitHubIcon = ({ className }: { className?: string }) => (
  <svg viewBox="0 0 24 24" fill="currentColor" className={className}>
    <path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0 0 24 12c0-6.63-5.37-12-12-12z" />
  </svg>
);
import { useState } from 'react';

const Nav = () => {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <nav className="fixed top-0 w-full z-50 border-b border-white/5 bg-background/80 backdrop-blur-xl">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between h-16 items-center">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-primary to-accent flex items-center justify-center font-bold text-white shadow-lg shadow-primary/20">
              IR
            </div>
            <span className="text-xl font-bold tracking-tight">Iris</span>
          </div>

          <div className="hidden md:flex items-center space-x-8 text-sm font-medium text-muted-foreground">
            <a href="#features" className="hover:text-foreground transition-colors">Features</a>
            <a href="#integration" className="hover:text-foreground transition-colors">Docs</a>
            <a href="https://github.com" className="hover:text-foreground transition-colors flex items-center gap-1.5">
              <GitHubIcon className="w-4 h-4" /> GitHub
            </a>
            <button className="px-4 py-2 rounded-full bg-white text-background font-semibold hover:bg-white/90 transition-all">
              Get Started
            </button>
          </div>

          <div className="md:hidden">
            <button onClick={() => setIsOpen(!isOpen)} className="p-2">
              {isOpen ? <X /> : <Menu />}
            </button>
          </div>
        </div>
      </div>

      {isOpen && (
        <div className="md:hidden bg-background border-b border-white/5 p-4 space-y-4">
          <a href="#features" className="block text-muted-foreground px-2 py-1">Features</a>
          <a href="#integration" className="block text-muted-foreground px-2 py-1">Docs</a>
          <a href="#" className="block text-muted-foreground px-2 py-1">GitHub</a>
          <button className="w-full px-4 py-2 rounded-lg bg-white text-background font-semibold">Get Started</button>
        </div>
      )}
    </nav>
  );
};

const FeatureCard = ({ icon: Icon, title, description }: { icon: any, title: string, description: string }) => (
  <motion.div
    whileHover={{ y: -5 }}
    className="p-8 rounded-3xl glass relative overflow-hidden group"
  >
    <div className="absolute top-0 right-0 p-8 opacity-10 group-hover:opacity-20 transition-opacity">
      <Icon className="w-24 h-24" />
    </div>
    <div className="w-12 h-12 rounded-2xl bg-white/5 flex items-center justify-center mb-6 border border-white/10">
      <Icon className="w-6 h-6 text-primary" />
    </div>
    <h3 className="text-xl font-bold mb-3">{title}</h3>
    <p className="text-muted-foreground leading-relaxed">{description}</p>
  </motion.div>
);

export default function App() {
  return (
    <div className="min-h-screen bg-background selection:bg-primary/30">
      <div className="fixed inset-0 bg-grid pointer-events-none opacity-20" />
      <div className="fixed inset-0 bg-gradient-radial from-primary/5 via-transparent to-transparent pointer-events-none" />

      <Nav />

      <main className="relative pt-32">
        {/* Hero Section */}
        <section className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center py-20">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
          >
            <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-primary/10 text-primary border border-primary/20 mb-8">
              v0.1.0 — Now in Public Alpha
            </span>
            <h1 className="text-6xl md:text-8xl font-extrabold tracking-tighter mb-8 text-gradient">
              Own Your Analytics. <br />
              <span className="text-white/40">Keep Your Privacy.</span>
            </h1>
            <p className="text-lg md:text-xl text-muted-foreground max-w-2xl mx-auto mb-12 leading-relaxed">
              The lightweight, self-hosted web analytics platform built for developers.
              Single Docker container. 100% data ownership. Zero cookies.
            </p>
            <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
              <button className="w-full sm:w-auto px-8 py-4 rounded-full bg-white text-black font-bold text-lg hover:scale-105 transition-all flex items-center justify-center gap-2">
                Deploy in 5 minutes <ArrowRight className="w-5 h-5" />
              </button>
              <button className="w-full sm:w-auto px-8 py-4 rounded-full bg-white/5 border border-white/10 text-white font-bold text-lg hover:bg-white/10 transition-all flex items-center justify-center gap-2">
                <GitHubIcon className="w-5 h-5" /> View on GitHub
              </button>
            </div>
          </motion.div>

          {/* Abstract Terminal/Visual */}
          <motion.div
            initial={{ opacity: 0, y: 40 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.3, duration: 0.8 }}
            className="mt-24 relative max-w-4xl mx-auto"
          >
            <div className="absolute -inset-4 bg-gradient-to-r from-primary/20 to-accent/20 blur-3xl opacity-30 animate-pulse" />
            <div className="relative glass rounded-2xl overflow-hidden border border-white/10 shadow-2xl">
              <div className="bg-white/5 px-4 py-3 flex items-center gap-1.5 border-b border-white/10">
                <div className="w-3 h-3 rounded-full bg-red-500/50" />
                <div className="w-3 h-3 rounded-full bg-yellow-500/50" />
                <div className="w-3 h-3 rounded-full bg-green-500/50" />
                <span className="text-[10px] text-muted-foreground font-mono ml-4">iris-deploy.sh</span>
              </div>
              <div className="p-8 text-left font-mono text-sm leading-relaxed overflow-x-auto whitespace-nowrap bg-black/40">
                <p className="text-primary">$ <span className="text-white">docker run -d -p 8080:8080 \</span></p>
                <p className="pl-4 text-white">  -v ./data:/app/data \</p>
                <p className="pl-4 text-white">  -e SITE_ID="my-cool-site" \</p>
                <p className="pl-4 text-white font-bold">  ghcr.io/iris/iris:latest</p>
                <br />
                <p className="text-green-400">✓ Server listening on :8080</p>
                <p className="text-green-400">✓ Database initialized (SQLite)</p>
                <p className="text-green-400">✓ Dashboard available at /dashboard</p>
              </div>
            </div>
          </motion.div>
        </section>

        {/* Features Section */}
        <section id="features" className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-32">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            <FeatureCard
              icon={ShieldCheck}
              title="Privacy-First"
              description="Fully anonymous, cookieless tracking out of the box. No GDPR consent banners required because you don't track personal data."
            />
            <FeatureCard
              icon={Container}
              title="Single Binary"
              description="A single Docker image contains the Go backend, SQLite database, and the analytics dashboard. No complex setup."
            />
            <FeatureCard
              icon={Zap}
              title="Lightning Fast"
              description="Built with Go and React. The SDK uses navigator.sendBeacon to ensure zero impact on your site's performance."
            />
            <FeatureCard
              icon={Database}
              title="Own Your Data"
              description="Your data never leaves your server. No third-party scripts, no cross-site tracking, and no selling of visitor info."
            />
            <FeatureCard
              icon={MousePointer2}
              title="Auto-capture"
              description="Automatically track clicks and pageviews without writing manual event handlers for every element."
            />
            <FeatureCard
              icon={Lock}
              title="Self-Hosted"
              description="Full control over your infrastructure. Deploy on any VPS, Raspberry Pi, or home server in seconds."
            />
          </div>
        </section>

        {/* Integration Section */}
        <section id="integration" className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-32 border-t border-white/5">
          <div className="flex flex-col lg:flex-row items-center gap-16">
            <div className="flex-1">
              <h2 className="text-4xl md:text-5xl font-bold mb-6 tracking-tight">Integrate in <span className="text-primary underline decoration-primary/30 underline-offset-8">seconds</span>.</h2>
              <p className="text-lg text-muted-foreground leading-relaxed mb-8">
                Installing Iris is as simple as adding a single script tag or using our lightweight npm package.
                Everything is optimized for the modern web.
              </p>
              <ul className="space-y-4">
                {[
                  "No configuration required",
                  "TypeScript first SDK",
                  "Supports Next.js, Remix, and SPAs",
                  "Automatic Web Vitals tracking"
                ].map((item, i) => (
                  <li key={i} className="flex items-center gap-3 font-medium">
                    <div className="w-5 h-5 rounded-full bg-green-500/10 border border-green-500/20 flex items-center justify-center">
                      <Zap className="w-3 h-3 text-green-500" />
                    </div>
                    {item}
                  </li>
                ))}
              </ul>
            </div>
            <div className="flex-1 w-full">
              <div className="glass rounded-3xl overflow-hidden border border-white/10 shadow-2xl">
                <div className="bg-white/5 p-4 flex items-center gap-2">
                  <div className="px-3 py-1 rounded-md bg-primary/20 text-primary text-[10px] font-bold uppercase tracking-wider">SDK</div>
                  <span className="text-[11px] text-muted-foreground font-mono">index.ts</span>
                </div>
                <div className="p-8 font-mono text-sm leading-relaxed overflow-x-auto bg-black/20">
                  <p className="text-muted-foreground italic">// Install iris-analytics</p>
                  <p className="mb-4"><span className="text-primary">import</span> &#123; Iris &#125; <span className="text-primary">from</span> <span className="text-green-400">'iris-analytics'</span>;</p>

                  <p className="text-muted-foreground italic">// Initialize</p>
                  <p className="text-white">const <span className="text-accent">iris</span> = new <span className="text-yellow-400">Iris</span>(&#123;</p>
                  <p className="pl-4">host: <span className="text-green-400">'https://your-iris.com'</span>,</p>
                  <p className="pl-4">siteId: <span className="text-green-400">'marketing-site'</span></p>
                  <p className="text-white">&#125;);</p>
                  <br />
                  <p className="text-accent">iris.<span className="text-white">start</span>();</p>
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* Call to Action */}
        <section className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-32 text-center">
          <div className="p-16 rounded-[3rem] bg-gradient-to-br from-primary/20 via-accent/10 to-transparent border border-white/10 relative overflow-hidden group">
            <div className="absolute top-0 left-0 w-full h-full bg-grid opacity-10" />
            <h2 className="text-4xl md:text-5xl font-bold mb-8 relative z-10">Ready to own your data?</h2>
            <p className="text-lg text-muted-foreground mb-12 max-w-xl mx-auto relative z-10 leading-relaxed">
              Join hundreds of developers who have switched to a more private, faster, and simpler analytics tool.
            </p>
            <div className="flex flex-col sm:flex-row items-center justify-center gap-4 relative z-10">
              <button className="w-full sm:w-auto px-10 py-5 rounded-full bg-white text-black font-bold text-xl hover:scale-105 transition-all">
                Get Started Now
              </button>
              <button className="w-full sm:w-auto px-10 py-5 rounded-full bg-white/5 border border-white/10 text-white font-bold text-xl hover:bg-white/10 transition-all">
                Read the Docs
              </button>
            </div>
          </div>
        </section>
      </main>

      <footer className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-20 border-t border-white/5 text-center text-muted-foreground text-sm">
        <div className="flex flex-col md:flex-row justify-between items-center gap-8">
          <div className="flex items-center gap-2 text-foreground font-bold">
            <div className="w-6 h-6 rounded bg-gradient-to-br from-primary to-accent flex items-center justify-center text-[10px] text-white">IR</div>
            Iris Analytics
          </div>
          <div className="flex items-center gap-8">
            <a href="#" className="hover:text-foreground">GitHub</a>
            <a href="#" className="hover:text-foreground">Discord</a>
            <a href="#" className="hover:text-foreground">Twitter</a>
          </div>
          <p>© 2026 Iris Analytics. Built for the privacy-conscious web.</p>
        </div>
      </footer>
    </div>
  );
}
