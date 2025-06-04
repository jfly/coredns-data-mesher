{ inputs, ... }:
{
  imports = [
    inputs.treefmt-nix.flakeModule
    inputs.git-hooks-nix.flakeModule
  ];

  perSystem.pre-commit.settings.hooks.treefmt.enable = true;

  perSystem.treefmt = {
    projectRootFile = "flake.nix";
    programs = {
      gofumpt.enable = true;
      nixfmt.enable = true;
    };
  };
}
