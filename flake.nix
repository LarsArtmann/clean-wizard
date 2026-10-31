{
  description = "clean-wizard: Comprehensive CLI/TUI tool for system cleanup with type-safe architecture";

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
    inputs@{ self, flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      # Inline systems: nixpkgs 26.11 dropped x86_64-darwin, which github:nix-systems/default still lists.
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
      ];

      imports = [ inputs.treefmt-nix.flakeModule ];

      perSystem =
        {
          config,
          pkgs,
          lib,
          system,
          ...
        }:
        let
          version = self.rev or self.dirtyRev or "dev";

          vendorHash = "sha256-AXvndIraR9nGJzI0LgjhNw0gVySXPEdjCYvuWZVC6vs=";

          ldflags = [
            "-s"
            "-w"
            "-X main.version=${version}"
            "-X main.commit=${self.rev or "dirty"}"
            "-X main.date=${self.lastModifiedDate or "1970-01-01"}"
            "-X main.builtBy=nix"
          ];

          src = lib.fileset.toSource {
            root = ./.;
            fileset = lib.fileset.unions [
              ./go.mod
              ./go.sum
              ./cmd
              ./internal
            ];
          };

          clean-wizard = (pkgs.buildGoModule.override { go = pkgs.go_1_27; }) {
            pname = "clean-wizard";
            inherit
              version
              src
              vendorHash
              ldflags
              ;

            subPackages = [ "cmd/clean-wizard" ];

            env.CGO_ENABLED = 0;
            env.GOEXPERIMENT = "jsonv2";

            tags = [
              "netgo"
              "osusergo"
            ];

            doCheck = false;

            meta = {
              description = "Comprehensive CLI/TUI tool for system cleanup with type-safe architecture";
              homepage = "https://github.com/LarsArtmann/clean-wizard";
              license = lib.licenses.mit;
              mainProgram = "clean-wizard";
              maintainers = [
                {
                  name = "Lars Artmann";
                  github = "LarsArtmann";
                }
              ];
              platforms = lib.platforms.unix;
            };
          };
        in
        {
          packages = {
            default = clean-wizard;
          };

          checks = {
            format = config.treefmt.build.check self;
            build = clean-wizard;

            test = clean-wizard.overrideAttrs (_old: {
              doCheck = true;
            });

            go-vet = clean-wizard.overrideAttrs (_old: {
              doCheck = false;
              buildPhase = ''
                runHook preBuild
                go vet ./...
                runHook postBuild
              '';
              installPhase = ''
                runHook preInstall
                touch $out
                runHook postInstall
              '';
            });

            # H007 (manual byte parsing) regression gate: the humanize-linter
            # runs over the repo source; findings fail the check (TODO 18).
            humanize-lint =
              let
                lintTool = (pkgs.buildGoModule.override { go = pkgs.go_1_27; }) {
                  pname = "clean-wizard-humanize-lint";
                  version = "0.1.0";
                  src = self;
                  sourceRoot = "source/tools/lint";
                  # vendor/ is committed under tools/lint, so no hash needed.
                  vendorHash = null;
                  env.GOEXPERIMENT = "jsonv2";
                  doCheck = false;
                };
              in
              pkgs.runCommand "humanize-lint-check"
                {
                  nativeBuildInputs = [ lintTool ];
                }
                ''
                  cd ${self}
                  lint
                  touch $out
                '';
          };

          apps = {
            default = {
              type = "app";
              program = lib.getExe clean-wizard;
            };
          };

          devShells = {
            default = pkgs.mkShell {
              inputsFrom = [ config.packages.default ];

              packages = with pkgs; [
                go_1_27
                gopls
                golangci-lint
                gotools
                gofumpt
                govulncheck
                jq
                goreleaser
                cosign
              ];

              GOWORK = "off";
              GOEXPERIMENT = "jsonv2";

              shellHook = ''
                echo "🧙 clean-wizard development environment"
                echo "  Go:    $(go version)"
                echo "  Shell: ${system}"
              '';
            };

            ci = pkgs.mkShellNoCC {
              packages = with pkgs; [
                go_1_27
                golangci-lint
                jq
              ];

              GOEXPERIMENT = "jsonv2";
            };
          };

          treefmt = {
            # Vendored third-party code is formatted upstream, not here.
            settings.excludes = [ "tools/lint/vendor/*" ];
            programs = {
              nixfmt.enable = true;
              gofmt = {
                enable = true;
                package = pkgs.go_1_27;
              };
              goimports = {
                enable = true;
                # goimports resolves `go` from PATH; with an older go it tries
                # to download the go1.27 toolchain the module requires, which
                # dies in the offline sandbox (ticketed TODO #29). The wrapper
                # script puts go_1_27 on PATH first, so toolchain resolution
                # stays local inside the sandbox.
                package = pkgs.writeShellScriptBin "goimports" ''
                  export PATH="${pkgs.go_1_27}/bin:$PATH"
                  exec "${pkgs.gotools}/bin/goimports" "$@"
                '';
              };
            };
          };
        };

      flake.overlays.default = final: _prev: {
        clean-wizard = self.packages.${final.stdenv.system}.default;
      };
    };
}
