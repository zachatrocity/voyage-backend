{
  description = "Voyage - Email Travel Aggregator";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            gotools
            notmuch
            pkg-config
          ];

          shellHook = ''
            echo "Voyage development environment"
            echo "Notmuch version: $(notmuch --version)"
            export CGO_LDFLAGS="-lnotmuch"
            export CGO_CFLAGS="-I${pkgs.notmuch}/include"
            export LD_LIBRARY_PATH="${pkgs.notmuch}/lib:$LD_LIBRARY_PATH"
            export PKG_CONFIG_PATH="${pkgs.notmuch}/lib/pkgconfig:$PKG_CONFIG_PATH"
          '';
        };
      }
    );
}
