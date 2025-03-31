# PastePal-OS: Zero Knowledge Security Paste Application

PastePal-OS is a secure paste sharing application that implements zero-knowledge security principles. All data is encrypted locally before being sent to the server, ensuring that only you can access your content.

## Key Security Features

- **Zero-Knowledge Security**: Your data is encrypted locally before being sent to the server
- **Client-Side Encryption**: All encryption/decryption happens on your device
- **Master Password**: Your master password never leaves your device
- **Symmetric Key Encryption**: Your pastes are encrypted with a strong symmetric key
- **Password-Based Key Derivation**: Uses PBKDF2 with SHA-256 for secure key derivation

## How It Works

### Registration

1. You create an account with an email and master password
2. A master key is derived from your password using PBKDF2
3. A random symmetric key is generated for encrypting your pastes
4. The symmetric key is encrypted with your master key
5. Only your email, password hash, and encrypted symmetric key are sent to the server

### Login

1. You enter your email and master password
2. Your master key is derived locally from your password
3. A password hash is sent to the server for authentication
4. The server returns your encrypted symmetric key
5. Your master key decrypts the symmetric key locally

### Creating Pastes

1. Your paste content is encrypted locally with your symmetric key
2. Only the encrypted data is sent to the server
3. The server stores the encrypted data but cannot read it

### Reading Pastes

1. Encrypted paste data is retrieved from the server
2. The data is decrypted locally using your symmetric key

## Building and Running

```bash
# Build the application
go build -o pastepal.exe ./cmd/pastepal

# Run the application
./pastepal.exe
```

## Project Structure

- `cmd/pastepal`: Main application entry point
- `internal/auth`: Authentication and key management
- `internal/crypto`: Encryption and decryption utilities
- `internal/models`: Data models
- `internal/storage`: Local storage management
- `internal/api`: Server API client
- `internal/config`: Application configuration
- `internal/core`: Core application logic

## Dependencies

- golang.org/x/crypto: For cryptographic functions

## Security Considerations

- Your master password is never stored or transmitted in plain text
- The symmetric key is only stored in encrypted form
- All encryption/decryption happens locally on your device
- The server only sees encrypted data and cannot decrypt it

## Note

This application is designed to connect to a server that implements the corresponding API. The server component is not included in this repository.
