{
  perSystem =
    { self', pkgs, ... }:
    {
      devShells.default = pkgs.mkShell {
        inputsFrom = [ self'.packages.default ];
      };
    };

}
