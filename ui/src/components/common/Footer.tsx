export function Footer() {
  const currentYear = new Date().getFullYear()

  return (
    <footer className="border-t bg-background">
      <div className="container mx-auto flex h-14 items-center justify-between px-4">
        <p className="text-sm text-muted-foreground">
          {currentYear} Middleware Manager by{' '}
          <a
            href="https://forum.hhf.technology"
            target="_blank"
            rel="noopener noreferrer"
            className="hover:text-foreground transition-colors"
          >
            HHF Technology
          </a>
          {' '}- MIT License
        </p>
        <div className="flex items-center gap-4">
          <a
            href="https://github.com/hhftechnology/middleware-manager"
            target="_blank"
            rel="noopener noreferrer"
            className="text-muted-foreground hover:text-foreground transition-colors"
            aria-label="GitHub Repository"
          >
            <svg className="h-5 w-5" viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" />
            </svg>
            <span className="sr-only">GitHub</span>
          </a>
          <a
            href="https://discord.gg/PEGcTJPfJ2"
            target="_blank"
            rel="noopener noreferrer"
            className="text-muted-foreground hover:text-foreground transition-colors"
            aria-label="Discord Server"
          >
            <svg className="h-5 w-5" viewBox="0 0 24 24" fill="currentColor">
              <path d="M20.317 4.37a19.791 19.791 0 0 0-4.885-1.515.074.074 0 0 0-.079.037c-.211.375-.444.864-.607 1.25a18.27 18.27 0 0 0-5.487 0c-.163-.386-.395-.875-.607-1.25a.077.077 0 0 0-.079-.037 19.736 19.736 0 0 0-4.778.755.07.07 0 0 0-.032.027C.533 9.046-.32 13.58.099 18.057a.082.082 0 0 0 .031.057 19.9 19.9 0 0 0 5.993 3.03.078.078 0 0 0 .084-.028 14.975 14.975 0 0 0 1.293-2.1.07.07 0 0 0-.038-.098 13.11 13.11 0 0 1-1.872-.892.072.072 0 0 1-.009-.119c.125-.093.25-.19.371-.287a.075.075 0 0 1 .078-.01c3.928 1.793 8.18 1.793 12.062 0a.075.075 0 0 1 .079.009c.12.098.246.195.371.288a.072.072 0 0 1-.01.119 12.901 12.901 0 0 1-1.873.892.07.07 0 0 0-.037.099 14.993 14.993 0 0 0 1.293 2.1.078.078 0 0 0 .084.028 19.963 19.963 0 0 0 6.002-3.03.079.079 0 0 0 .033-.057c.5-4.566-.838-8.934-3.551-12.66a.071.071 0 0 0-.031-.028zM8.02 15.33c-1.183 0-2.157-.965-2.157-2.156 0-1.193.931-2.157 2.157-2.157 1.226 0 2.157.964 2.157 2.157 0 1.19-.93 2.155-2.157 2.155zm7.975 0c-1.183 0-2.157-.965-2.157-2.156 0-1.193.931-2.157 2.157-2.157 1.226 0 2.157.964 2.157 2.157 0 1.19-.931 2.155-2.157 2.155z" />
            </svg>
            <span className="sr-only">Discord</span>
          </a>
        </div>
      </div>
    </footer>
  )
}
