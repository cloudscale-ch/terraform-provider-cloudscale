default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v -parallel 2 $(TESTARGS) -timeout 120m
