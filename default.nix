{
  lib,

  buildGoModule,
}:
buildGoModule {
  pname = "kata";
  version = "main";

  src = ./.;

  vendorHash = "sha256-oFGo3SbyQoz5smikUd7TZuHTQjDJ0GAkxOF0zBhKtUc=";

  meta = with lib; {
    description = "A minimalist, high-performance terminal typing trainer for developers. Master your keyboard with real-world code (Go, Rust, Python) and natural languages.";
    homepage = "https://github.com/stiffis/kata";
    license = licenses.gpl3;
  };
}
