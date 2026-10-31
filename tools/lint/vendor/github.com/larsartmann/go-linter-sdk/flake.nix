{
  description = "go-linter-sdk — Shared scaffolding for LarsArtmann Go linters";

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
      nixpkgs,
      flake-parts,
      treefmt-nix,
      ...
    }:
    let
      inherit (nixpkgs) lib;
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
          pkgs,
          ...
        }:
        let
          goPkg = pkgs.go_1_27;

          # gitTracked excludes gitignored files (go.work, go.work.sum,
          # reports/ contents), so workspace mode can never leak into the
          # hermetic build the way it can in the developer's shell.
          src = lib.fileset.toSource {
            root = ./.;
            fileset = lib.fileset.gitTracked ./.;
          };

          # The vendored dependency graph is fetched through vendorHash and
          # shared by every Go check. buildGoModule's Go must match goPkg,
          # or the vendor output hash differs (see go-finding flake note).
          mkCheckModule =
            {
              doCheck ? true,
            }:
            (pkgs.buildGoModule.override { go = goPkg; }) {
              inherit src;
              pname = "go-linter-sdk";
              version = "dev";
              vendorHash = "sha256-oK66kB5PJEC1z0iDpNkqG1KAJFe2qkAlhknRGBVyj0M=";
              inherit doCheck;
            };

          installTouch = ''
            runHook preInstall
            touch $out
            runHook postInstall
          '';

          mkApp = name: description: script: {
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
              inherit description;
              mainProgram = name;
              homepage = "https://github.com/larsartmann/go-linter-sdk";
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
              # goimports shells out to `go` at runtime; unless it sees a Go at
              # least as new as go.mod's floor, the toolchain auto-switch tries
              # to download a newer Go and fails in the network-less check
              # sandbox. Pin it to the same toolchain as the rest of the flake.
              goimports = {
                enable = true;
                package = pkgs.gotools.override { go = goPkg; };
              };
              golines = {
                enable = true;
                maxLength = 120;
              };
              nixfmt.enable = true;
            };
          };

          devShells.default = pkgs.mkShell {
            packages = [
              goPkg
              pkgs.golangci-lint
              pkgs.gofumpt
              pkgs.golines
              pkgs.gopls
              pkgs.gotools
              pkgs.trash-cli
            ];

            shellHook = ''
              echo "go-linter-sdk dev shell — $(go version)"
            '';
          };

          devShells.ci = pkgs.mkShellNoCC {
            packages = [
              goPkg
              pkgs.golangci-lint
            ];
          };

          checks = {
            # go build + go test ./... in the network-less sandbox.
            go-test = mkCheckModule { };

            go-vet = (mkCheckModule { doCheck = false; }).overrideAttrs (_: {
              pname = "go-linter-sdk-vet";
              buildPhase = ''
                runHook preBuild
                go vet ./...
                runHook postBuild
              '';
              inherit installTouch;
            });

            go-lint = (mkCheckModule { doCheck = false; }).overrideAttrs (_: {
              pname = "go-linter-sdk-lint";
              nativeBuildInputs = [
                pkgs.golangci-lint
                goPkg
              ];
              preBuild = ''
                # golangci-lint and its formatters need a writable HOME/cache
                # (the sandbox HOME is the read-only /homeless-shelter) and a
                # `go` on PATH (gci shells out to `go env`).
                export HOME=$TMPDIR
                export GOLANGCI_LINT_CACHE=$TMPDIR/golangci-lint-cache
                export GOCACHE=$TMPDIR/go-cache
              '';
              buildPhase = ''
                runHook preBuild
                golangci-lint run ./...
                runHook postBuild
              '';
              inherit installTouch;
            });

            # The doc-drift guard (see scripts/check-doc-versions.sh). Needs a
            # git repo because the script anchors on `git rev-parse`.
            doc-versions =
              pkgs.runCommand "doc-versions-check"
                {
                  nativeBuildInputs = [
                    pkgs.git
                    pkgs.bash
                  ];
                }
                ''
                  export HOME=$TMPDIR
                  cp -r ${src} repo
                  cd repo
                  chmod -R u+w .
                  git init -q .
                  git add -A
                  git -c user.name=check -c user.email=check@local commit -qm init
                  ${pkgs.bash}/bin/bash scripts/check-doc-versions.sh
                  touch $out
                '';
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
              mkdir -p reports
              go test ./... -coverprofile=reports/coverage.out -covermode=atomic "$@"
              go tool cover -func=reports/coverage.out
            '';

            clean = mkApp "clean" "Clean build and test artifacts" ''
              # reports/coverage.out is the canonical path; the bare coverage.out
              # sweep removes artifacts from before the path was aligned.
              trash-put reports/coverage.out coverage.out 2>/dev/null || true
              go clean -testcache
            '';
          };
        };
    };
}
