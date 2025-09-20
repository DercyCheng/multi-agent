fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Generate gRPC code from protobuf definitions
    tonic_build::configure()
        .build_server(true)
        .build_client(true)
        .out_dir("src/generated")
        .compile(
            &[
                "proto/agent_core.proto",
                "proto/health.proto",
            ],
            &["proto"],
        )?;

    println!("cargo:rerun-if-changed=proto/agent_core.proto");
    println!("cargo:rerun-if-changed=proto/health.proto");
    println!("cargo:rerun-if-changed=build.rs");

    Ok(())
}