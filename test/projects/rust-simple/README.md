# Rust Simple Test Project

This is a test project used by mvx integration tests to verify Rust/Cargo functionality.

## Project Structure

```
rust-simple/
├── Cargo.toml           # Cargo project configuration
├── .gitignore          # Git ignore patterns
├── src/
│   ├── main.rs         # Main application with comprehensive examples
│   └── lib.rs          # Library code with extensive unit tests
└── README.md           # This file
```

## Features

- **Rust Application**: Console app demonstrating various Rust features
- **Comprehensive Tests**: 10 unit tests covering different scenarios
- **Dependencies**: Uses serde for JSON serialization (tests dependency management)
- **Modern Rust**: Uses Rust 2021 edition with idiomatic code
- **Real-world Examples**: Person struct, Calculator, JSON handling

## Code Examples

### Person Management
- Create persons with name and age
- Optional email addresses
- Adult filtering (18+ years)
- JSON serialization/deserialization

### Calculator
- Basic arithmetic operations (add, subtract, multiply, divide)
- Safe division with Option return type
- Comprehensive test coverage

### JSON Handling
- Serialize structs to JSON
- Deserialize JSON back to structs
- Error handling for malformed JSON

## Usage in Tests

This project is used by the integration test `TestMvxBinary/RustIntegration` to:

1. Test Rust tool installation and setup
2. Verify environment variable configuration (RUSTUP_HOME, CARGO_HOME)
3. Test Cargo build execution through mvx custom commands
4. Verify Rust compilation and test execution
5. Test dependency management (serde crates)
6. Confirm proper integration of rustc and cargo tools

## Manual Testing

You can also use this project for manual testing:

```bash
# Copy to a temporary directory
cp -r test/projects/rust-simple /tmp/rust-test
cd /tmp/rust-test

# Initialize mvx project
mvx init

# Add Rust tools
mvx tools add rust 1.84.0

# Setup environment
mvx setup

# Add custom commands to .mvx/config.json5:
# "cargo-build": { "description": "Build with Cargo", "script": "cargo build" }
# "cargo-test": { "description": "Test with Cargo", "script": "cargo test" }
# "cargo-run": { "description": "Run application", "script": "cargo run" }

# Test the build
mvx cargo-build
mvx cargo-test
mvx cargo-run
```

## Expected Outputs

- **Build**: Should compile successfully and create `target/` directory
- **Test**: Should run 10 unit tests, all passing
- **Run**: Should output "🦀 Hello from Rust via mvx!" and demonstrate various features

## Test Coverage

The project includes tests for:
- Person struct creation and methods
- Calculator arithmetic operations
- Adult filtering functionality
- JSON serialization/deserialization
- Edge cases (division by zero, empty collections, etc.)
