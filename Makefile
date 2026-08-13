BINARY_NAME := whoo
BUILD_FLAGS := -ldflags "-X whoo-cli/config.AppID=$(WHOOING_APP_ID) -X whoo-cli/config.AppSecret=$(WHOOING_APP_SECRET)"

ifneq (,$(wildcard .env))
	include .env
	export
endif

.PHONY: build dev run test vet check clean help

build:
	@if [ -z "$(WHOOING_APP_ID)" ] || [ -z "$(WHOOING_APP_SECRET)" ]; then \
		echo "[오류] WHOOING_APP_ID와 WHOOING_APP_SECRET이 필요합니다"; \
		echo ".env.example을 .env로 복사한 뒤 실제 값을 입력하세요"; \
		exit 1; \
	fi
	go build $(BUILD_FLAGS) -o $(BINARY_NAME) .

dev:
	go build -o $(BINARY_NAME) .

run: dev
	./$(BINARY_NAME)

test:
	go test ./...

vet:
	go vet ./...

check: test vet

clean:
	rm -f $(BINARY_NAME)

help:
	@echo "사용 가능한 명령어:"
	@echo "  make dev    개발 바이너리 빌드 (.env는 실행 시 자동 로드)"
	@echo "  make build  앱 자격 증명을 주입한 바이너리 빌드"
	@echo "  make run    개발 빌드 후 TUI 실행"
	@echo "  make check  테스트와 go vet 실행"
	@echo "  make clean  로컬 바이너리 삭제"
