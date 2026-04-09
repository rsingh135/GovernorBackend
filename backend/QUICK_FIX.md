# Quick Fix for Go Dependencies

Since you were able to download the modules (you saw the checksum mismatch error), the certificate issue might be resolved or intermittent. 

## Try this:

1. **Delete go.sum and regenerate it:**
   ```bash
   cd backend
   rm go.sum
   go mod tidy
   ```

2. **If you still get certificate errors, use this workaround:**
   ```bash
   cd backend
   rm go.sum
   export GOSUMDB=off
   go mod tidy
   ```

3. **After go.sum is generated, verify it works:**
   ```bash
   go build ./cmd/api
   ```

The checksum mismatch error you saw means Go successfully downloaded the modules but the checksum in go.sum was wrong. After regenerating go.sum, it should work!

## If certificate errors persist:

See `FIX_CERTIFICATES.md` for detailed certificate troubleshooting steps.
