{
  description = "Prints the effective Claude Code sandbox status as a colored statusline label";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        version = "1.3.0"; # x-release-please-version
      in
      {
        packages = {
          boxed = pkgs.buildGoModule {
            pname = "boxed";
            inherit version;

            src = ./.;

            vendorHash = "sha256-tvyLq3Bi+xniB5/QhW1qu16d3VhPb4J5J9ls5OVAWAs=";

            # Keep in sync with the ldflags in .goreleaser.yaml so `nix build`
            # produces the same stripped binary and version metadata as a
            # release build. -trimpath is already passed by buildGoModule.
            # commit/date are omitted: they would break Nix's reproducibility
            # and buildVersion() degrades gracefully without them.
            ldflags = [
              "-w"
              "-s"
              "-X main.version=${version}"
            ];

            # Match .goreleaser.yaml (CGO_ENABLED=0): a static, cgo-free binary.
            env.CGO_ENABLED = 0;

            meta = {
              description = "Prints the effective Claude Code sandbox status as a colored statusline label";
              homepage = "https://github.com/jamestelfer/boxed";
              license = pkgs.lib.licenses.asl20;
              mainProgram = "boxed";
            };
          };

          default = self.packages.${system}.boxed;
        };
      });
}
