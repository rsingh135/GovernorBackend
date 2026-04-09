# Fixing Go Certificate Issues on macOS

If you're seeing TLS certificate errors when running `go mod tidy` or `go get`, try these solutions:

## Solution 1: Update macOS Certificates

```bash
# Update certificates via Homebrew
brew install ca-certificates

# Or update via macOS Software Update
softwareupdate --install --all
```

## Solution 2: Configure Go to Use System Certificates

```bash
# Find your certificate bundle location
ls -la /etc/ssl/certs/cert.pem
ls -la /usr/local/etc/ca-certificates/cert.pem

# Set Go to use system certificates (if available)
export SSL_CERT_FILE=/etc/ssl/certs/cert.pem
```

## Solution 3: Reinstall Go

Sometimes reinstalling Go fixes certificate issues:

```bash
brew reinstall go
# Or if using official installer, reinstall from go.dev
```

## Solution 4: Temporary Workaround

If you need to proceed immediately, you can disable checksum verification temporarily:

```bash
export GOSUMDB=off
go mod tidy
```

**Note:** This is less secure and should only be used temporarily.

## Solution 5: Use Git Directly

As a last resort, you can clone modules via git:

```bash
cd $GOPATH/pkg/mod
git clone https://github.com/google/uuid.git github.com/google/uuid@v1.6.0
git clone https://github.com/lib/pq.git github.com/lib/pq@v1.10.9
```

Then run `go mod tidy` to generate go.sum.

## Verify Fix

After trying a solution, verify it works:

```bash
cd backend
go mod tidy
go mod verify
```

If `go mod verify` succeeds, the certificate issue is resolved.
