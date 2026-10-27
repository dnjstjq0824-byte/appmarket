# Work Order Progress Report: AM-ARCH-001

- **담당**: Lead Architect
- **상태**: `PARTIAL`
- **Git 브랜치**: `feat/AM-ARCH-001`
- **Git Commit SHA**: 없음. 신규 로컬 저장소이며 아직 커밋하지 않았다.

## 변경 파일

현재 파일은 모두 신규 미추적 파일이므로 `git diff --stat`에는 나타나지 않는다. 이 Work Order에서 작성한 파일은 `go.work`, `pnpm-workspace.yaml`, `package.json`, `backend/go.mod`, `backend/go.sum`, `backend/cmd/api/main.go`, `backend/internal/architecture_test.go`, `backend/internal/platform/errors/problem.go`, `backend/internal/platform/httpserver/server.go`, `backend/internal/platform/httpserver/server_test.go`이다.

## 실행 결과

| 명령 | 실제 결과 | 증빙 |
| --- | --- | --- |
| `go build ./...` | 종료 코드 0 | [go_build.log](go_build.log), SHA-256 `ddece55718a2bb5cca09e57ec72ad02c108c4441a5d8ee856d348e6197845ed3` |
| `go test -v -coverprofile=../artifacts/evidence/AM-ARCH-001/coverage.out ./...` | 종료 코드 0, 테스트 2개 통과. `httpserver` 패키지 커버리지 58.2%, 전체 Go 코드 커버리지 32.5%. | [go_test_arch.log](go_test_arch.log), SHA-256 `e0aede174c28bc99631f32bb46ae5d5081968c5806af0c2c88ba23091bdfb18b`; [coverage.out](coverage.out), SHA-256 `d3c6e308dbca5e436b9d53fcb919aab6cc2ec3554c4ffdc900d6748b552a5aff` |
| `go vet ./...` | 종료 코드 0 | [go_vet.log](go_vet.log), SHA-256 `ce8fce1ece304e130d9d8a580c1d95c1578d54c5aa21260a7058e927c6467511` |
| `go test -v -race ./internal/...` | 종료 코드 2. Windows 환경에 CGO 컴파일러가 없어 실행 불가. | [go_race_unavailable.log](go_race_unavailable.log), SHA-256 `e8fc29b727636a564d8a0b1a3372536f885f4fcac2b8f8849cda9aa3698b8264` |

첫 커버리지 명령은 PowerShell 인자 해석 때문에 실패했다. [go_test_arch_attempt1.log](go_test_arch_attempt1.log)에 원본 실패를 보존하고 올바른 인자로 재실행했다.

## 미완료 및 제한

- `TC-FUNC-001`의 아키텍처 import 검사 코드는 실행했으나 아직 도메인 모듈이 없어 실제 교차 모듈 금지 사례를 검증하지 못했다.
- `TC-OPS-001`의 격리 워커 강제 장애 테스트는 Docker 부재와 스캐너 워커 미구현으로 실행하지 못했다.
- Redis 기반 슬라이딩 윈도우 제한, OTel 추적, RFC 9457 오류, 실제 의존성 준비 상태 확인은 코드에 있다. `IdempotencyGuard`, 인증/RBAC, 감사 기록, 워커 디스패처, 운영 서비스 연결 통합 테스트는 남아 있다.
- 독립 감사가 없으므로 `VERIFIED`로 승격하지 않는다.

## 2026-10-28 후속 작업

- 도메인 import 경계 검사에 금지된 교차 도메인 및 직접 DB 쿼리 import를 실제로 거부하는 fixture 테스트를 추가했다. 아직 도메인 모듈이 없으므로 실제 모듈 간 참조 검증은 계속 미완료다.
- 신뢰 프록시 CIDR을 명시적으로 설정한 경우에만 `X-Forwarded-For`를 역순으로 검증해 클라이언트 IP를 산출한다. Redis 요청 제한은 이 IP를 사용한다. 미설정 시 직접 연결 주소를 사용한다.
- `scripts/verify-backend.ps1`로 `go build ./...`, `go test -v -coverprofile=...`, `go vet ./...`를 실행했다. 모두 종료 코드 0이다. 실제 출력은 [go_build_current.log](go_build_current.log), [go_test_current.log](go_test_current.log), [go_vet_current.log](go_vet_current.log), [coverage-current.out](coverage-current.out)에 남겼다. HTTP 서버 패키지 테스트 커버리지는 62.5%이다.
- 컨테이너 검증, `-race` 테스트, outbox/worker, IdempotencyGuard, 독립 감사가 아직 완료되지 않았으므로 상태는 `PARTIAL`이다.

