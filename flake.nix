{
  description = "Cross-platform shared clipboard";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }: flake-utils.lib.eachDefaultSystem (system: let
    pkgs = nixpkgs.legacyPackages.${system};
  in rec {
    packages.uniclip = pkgs.buildGoModule {
      name = "uniclip";
      src = ./.;
      vendorSha256 = "sha256-Cmgb1BAAmQ5CnruNzCGKYd2qjTNttWemrlVIRHpfS2I=";
      meta = with pkgs.lib; {
        description = "Cross-platform shared clipboard";
        homepage = "https://github.com/quackduck/uniclip";
        license = licenses.mit;
        platforms = platforms.linux ++ platforms.darwin;
      };
    };
    defaultPackage = packages.uniclip;
  });
}
