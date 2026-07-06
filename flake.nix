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
      in
      {
        packages = {
          boxed = pkgs.buildGoModule {
            pname = "boxed";
            version = "1.2.0"; # x-release-please-version

            src = ./.;

            vendorHash = "sha256-tvyLq3Bi+xniB5/QhW1qu16d3VhPb4J5J9ls5OVAWAs=";

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
