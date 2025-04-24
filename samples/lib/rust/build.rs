fn main() {
    // Specify the library search path (equivalent to -Lbin)
    println!("cargo:rustc-link-search=native=../../../bin");
    // Link against libzscript.so (equivalent to -lzscript)
    println!("cargo:rustc-link-lib=zscript");
    // Embed the library path for runtime (equivalent to -Wl,-rpath,bin)
    println!("cargo:rustc-link-arg=-Wl,-rpath,../../../bin");
}