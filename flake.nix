{
  description = "clean-wizard: Comprehensive CLI/TUI tool for system cleanup with type-safe architecture";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };

    systems.url = "github:nix-systems/default";

    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs@{ self, flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = import inputs.systems;

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

          vendorHash = "sha256-g17cVEsjDmBDiDHObW2Pp/X1y6Xrr4Qu7emWbVSm3mo=";
          proxyVendor = true;

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
              ./pkg
            ];
          };

          clean-wizard = pkgs.buildGoModule {
            pname = "clean-wizard";
            inherit
              version
              src
              vendorHash
              ldflags
              ;

            subPackages = [ "cmd/clean-wizard" ];

            env.CGO_ENABLED = 0;

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
              maintainers = [ lib.maintainers.larsartmann ];
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

            test = clean-wizard.overrideAttrs (old: {
              doCheck = true;
            });

            go-vet = clean-wizard.overrideAttrs (old: {
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
                go
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

              shellHook = ''
                echo "🧙 clean-wizard development environment"
                echo "  Go:    $(go version)"
                echo "  Shell: ${system}"
              '';
            };

            ci = pkgs.mkShellNoCC {
              packages = with pkgs; [
                go
                golangci-lint
                jq
              ];
            };
          };

          treefmt = {
            programs = {
              nixfmt.enable = true;
              gofmt.enable = true;
              goimports.enable = true;
            };
          };
        };

      flake.overlays.default = final: _prev: {
        clean-wizard = final.callPackage ./package.nix { };
      };
    };
}
