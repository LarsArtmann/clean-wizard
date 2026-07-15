import type { Feature } from "./types";

export const features: Feature[] = [
  {
    icon: "broom",
    title: "13 Dev Caches, One Command",
    desc: "Go build caches, Docker images, Nix generations, Homebrew downloads, node_modules, Cargo registries, Gradle artifacts, and 7 more. Each is detected automatically based on what you have installed.",
  },
  {
    icon: "eye",
    title: "Dry-Run Previews",
    desc: "See exactly what gets cleaned and how much space comes back before a single byte is touched. Run --dry-run on any command, always.",
  },
  {
    icon: "terminal",
    title: "Interactive Terminal UI",
    desc: "Pick what to clean from a beautiful interactive menu. See sizes, toggle cleaners on and off, and confirm, all without leaving your terminal.",
  },
  {
    icon: "bolt",
    title: "All Cleaners at Once",
    desc: "Every cleaner runs in parallel, not one by one. Set how many run at the same time, and failed cleaners retry automatically with exponential backoff.",
  },
  {
    icon: "shield",
    title: "Nothing Deleted Without Consent",
    desc: "Confirmation required before any deletion. Current Nix generation is always protected. Tools that are not installed are skipped, never crash your run.",
  },
  {
    icon: "code",
    title: "Built for Automation",
    desc: "Pipe results to jq, integrate with CI, schedule with cron. JSON output on every command makes Clean Wizard scriptable, predictable, and machine-readable.",
  },
];
