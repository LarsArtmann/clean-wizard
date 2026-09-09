import type { StepCard, ComparisonItem, UseCase, ComparisonMatrix, PainPoint } from "./types";

export const painPoints: PainPoint[] = [
  { tool: "Go", detail: "Build cache + module cache + test cache", size: "11 GiB" },
  { tool: "Docker", detail: "Stopped containers, dangling images, volumes", size: "5+ GiB" },
  { tool: "Nix", detail: "Old generations from every system update", size: "10+ GiB" },
  { tool: "Homebrew", detail: "Download cache, dead symlinks, old casks", size: "1+ GiB" },
  { tool: "Node", detail: "npm, pnpm, yarn, bun caches combined", size: "2+ GiB" },
  { tool: "Cargo", detail: "Registry cache + git cache + old artifacts", size: "1.6 GiB" },
];

export const steps: StepCard[] = [
  {
    step: "1",
    stepColor: "accent",
    title: "Scan",
    desc: "Detects all available cleaners based on what is installed on your system. Nothing runs yet, just discovery.",
    code: "$ clean-wizard scan",
  },
  {
    step: "2",
    stepColor: "accent",
    title: "Select",
    desc: "Pick cleaners interactively or use a preset mode: quick, standard, or aggressive. Each shows its estimated size.",
    code: "$ clean-wizard clean --mode quick",
  },
  {
    step: "3",
    stepColor: "amber",
    title: "Confirm",
    desc: "Dry-run shows exactly what gets removed and how much space comes back. Nothing is deleted without your explicit yes.",
    code: "$ clean-wizard clean --dry-run",
  },
  {
    step: "4",
    stepColor: "amber",
    title: "Clean",
    desc: "All selected cleaners run in parallel. Results reported per-cleaner with freed bytes. Failed cleaners retry automatically.",
    code: "Freed 12 GiB in 1.1s",
  },
];

export const comparisons: ComparisonItem[] = [
  {
    variant: "Manual Commands",
    accent: false,
    pros: ["No installation needed", "Full control over every command"],
    cons: [
      "Must remember 13+ different cleanup commands",
      "No dry-run or preview before deleting",
      "Easy to delete something important",
      "Sequential, no parallel execution",
    ],
  },
  {
    variant: "Generic Cleaners",
    accent: false,
    pros: ["Broad system coverage", "GUI available", "Well-established tools"],
    cons: [
      "Not developer-focused",
      "Misses dev-specific caches like Nix generations",
      "Limited JSON output for automation",
      "No retry or error classification",
    ],
  },
  {
    variant: "Clean Wizard",
    accent: true,
    pros: [
      "13 dev caches in one command",
      "Dry-run previews before any change",
      "Protected generations and paths",
      "Parallel execution with auto-retry",
      "JSON output for CI/CD pipelines",
    ],
    cons: [],
  },
];

export const comparisonMatrix: ComparisonMatrix = {
  columns: ["Manual Commands", "Generic Cleaners", "Clean Wizard"],
  rows: [
    { feature: "Dev-cache aware", values: ["no", "partial", "yes"] },
    { feature: "Dry-run preview", values: ["partial", "partial", "yes"] },
    { feature: "Interactive TUI", values: ["no", "partial", "yes"] },
    { feature: "JSON for CI/CD", values: ["no", "no", "yes"] },
    { feature: "Parallel execution", values: ["no", "partial", "yes"] },
    { feature: "Auto-retry on failure", values: ["no", "no", "yes"] },
    { feature: "Nix generation protection", values: ["no", "no", "yes"] },
    { feature: "Safe by default", values: ["no", "partial", "yes"] },
  ],
};

export const useCases: UseCase[] = [
  {
    title: "Developer Workstations",
    desc: "Reclaim 5-20 GB from Go, Node, Docker, and Homebrew caches in one daily command",
    icon: "dev",
  },
  {
    title: "NixOS Servers",
    desc: "Clean old generations safely without touching the current one",
    icon: "server",
  },
  {
    title: "CI/CD Pipelines",
    desc: "JSON output and dry-run mode make it safe to automate in any pipeline",
    icon: "ci",
  },
];
