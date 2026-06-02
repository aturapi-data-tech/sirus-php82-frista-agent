#!/bin/bash
# Build agent + bundle installer (user-session autostart) untuk Windows.
#
# Output: dist/ — copy folder ini ke PC Windows, dobel-klik setup.bat (TANPA admin).
#
# Beda dengan print-agent: frista-agent BUKAN Windows Service. Ia jalan di sesi
# desktop user (autostart saat logon) supaya bisa mengetik ke jendela FRISTA.
#
# Pemakaian: ./build-windows.sh

set -e

DIST=dist

rm -rf "$DIST"
mkdir -p "$DIST"

echo "[1/3] Build Go agent for Windows AMD64..."
GOOS=windows GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o "$DIST/sirus-frista-agent.exe" \
    .

echo "[2/3] Bundle config + installer scripts..."
cp config.json "$DIST/config.json"
cp installer/setup.bat "$DIST/setup.bat"
cp installer/setup.ps1 "$DIST/setup.ps1"
cp installer/uninstall.bat "$DIST/uninstall.bat"
cp installer/uninstall.ps1 "$DIST/uninstall.ps1"
cp installer/README-INSTALL.txt "$DIST/README-INSTALL.txt"

echo "[3/3] Done!"
echo ""
echo "===== Bundle siap distribusi: $DIST/ ====="
ls -lh "$DIST/"
du -sh "$DIST/"
echo ""
echo "Cara deploy ke PC Windows:"
echo "  1. Copy seluruh isi $DIST/ ke PC user (mis. C:\\SirusFristaAgent)"
echo "  2. Edit config.json — isi username, password, fristaPath"
echo "  3. Dobel-klik setup.bat (tanpa admin) → daftar autostart + langsung jalan"
