#!/bin/bash
# Script untuk mengekspor backend menjadi file image (.tar) ke folder swiftlead-backend-local
set -e

echo "Mengekspor SwiftLead Backend..."
echo "[1/2] Sedang mem-build Docker Image..."
docker build --platform linux/amd64 -t swiftlead-backend:offline .

echo "[2/2] Menyimpan Image ke swiftlead-backend-local (membutuhkan waktu beberapa saat)..."
docker save -o ../swiftlead-backend-local/swiftlead-backend-image.tar swiftlead-backend:offline

echo "Selesai! File swiftlead-backend-image.tar telah berhasil dibuat di folder swiftlead-backend-local."
echo "Sekarang folder 'swiftlead-backend-local' sudah LENGKAP dan SIAP JALAN."
echo "Anda bisa langsung membagikan folder tersebut kepada teman Anda (tanpa memberikan source code)."
