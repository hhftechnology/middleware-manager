import Link from 'next/link';

const featureHighlights = [
  {
    title: 'Unified Control Plane',
    body: 'Manage Traefik middlewares, routers, services, plugins, and mTLS from a single UI.',
  },
  {
    title: 'Safe Overrides',
    body: 'Attach policies with priority, override services, and regenerate file-provider rules automatically.',
  },
  {
    title: 'Security First',
    body: 'mTLS via mtlswhitelist, access logging, and plugin hygiene—built in for zero-trust edge setups.',
  },
];

const quickLinks = [
  { href: '/docs/getting-started/onboarding', label: 'Onboarding' },
  { href: '/docs/getting-started/deploy-pangolin', label: 'Deploy with Pangolin' },
  { href: '/docs/getting-started/deploy-standalone', label: 'Deploy with Traefik' },
  { href: '/docs/ui-guides/resources', label: 'Resources & Routers' },
  { href: '/docs/ui-guides/middlewares', label: 'Middlewares' },
  { href: '/docs/ui-guides/plugin-hub', label: 'Plugin Hub' },
  { href: '/docs/security/risks', label: 'Security & Risks' },
];

export default function HomePage() {
  return (
    <div
      className="relative min-h-screen flex flex-col"
      style={{ backgroundColor: '#050505' }}
    >
      <div className="absolute inset-0 bg-gradient-to-b from-white/5 via-transparent to-transparent pointer-events-none" />

      <main className="relative flex-1 w-full px-6 py-16 md:py-20 lg:py-24">
        <div className="max-w-6xl mx-auto flex flex-col gap-16">
          <section className="grid gap-8 lg:grid-cols-[1.2fr_0.8fr] items-center">
            <div className="space-y-6">
              <p className="text-sm uppercase tracking-[0.28em] text-white/40">
                Middleware Manager · Traefik & Pangolin
              </p>
              <h1 className="text-4xl md:text-5xl lg:text-6xl font-semibold tracking-tight text-white/90 leading-tight">
                Ship safer edge traffic with a single source of truth for Traefik.
              </h1>
              <p className="text-lg text-white/60 max-w-2xl">
                Discover resources, apply middleware chains with priorities, install plugins, and enforce mTLS—all without touching raw YAML. Built for operators who need confidence and clarity at the edge.
              </p>
              <div className="flex flex-wrap items-center gap-4">
                <Link
                  href="/docs/getting-started/onboarding"
                  className="px-4 py-2.5 rounded-full bg-white text-black font-medium hover:bg-white/90 transition"
                >
                  Start the Guide
                </Link>
                <Link
                  href="/docs"
                  className="px-4 py-2.5 rounded-full border border-white/15 text-white/80 hover:border-white/40 hover:text-white transition"
                >
                  Browse Docs
                </Link>
                <Link
                  href="https://github.com/hhftechnology/middleware-manager"
                  className="px-4 py-2.5 rounded-full border border-white/15 text-white/80 hover:border-white/40 hover:text-white transition"
                >
                  GitHub
                </Link>
                <Link
                  href="https://discord.gg/PEGcTJPfJ2"
                  className="px-4 py-2.5 rounded-full border border-white/15 text-white/80 hover:border-white/40 hover:text-white transition"
                >
                  Discord
                </Link>
              </div>
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 pt-4 text-sm text-white/60">
                <div className="rounded-2xl border border-white/10 bg-white/5 p-4">
                  <p className="text-white/90 font-semibold">Dual Data Sources</p>
                  <p>Pangolin or Traefik as the source of truth.</p>
                </div>
                <div className="rounded-2xl border border-white/10 bg-white/5 p-4">
                  <p className="text-white/90 font-semibold">File Provider Output</p>
                  <p>Regenerated rules with cache invalidation hooks.</p>
                </div>
                <div className="rounded-2xl border border-white/10 bg-white/5 p-4">
                  <p className="text-white/90 font-semibold">Security Built-In</p>
                  <p>mTLS, plugin hygiene, and access log guidance.</p>
                </div>
              </div>
            </div>

            <div className="rounded-3xl border border-white/10 bg-white/5 overflow-hidden">
              <div className="rounded-3xl border border-white/10 bg-white/[0.04] p-6 space-y-5">
                <p className="text-sm font-semibold text-white/70">Quick paths</p>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  {quickLinks.map((link) => (
                    <Link
                      key={link.href}
                      href={link.href}
                      className="group rounded-2xl border border-white/10 bg-white/[0.02] px-4 py-3 text-white/70 hover:text-white hover:border-white/30 transition"
                    >
                      <span className="text-sm font-medium">{link.label}</span>
                      <span className="block text-xs text-white/40 group-hover:text-white/60">
                        {link.href}
                      </span>
                    </Link>
                  ))}
                </div>
              </div>
            </div>
          </section>

          <section className="grid gap-6 md:grid-cols-3">
            {featureHighlights.map((item) => (
              <div
                key={item.title}
                className="rounded-2xl border border-white/10 bg-white/[0.02] p-6 hover:border-white/25 transition"
              >
                <p className="text-lg font-semibold text-white/90">{item.title}</p>
                <p className="mt-3 text-white/60">{item.body}</p>
              </div>
            ))}
          </section>
        </div>
      </main>
    </div>
  );
}
