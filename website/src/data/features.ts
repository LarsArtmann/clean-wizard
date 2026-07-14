import type { Feature } from "./types";

export const features: Feature[] = [
  {
    icon: "broom",
    title: "13 Specialized Cleaners",
    desc: "Nix, Docker, Go, Homebrew, Cargo, Node, Gradle, Maven, Xcode, and more. Each auto-detects whether its target is installed.",
  },
  {
    icon: "eye",
    title: "Dry-Run Previews",
    desc: "See exactly what would be cleaned and how much space would be freed before a single byte is touched.",
  },
  {
    icon: "terminal",
    title: "Interactive TUI",
    desc: "Beautiful multi-select forms powered by Charm Bracelet. Pick cleaners, see sizes, confirm in one flow.",
  },
  {
    icon: "bolt",
    title: "Parallel Execution",
    desc: "DAG-based workflow engine runs cleaners concurrently with configurable max concurrency and retry support.",
  },
  {
    icon: "shield",
    title: "Safety First",
    desc: "Confirmation dialogs, protected Nix generations, availability detection. Unavailable tools are skipped, not crashed.",
  },
  {
    icon: "code",
    title: "Type-Safe Core",
    desc: "27 cache types as compile-time enums. Registry pattern, DI container, Result[T] error handling. 300+ tests.",
  },
];
