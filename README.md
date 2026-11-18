cat > README.md <<'EOF'
# qrgen
Generates vCard QR codes from CSV files. Runs as a CLI or as a small HTTP service.

CSV header (`;` separated): first_name;last_name;org;title;phone;email

    go run ./cmd/qrgen -in ./input -out ./qrcodes   # CLI
    go run ./cmd/qrgen                               # server on :8080
    make docker                                      # container image
EOF