{
  description = "go-humanize-linter — AST linter that detects hand-rolled reimplementations of go-humanize";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };

    go-nix-helpers = {
      url = "github:LarsArtmann/go-nix-helpers/master";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    go-finding = {
      url = "github:LarsArtmann/go-finding/master";
      flake = false;
    };

    go-linter-sdk = {
      url = "github:LarsArtmann/go-linter-sdk/master";
      flake = false;
    };

    go-error-family = {
      url = "github:LarsArtmann/go-error-family/master";
      flake = false;
    };

    gogenfilter = {
      url = "github:LarsArtmann/gogenfilter/master";
      flake = false;
    };
  };

  outputs =
    inputs@{
      self,
      flake-parts,
      ...
    }:
    let
      version = self.shortRev or self.dirtyShortRev or "dev";
    in
    flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [ inputs.go-nix-helpers.flakeModules.go-standard ];

      go-standard = {
        pname = "go-humanize-linter";
        vendorHash = "sha256-/XGqi+IY3WYD30bSFOp73wyfm+oeY3BdL+/ozXJsSZc=";
        description = "AST linter that detects hand-rolled reimplementations of go-humanize";
        enableCheck = false;
        subPackages = [ "cmd/go-humanize-linter" ];

        deps = {
          "github.com/larsartmann/go-finding" = inputs.go-finding;
          "github.com/larsartmann/go-linter-sdk" = inputs.go-linter-sdk;
          "github.com/larsartmann/go-error-family" = inputs.go-error-family;
          "github.com/LarsArtmann/gogenfilter/v3" = inputs.gogenfilter;
        };

        src = inputs.nixpkgs.lib.cleanSourceWith {
          src = ./.;
          filter =
            path: _type:
            let
              base = baseNameOf path;
              excludedDirs = [
                ".direnv"
                "result"
                "reports"
                "custom-gcl"
              ];
              isInExcludedDir = inputs.nixpkgs.lib.any (
                d: inputs.nixpkgs.lib.hasInfix "/${d}/" path
              ) excludedDirs;
              isExcludedBase = inputs.nixpkgs.lib.elem base excludedDirs;
            in
            !(isInExcludedDir || isExcludedBase);
        };

        ldflags = [
          "-s"
          "-w"
          "-X main.version=${version}"
        ];

        extraBuildAttrs.preBuild = "export GOEXPERIMENT=jsonv2";

        shellExtraEnv = {
          GOEXPERIMENT = "jsonv2";
        };

        devShellExtraPackages = pkgs: [
          pkgs.golines
          pkgs.gopls
          pkgs.gotools
          pkgs.trash-cli
        ];

        enableNixfmt = true;
      };

      perSystem =
        {
          pkgs,
          lib,
          config,
          ...
        }:
        let
          mkApp = name: script: {
            type = "app";
            program = "${
              pkgs.writeShellApplication {
                inherit name;
                runtimeInputs = [
                  pkgs.go_1_26
                  pkgs.golangci-lint
                  pkgs.trash-cli
                ];
                text = script;
              }
            }/bin/${name}";
          };

          goEnv = ''
            export GOEXPERIMENT=jsonv2
          '';
        in
        {
          treefmt.programs.golines = {
            enable = true;
            maxLength = 120;
          };

          apps = {
            test = lib.mkForce (
              mkApp "test" ''
                ${goEnv}
                go test ./... -count=1 "$@"
              ''
            );

            test-race = mkApp "test-race" ''
              ${goEnv}
              go test ./... -race -count=1 "$@"
            '';

            bench = mkApp "bench" ''
              ${goEnv}
              go test ./... -bench=. -benchmem "$@"
            '';

            build = mkApp "build" ''
              ${goEnv}
              go build ./...
              go build -o go-humanize-linter ./cmd/go-humanize-linter/
            '';

            vet = mkApp "vet" ''
              ${goEnv}
              go vet ./...
            '';

            lint = lib.mkForce (
              mkApp "lint" ''
                ${goEnv}
                output=$(golangci-lint run ./... 2>&1)
                code=$?
                echo "$output" | grep -v 'Found unknown linters in //nolint directives'
                exit $code
              ''
            );

            coverage = mkApp "coverage" ''
              ${goEnv}
              mkdir -p reports
              go test ./... -coverprofile=reports/coverage.out -covermode=atomic "$@"
              go tool cover -func=reports/coverage.out
            '';

            custom-lint = mkApp "custom-lint" ''
              ${goEnv}
              golangci-lint custom
              ./custom-gcl run -c .golangci.custom.yml ./... "$@"
            '';
          };

          # Fast vendorHash drift check: forces realization of the goModules
          # FOD. If vendorHash doesn't match go.sum, the FOD fails with a
          # clear hash mismatch error — before any Go code compiles.
          checks.vendor-hash = pkgs.runCommand "vendor-hash" { } ''
            echo "vendor hash verified: ${config.packages.default.goModules}"
            touch $out
          '';
        };
    };
}
