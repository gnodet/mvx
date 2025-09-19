# Gradle Java Simple Test Project

This is a test project used by mvx integration tests to verify Gradle functionality.

## Project Structure

```
gradle-java-simple/
├── build.gradle          # Gradle build configuration
├── gradle.properties     # Gradle settings optimized for testing
├── src/
│   ├── main/java/com/example/
│   │   └── App.java      # Simple Java application
│   └── test/java/com/example/
│       └── AppTest.java  # JUnit 5 tests
└── README.md            # This file
```

## Features

- **Java Application**: Simple console app with main method
- **JUnit 5 Tests**: Comprehensive unit tests with multiple test cases
- **Gradle Configuration**: Standard Java plugin with application plugin
- **Maven Central**: Uses standard Maven Central repository
- **Test Optimized**: Gradle settings optimized for CI/testing (no daemon, no parallel)

## Usage in Tests

This project is used by the integration test `TestMvxBinary/GradleIntegration` to:

1. Test Gradle tool installation and setup
2. Verify environment variable configuration (JAVA_HOME, GRADLE_HOME)
3. Test Gradle build execution through mvx custom commands
4. Verify Java compilation and test execution
5. Confirm proper integration between Java and Gradle tools

## Manual Testing

You can also use this project for manual testing:

```bash
# Copy to a temporary directory
cp -r test/projects/gradle-java-simple /tmp/gradle-test
cd /tmp/gradle-test

# Initialize mvx project
mvx init

# Add tools
mvx tools add java 17
mvx tools add gradle 8.5

# Setup environment
mvx setup

# Add custom commands to .mvx/config.json5:
# "gradle-build": { "description": "Build with Gradle", "script": "gradle build" }
# "gradle-test": { "description": "Test with Gradle", "script": "gradle test" }
# "gradle-run": { "description": "Run application", "script": "gradle run" }

# Test the build
mvx gradle-build
mvx gradle-test
mvx gradle-run
```

## Expected Outputs

- **Build**: Should compile successfully and create `build/` directory
- **Test**: Should run 6 JUnit tests, all passing
- **Run**: Should output "Hello, Gradle from mvx!"
