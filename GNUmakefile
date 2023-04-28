default: testacc

sweep:
	@echo "WARNING: This will destroy infrastructure. Use only in development accounts."
	TF_ACC=1 go test ./... -v -sweep -timeout 10m

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v -count=1 -parallel 2 $(TESTARGS) -timeout 120m
