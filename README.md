# OSS Manager

OSS Manager is a command-line tool for managing Aliyun OSS (Object Storage Service) resources. It provides a terminal-based user interface for browsing, uploading, downloading, and managing files in OSS buckets.

## Project Initialization

This project was initialized as part of Task 1 in the development task list. The initialization process included:

1. Creating the Go project basic structure
2. Initializing the Go module
3. Adding necessary dependencies
4. Setting up the basic project layout
5. Testing the compilation

## Project Structure

The project follows standard Go project layout:

```
ossmanager/
├── cmd/
│   └── ossmanager/     # Application entry point
├── docs/               # Documentation
├── internal/           # Private application code
│   ├── config/         # Configuration management
│   ├── logger/         # Logging system
│   ├── oss/            # OSS interface
│   └── ui/             # User interface
├── pkg/                # Public libraries
│   └── utils/          # Utility functions
├── go.mod              # Go module file
├── go.sum              # Go module checksum
├── main.go             # Main application entry point
└── README.md           # This file
```

## Dependencies

The project uses the following main dependencies:

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - A powerful TUI framework
- [Viper](https://github.com/spf13/viper) - Configuration management
- [Aliyun OSS SDK](https://github.com/aliyun/aliyun-oss-go-sdk) - Aliyun OSS API client

## Build and Run

To build the project:

```bash
go build
```

To run the application:

```bash
./ossmanager
```

## Features

### Credential Management

The application includes a secure credential management system with the following features:

- **Secure Storage**: OSS credentials (AccessKeyID and AccessKeySecret) are stored in an encrypted format using AES-256-GCM encryption.
- **Password Protection**: All credentials are protected by a master password.
- **Master Password Management**: Users can change the master password securely, which will re-encrypt all stored credentials.
- **Multiple Accounts**: Support for managing credentials for multiple OSS accounts.
- **Secure File Format**: Credentials are stored in a JSON file with appropriate file permissions.
- **Password Verification**: The system verifies the master password by attempting to decrypt stored credentials.

### Configuration Management

Configuration capabilities include:

- **Hot-reload**: Configuration changes are automatically detected and applied.
- **Observers Pattern**: Components can register to be notified of configuration changes.
- **Default Values**: Sensible defaults for all configuration options.

### Logging System

The application includes a flexible logging system with:

- **Multiple Log Levels**: Support for Debug, Info, Warn, Error, and Fatal log levels.
- **Context Support**: Logging with contextual information.
- **File and Console Output**: Logs can be directed to files, console, or both.

## Development Status

This project is under active development. Current implementations include:

- Complete credential management system
- Configuration management with hot-reload support
- Robust logging system
- Basic project structure

The next tasks will focus on implementing the OSS interface and user interface components.

## Next Steps

The next development tasks will focus on:

1. Implementing the OSS interface (Tasks 7-11)
2. Implementing the user interface (Tasks 17-23)
3. Testing and optimization