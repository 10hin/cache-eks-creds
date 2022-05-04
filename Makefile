BIN = cache-eks-creds
ifeq ($(OS),Windows_NT)
	BIN = cache-eks-creds.exe
endif

RM_F = rm -f

.PHONY: build
build:
	go build -o $(BIN) ./main.go

.PHONY: clean
clean:
	$(RM_F) $(BIN)
