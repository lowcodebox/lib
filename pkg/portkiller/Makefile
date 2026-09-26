.PHONY: test test-short test-race test-linux test-darwin test-windows

test:            ## все тесты
	go test ./... -v -count=1

test-short:      ## только юнит-тесты, без интеграционных
	go test ./... -short -v -count=1

test-race:       ## с race detector
	go test ./... -race -count=1

test-linux:
	GOOS=linux go test ./... -count=1

test-darwin:
	GOOS=darwin go test ./... -count=1

test-windows:
	GOOS=windows go test ./... -count=1