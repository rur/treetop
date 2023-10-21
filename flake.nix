{
  description = "Nix based developement env for Treetop library";

  inputs.nixpkgs.url = "github:nixos/nixpkgs";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (
      system:
        let
          pkgs = import nixpkgs { inherit system; };
        in 
          {
            devShell = pkgs.mkShell {
              buildInputs = with pkgs; [ 
                go
                gopls
                gops
                delve
                go-tools
                errcheck
                reftools
                revive
                golangci-lint
                gomodifytags
                gotags
                impl
                go-motion
                iferr
              ];
            };
          }
    );
}
