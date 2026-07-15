export const heroCode = `$ clean-wizard scan

Scanning system for cleanable items...

Found 11 available cleaner(s)

  System Cache          9 items    20 GiB
  Compiled Binaries     162 items   5.1 GiB
  Go Packages           2 items    11 GiB
  Cargo Packages        2 items     1.6 GiB
  golangci-lint Cache   1 item      2.2 MiB
  Project Executables   1 item      100 MiB

Total cleanable: 38 GiB (177 items)

$ clean-wizard clean --mode quick --dry-run

Found 11 available cleaner(s)
Using preset mode: quick
  Homebrew
  Node.js Packages
  Go Packages
  Temp Files
  Build Cache

Starting cleanup... (DRY RUN)

  buildcache    3 items    900 MiB
  go            2 items    11 GiB

Total: 12 GiB freed, 6 items in 1.1s

$ clean-wizard clean --mode quick

Cleanup Results

  buildcache    3 items    900 MiB
  go            2 items    11 GiB

Total: 12 GiB freed in 1.1s`;
