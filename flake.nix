{
  description = "Language Server Protocol implementation for QML";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        version = if self ? shortRev then "0.0.0-${self.shortRev}" else "dev";
      in
      {
        packages = {
          default = pkgs.buildGoModule {
            pname = "qml-language-server";
            inherit version;
            src = self;

            vendorHash = "sha256-J+0kFTKgluf+mabJepW+MGXUdHqYLFaUVAZEWcyHmyk=";

            ldflags = [
              "-s" "-w"
              "-X main.version=${version}"
            ];

            meta = with pkgs.lib; {
              description = "Language Server Protocol implementation for QML";
              homepage = "https://github.com/cushycush/qml-language-server";
              license = licenses.mit;
              maintainers = [ ];
              mainProgram = "qml-language-server";
            };
          };
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            golangci-lint
          ];
        };
      }
    );
}
