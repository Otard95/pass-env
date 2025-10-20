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
          version = "1.0.0";
        in pkgs.buildGoModule {

          pname = "ngm";
          inherit version;

          src = pkgs.fetchFromGitHub {
            owner = "otard95";
            repo = "pass-env";
            rev = "v${version}";
            hash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
          };

          vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";

        };
      }
    );
}
