.PHONY: generate clean install-tools bootstrap

# Generate gRPC code from proto files
generate:
	@echo "Generating gRPC code from proto files..."
	@protoc -I. --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/oak/v1/*.proto
	@echo "[SUCCESS] Proto code generated successfully"

# Install required protoc plugins (using tools.go pattern)
install-tools:
	@echo "Installing build tools (tracked in tools.go)..."
	@cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI {} go install {}@latest

# Bootstrap: install tools and generate code (for new developers)
bootstrap: install-tools
	@echo "Running initial code generation..."
	$(MAKE) generate
	@echo "[SUCCESS] Bootstrap complete! You're ready to develop."

# Clean generated files
clean:
	find proto -name "*.pb.go" -delete