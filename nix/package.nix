{
  perSystem =
    { pkgs, ... }:
    {
      packages.default = pkgs.buildGoModule {
        name = "coredns-data-mesher";
        src = ../.;
        vendorHash = "sha256-B99MCU7P604qyVAXP75gK6rb8DpKw4Kq8CUY71KIixY=";
      };
    };
}
