{
  description = "pass-env is like env (the unix util) but gets the env values from pass";
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  inputs.systems.url = "github:nix-systems/default";
  inputs.flake-utils = {
    url = "github:numtide/flake-utils";
    inputs.systems.follows = "systems";
  };

  outputs =
    { nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            cobra-cli
          ];
        };

        packages.default = let
        in pkgs.buildGoModule rec {

          pname = "pass-env";
          version = "1.1.1";

          src = pkgs.fetchFromGitHub {
            owner = "otard95";
            repo = "pass-env";
            rev = "v${version}";
            hash = "sha256-SPZnGKooe2srbRZGDjLN8IuT+Wp4/XXSrKCbXawfPcs=";
          };

          vendorHash = "sha256-hpAsYPhiYnTpY5Z7QZz9cr5RtleHnR1ezgoVaQ+cvp0=";

          subPackages = ["."];

        };
      }
    );
}
