/**
 * Log git-filter-repo command with appropriate provider prefix
 */
export function logFilterRepoCommand(verbose: boolean, args: string[]): void {
  if (!verbose) return;

  const provider = DetectFilterRepoProvider();
  if (provider === FilterRepoNix) {
    console.log(`Running: nix run nixpkgs#git-filter-repo -- ${args.join(" ")}`);
  } else {
    console.log(`Running: git filter-repo ${args.join(" ")}`);
  }
}
