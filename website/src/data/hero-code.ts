export const heroCode = `$ clean-wizard clean --mode quick --dry-run

Scanning for available cleaners...
  Found 7 available cleaners

Would clean:
  Homebrew      cache + autoremove     ~245 MB
  Go            build + mod + test     ~1.2 GB
  Node          npm + pnpm + yarn      ~380 MB
  BuildCache    Gradle + Maven         ~890 MB
  TempFiles     older than 7 days      ~156 MB

Estimated total: ~2.9 GB

$ clean-wizard clean --mode quick

Running 5 cleaners in parallel...
  Homebrew      245 MB
  Go          1,204 MB
  Node          380 MB
  BuildCache    890 MB
  TempFiles     156 MB

Freed 2.9 GB in 3.1s`;
