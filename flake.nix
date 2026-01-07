{
  description = "A minimalist, high-performance terminal typing trainer for developers. Master your keyboard with real-world code (Go, Rust, Python) and natural languages.";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
    systems.url = "github:nix-systems/default";
  };

  outputs =
    inputs@{
      self,
      flake-parts,
      ...
    }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = import inputs.systems;

      perSystem =
        {
          self',
          pkgs,
          system,
          ...
        }:
        {
          packages = {
            kata = pkgs.callPackage ./default.nix { };
            default = self'.packages.kata;
          };

          devShells.default = pkgs.callPackage ./shell.nix {
            inherit (self'.packages) kata;
          };
        };
    };
}
