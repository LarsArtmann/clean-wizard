{
  description = "go-error-family — Error classification and handling for Go";

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
      flake-parts,
      treefmt-nix,
      ...
    }:
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
          goPkg = pkgs.go_1_26;

          mkApp = name: runtimeInputs: text: {
            type = "app";
            program = "${
              pkgs.writeShellApplication {
                inherit name runtimeInputs;
                text = ''
                  ${text}
                '';
              }
            }/bin/${name}";
          };
        in
        {
          treefmt = {
            projectRootFile = "go.mod";
            programs = {
              gofumpt.enable = true;
              goimports.enable = true;
              golines.enable = true;
              nixfmt.enable = true;
            };
          };

          checks.format = config.treefmt.build.check self;
          devShells = {
            default = pkgs.mkShellNoCC {
              packages = [
                goPkg
                pkgs.golangci-lint
                pkgs.gotools
                pkgs.trash-cli
              ];

              shellHook = ''
                echo "go-error-family dev shell — $(go version)"
              '';
            };

            ci = pkgs.mkShellNoCC {
              packages = [
                goPkg
                pkgs.golangci-lint
              ];
            };
          };

          checks = {
            build = pkgs.runCommand "go-error-family-build" { nativeBuildInputs = [ goPkg ]; } ''
              export HOME=$TMPDIR
              export CGO_ENABLED=0
              cp -r ${./.} src && chmod -R u+w src && cd src
              ${goPkg}/bin/go build ./...
              touch $out
            '';

            build-standalone =
              pkgs.runCommand "go-error-family-build-standalone" { nativeBuildInputs = [ goPkg ]; }
                ''
                  export HOME=$TMPDIR
                  export CGO_ENABLED=0
                  export GOWORK=off
                  cp -r ${./.} src && chmod -R u+w src && cd src
                  ${goPkg}/bin/go build ./...
                  touch $out
                '';

            lint =
              pkgs.runCommand "go-error-family-lint"
                {
                  nativeBuildInputs = [
                    goPkg
                    pkgs.golangci-lint
                  ];
                }
                ''
                  export HOME=$TMPDIR
                  export CGO_ENABLED=0
                  cp -r ${./.} src && chmod -R u+w src && cd src
                  ${pkgs.golangci-lint}/bin/golangci-lint run ./...
                  touch $out
                '';
          };

          apps = {
            test = mkApp "test" [ goPkg ] ''
              go test ./... -count=1 "$@"
            '';

            test-race = mkApp "test-race" [ goPkg ] ''
              go test ./... -race -count=1 "$@"
            '';

            build = mkApp "build" [ goPkg ] ''
              go build ./...
            '';

            vet = mkApp "vet" [ goPkg ] ''
              go vet ./...
            '';

            lint = mkApp "lint" [ pkgs.golangci-lint ] ''
              golangci-lint run ./...
            '';

            coverage = mkApp "coverage" [ goPkg ] ''
              go test ./... -coverprofile=coverage.out -covermode=atomic "$@"
              go tool cover -func=coverage.out
            '';

            clean = mkApp "clean" [ goPkg pkgs.trash-cli ] ''
              trash-put coverage.out 2>/dev/null || true
              go clean -testcache
            '';
          };
        };
    };
}
