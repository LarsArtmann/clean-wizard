{
  description = "go-finding — Code quality finding framework for Go";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs@{
      self,
      nixpkgs,
      flake-parts,
      treefmt-nix,
    }:
    let
      inherit (nixpkgs) lib;

      version = self.rev or self.dirtyRev or "dev";
      # go_1_27 vendor output (proxyVendor) differs from go_1_26's —
      # regenerated 2026-09-23 when the package build moved to go_1_27.
      vendorHash = "sha256-GP+M2qo/sGPKktmdQFy4FRpLfIMJI3oJYuSAzQxA8sI=";
      proxyVendor = true;

      goSrc = lib.fileset.toSource {
        root = ./.;
        fileset = lib.fileset.gitTracked ./.;
      };

      mkGoFinding =
        buildGoModule:
        buildGoModule {
          pname = "go-finding";
          inherit version vendorHash proxyVendor;
          src = goSrc;
          # Multi-module: build from cmd/go-finding module.
          # Remove go.work so buildGoModule uses replace directives (GOWORK=off).
          postPatch = "rm -f go.work";
          modRoot = "cmd/go-finding";
          subPackages = [ "." ];
          ldflags = [
            "-s"
            "-w"
          ];
          env = {
          };
          meta = {
            description = "Code quality finding framework for Go";
            homepage = "https://github.com/LarsArtmann/go-finding";
            license = lib.licenses.mit;
            maintainers = [
              {
                name = "Lars Artmann";
                github = "LarsArtmann";
              }
            ];
            mainProgram = "go-finding";
          };
        };
    in
    flake-parts.lib.mkFlake { inherit inputs; } {
      # Inline systems: nixpkgs 26.11 dropped x86_64-darwin, which github:nix-systems/default still lists.
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
      ];

      imports = [
        treefmt-nix.flakeModule
      ];

      perSystem =
        {
          config,
          pkgs,
          ...
        }:
        let
          # 2026-09-19: go-linter-sdk's go.mod floor is >= 1.27.1, which
          # nixpkgs go_1_26 (1.26.7) cannot satisfy (GOTOOLCHAIN=local).
          goPkg = pkgs.go_1_27;

          mkApp = name: _description: script: {
            type = "app";
            program = "${
              pkgs.writeShellApplication {
                inherit name;
                runtimeInputs = [
                  goPkg
                  pkgs.golangci-lint
                  pkgs.trash-cli
                ];
                text = script;
              }
            }/bin/${name}";
            meta = {
              description = "Unified data model and pipeline for static analysis tools";
              mainProgram = name;
              homepage = "https://github.com/larsartmann/go-finding";
              license = lib.licenses.mit;
              platforms = lib.platforms.unix;
              maintainers = [
                {
                  name = "Lars Artmann";
                  github = "LarsArtmann";
                }
              ];
            };
          };
        in
        {
          treefmt = {
            projectRootFile = "go.mod";
            programs = {
              gofumpt.enable = true;
              # goimports shells out to `go list`. With no go on PATH it
              # falls back to its build-time GOROOT (nixpkgs gotools =
              # go 1.26.7), which cannot satisfy go.mod's `go 1.27` floor
              # and tries a GOTOOLCHAIN download — fatal inside the offline
              # treefmt-check sandbox (and, outside it, silently downloads a
              # stray toolchain into the repo's HOME). Wrap it: go_1_27 on
              # PATH, toolchain switching forbidden.
              goimports = {
                enable = true;
                package = pkgs.writeShellScriptBin "goimports" ''
                  export PATH="${pkgs.go_1_27}/bin:$PATH"
                  export GOTOOLCHAIN=local
                  exec "${pkgs.gotools}/bin/goimports" "$@"
                '';
              };
              golines = {
                enable = true;
                maxLength = 120;
              };
              nixfmt.enable = true;
            };
          };

          # go.mod floor is 1.27 (ginkgo 2.32.2 + go-linter-sdk chain); the
          # nixpkgs default go (1.26.7, GOTOOLCHAIN=local) cannot satisfy it —
          # same reason devShells/apps pin goPkg = pkgs.go_1_27.
          packages.default = mkGoFinding (pkgs.buildGoModule.override { go = pkgs.go_1_27; });

          devShells.default = pkgs.mkShell {
            packages = [
              goPkg
              pkgs.golangci-lint
              pkgs.gofumpt
              pkgs.golines
              pkgs.gopls
              pkgs.gotools
              pkgs.trash-cli
              pkgs.dprint
              pkgs.actionlint
            ];

            env = {
            };

            shellHook = ''
              echo "go-finding dev shell — $(go version)"
              echo "Multi-module workspace active (go.work)"
              # Install the commit-time formatting gate (source of truth:
              # scripts/hooks/pre-commit). Re-links when the tracked script
              # changes, so the gate can never silently rot.
              if [ -d .git ] && [ -f scripts/hooks/pre-commit ]; then
                if ! cmp -s scripts/hooks/pre-commit .git/hooks/pre-commit 2>/dev/null; then
                  cp scripts/hooks/pre-commit .git/hooks/pre-commit
                  chmod +x .git/hooks/pre-commit
                  echo "pre-commit hook installed (scripts/hooks/pre-commit)"
                fi
              fi
            '';
          };

          devShells.ci = pkgs.mkShellNoCC {
            packages = [
              goPkg
              pkgs.golangci-lint
            ];

            env = {
            };
          };

          checks = {
            format = config.treefmt.build.check self;
            build = config.packages.default;
          };

          apps = {
            test = mkApp "test" "Run all tests" ''
              go test ./... -count=1 "$@"
            '';

            test-race = mkApp "test-race" "Run all tests with race detector" ''
              go test ./... -race -count=1 "$@"
            '';

            bench = mkApp "bench" "Run benchmarks" ''
              go test ./... -bench=. -benchmem "$@"
            '';

            build = mkApp "build" "Build all packages" ''
              go build ./...
            '';

            vet = mkApp "vet" "Run go vet" ''
              go vet ./...
            '';

            lint = mkApp "lint" "Run golangci-lint" ''
              golangci-lint run ./...
            '';

            coverage = mkApp "coverage" "Run tests with coverage report" ''
              go test ./... -coverprofile=coverage.out -covermode=atomic "$@"
              go tool cover -func=coverage.out
            '';

            art-dupl = mkApp "art-dupl" "Check code duplication with art-dupl (requires art-dupl in PATH)" ''
              if ! command -v art-dupl &>/dev/null; then
                echo "art-dupl not found. Install: go install github.com/LarsArtmann/art-dupl/cmd/art-dupl@latest" >&2
                exit 1
              fi
              art-dupl . -t 50 "$@"
            '';

            error-audit =
              mkApp "error-audit"
                "Gate error-handling violations with erraudit (requires erraudit in PATH; see AGENTS.md)"
                ''
                  bash scripts/error-audit.sh "$@"
                '';

            clean = mkApp "clean" "Clean build and test artifacts" ''
              trash-put coverage.out 2>/dev/null || true
              go clean -testcache
            '';
          };
        };

      flake.overlays.default = final: _prev: {
        go-finding = mkGoFinding (final.buildGoModule.override { go = final.go_1_27; });
      };
    };
}
