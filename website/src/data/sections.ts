import type { StepCard, ComparisonItem, UseCase, ComparisonMatrix } from "./types";

export const steps: StepCard[] = [
  {
    step: "1",
    stepColor: "accent",
    title: "Scan",
    desc: "Detects all available cleaners based on installed tools and system platform.",
    code: "$ clean-wizard scan",
  },
  {
    step: "2",
    stepColor: "accent",
    title: "Select",
    desc: "Interactive TUI or preset mode (quick, standard, aggressive) to choose what to clean.",
    code: "$ clean-wizard clean --mode quick",
  },
  {
    step: "3",
    stepColor: "amber",
    title: "Confirm",
    desc: "Dry-run preview shows what will be removed. Explicit yes/no before any deletion.",
    code: "$ clean-wizard clean --dry-run",
  },
  {
    step: "4",
    stepColor: "amber",
    title: "Clean",
    desc: "Parallel execution with retry support. Results reported per-cleaner with freed bytes.",
    code: "// Freed 5.2 GB in 3.1s",
  },
];

export const comparisons: ComparisonItem[] = [
  {
    variant: "Manual Cleanup",
    accent: false,
    pros: ["No installation needed"],
    cons: [
      "Remember every cache location",
      "No dry-run or preview",
      "Easy to delete something important",
      "No parallel execution",
    ],
  },
  {
    variant: "Clean Wizard",
    accent: true,
    pros: [
      "13 cleaners in one command",
      "Dry-run previews before any change",
      "Protected generations and paths",
      "Parallel execution with retries",
      "JSON output for CI/CD",
    ],
    cons: [],
  },
  {
    variant: "SystemNix",
    accent: false,
    pros: ["Simple shell script"],
    cons: [
      "No dry-run mode",
      "No TUI or interactive selection",
      "No JSON output",
      "Hardcoded configuration",
      "No retry on failure",
    ],
  },
];

export const comparisonMatrix: ComparisonMatrix = {
  columns: ["Manual Cleanup", "SystemNix", "Clean Wizard"],
  rows: [
    { feature: "Dry-run preview", values: ["no", "no", "yes"] },
    { feature: "Interactive TUI", values: ["no", "no", "yes"] },
    { feature: "JSON output", values: ["no", "no", "yes"] },
    { feature: "Parallel execution", values: ["no", "no", "yes"] },
    { feature: "Retry on failure", values: ["no", "no", "yes"] },
    { feature: "Error classification", values: ["no", "no", "yes"] },
    { feature: "Configuration profiles", values: ["no", "no", "yes"] },
    { feature: "Cleaners", values: ["manual", "8", "13"] },
  ],
};

export const useCases: UseCase[] = [
  {
    title: "Developer Workstations",
    desc: "Daily cleanup of Go, Node, Docker, and Homebrew caches",
    icon: "dev",
  },
  {
    title: "NixOS Servers",
    desc: "Old generation cleanup with protected current generation",
    icon: "server",
  },
  {
    title: "CI/CD Pipelines",
    desc: "JSON output and dry-run for safe automation",
    icon: "ci",
  },
];
