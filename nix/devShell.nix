{
  perSystem =
    { pkgs, ... }:
    {
      devShells.default = pkgs.mkShell {
        nativeBuildInputs = [ pkgs.go ];
        shellHook = ''
          # Generate go.mod and go.sum. See HACKING.md for details.
          if [ ! -f go.mod ]; then
            go mod init github.com/jfly/coredns-data-mesher
          fi
          if [ ! -f go.sum ]; then
            go mod tidy
          fi
        '';
      };
    };
}
