# 차단 4종 즉시-테스트 런북 (LIVE_VERIFICATION_RUNBOOK)

외부 열쇠/클라우드 없이 막혀 있던 검증 4종을, 열쇠가 생기는 즉시 바로 실행하는 절차서.
코드·스크립트·IaC는 모두 준비되어 있다. 각 항목은 **투입물 → 명령 → 판정** 순서로 읽는다.

사전 프로브(열쇠 없이 실행 가능, 현 상태 출력):
`powershell -File scripts/smoke/live-readiness.ps1`

공통 전제: 아래 명령은 저장소 루트에서 실행한다. PowerShell에서는 `&&` 대신 `;`을 쓴다.

---

## 차단 1종: Toss/PortOne 실결제 E2E

- 차단 사유: live PG 키 없음. 현재 `TOSS_SECRET_KEY`, `PORTONE_API_KEY`가 비어 있어
  mock 기반 단위 테스트(`pg_client_test.go`)까지만 검증됨.
- 투입물: Toss live secret key 1개 또는 PortOne API key + store id 1조.
- 코드 위치: `backend/internal/modules/commerce/application/pg_client.go`
  (BaseURL 기본값 `https://api.tosspayments.com/v1`, `https://api.portone.io`),
  배선 `backend/cmd/api/main.go` (`PG_PROVIDER`, `PG_IS_LIVE`, `TOSS_SECRET_KEY`,
  `PORTONE_API_KEY`, `PORTONE_STORE_ID`).

절차 (Toss 예시):

1. 키 주입 후 API 재기동:
   `$env:TOSS_SECRET_KEY='<live-key>'; $env:PG_IS_LIVE='true'; docker compose --env-file deploy/.env -f deploy/docker-compose.yml -f deploy/docker-compose.override.yml up -d --build api`
2. 테스트 결제로 `paymentKey` + `orderId` + `amount` 확보 (PG사 테스트 결제창).
3. Confirm 호출 (서버 로그에 `pg.confirm` 성공 기록 확인).
4. 웹훅 서명 검증: `PG_WEBHOOK_SECRET` 설정 상태에서 PG사 콘솔 → 재전송 → 200 확인.

절차 (PortOne 예시): `$env:PG_PROVIDER='portone'`, `$env:PORTONE_API_KEY`, `$env:PORTONE_STORE_ID` 설정 후 동일.
판정: Confirm 200 + 웹훅 200 + 정산 원장(order) 상태 전이. 실패 시 `PG_IS_LIVE`를
지우고 mock 모드로 복귀 (`go test ./internal/modules/commerce/...` 재실행).

---

## 차단 2종: 실 S3(GCS) 오브젝트 스토리지 연동

- 차단 사유: 로컬은 LocalStack으로 검증됨. 실 AWS S3/GCS 버킷 쓰기 미검증.
- 투입물: 실 버킷 3개(검역/공개CDN/미디어) + 쓰기 권한(IAM role 또는 access key).
- 코드 위치: `backend/internal/modules/distribution/application/s3_storage.go`,
  셀렉터 `backend/internal/platform/httpserver/server.go` (`OBJECT_STORAGE=s3`).
  `S3_ENDPOINT`를 비우면 실 AWS S3, 크리덴셜을 비우면 SDK 기본 체인(IAM role).

절차:

1. 버킷 3개 생성 (terraform 완료 시 `deploy/terraform/aws/outputs.tf`의
   `quarantine_bucket`, `public_cdn_bucket` 값 사용).
2. API env 설정 후 재기동:
   `$env:OBJECT_STORAGE='s3'; $env:S3_ENDPOINT=''; $env:S3_REGION='ap-northeast-2'; $env:S3_USE_PATH_STYLE='false'`
   (IAM role 사용 시 `S3_ACCESS_KEY_ID`/`S3_SECRET_ACCESS_KEY`는 비워 둔다.)
3. 기존 스모크를 실 S3 대상으로 재실행 (스크립트는 endpoint/network를 env로 받는다):
   `$env:S3ENDPOINT=''; $env:S3NETWORK=''; powershell -File scripts/smoke/media-e2e.ps1; powershell -File scripts/smoke/binary-e2e.ps1`
4. GCS 검증 시: GCS HMAC 키 발급 → terraform이 Secret Manager에 저장한 값을
   env로 주입하고 동일 스크립트 실행 (`S3_ENDPOINT=https://storage.googleapis.com`).

판정: 두 스모크 전부 `failures=0` + 실 버킷에 객체 존재(`aws s3api head-object`).
주의: LocalStack 컨테이너 재시작 시 버킷이 소멸한다. 로컬로 돌아올 때는
`docker compose up s3-buckets`로 버킷을 먼저 재생성한다.

---

## 차단 3종: Worker 엣지 → 실 백엔드 프록시

- 차단 사유: 공개 백엔드 URL 없음. 현재 `BACKEND_API_URL` 미설정이라 Worker가
  edge 폴백(인메모리 데모 데이터)으로 응답 중이다.
- 투입물: 인터넷에서 도달 가능한 백엔드 base URL 1개 (차단 4종 apply 결과물).
- 코드 위치: `worker/index.ts` (`env.BACKEND_API_URL`이 있으면 전량 리버스 프록시),
  판별 API `GET /healthz/ready` → `backend_proxy_configured` (true/false).

절차:

1. 시크릿 주입 후 재배포:
   `npx wrangler secret put BACKEND_API_URL` (값으로 공개 백엔드 URL 입력)
   `npx wrangler deploy`
2. 프록시 전환 확인:
   `curl.exe -s https://appmarket.dnjstjqppk.workers.dev/healthz/ready`
   → `backend_proxy_configured: true`, `architecture: Hybrid ...` 여야 한다.
3. 실경로 E2E: `$env:APPBASE='https://appmarket.dnjstjqppk.workers.dev'; powershell -File scripts/smoke/core-e2e.ps1`
   (DB 직접 접근 단계는 `-DbExec kubectl` + `-DbUrl` 필요. `scripts/smoke/common.ps1` 참조.)

판정: `/healthz/ready`가 `backend_proxy_configured: true` + core 스모크 `failures=0`.
롤백: `npx wrangler secret delete BACKEND_API_URL; npx wrangler deploy` → edge 폴백 복귀.

---

## 차단 4종: k8s / terraform 클라우드 apply

- 차단 사유: 대상 클러스터·클라우드 자격증명 없음.
  `deploy/k8s/*`는 kubeconform strict 7/7 valid, `deploy/terraform/{aws,gcp}`는
  `fmt+validate` 통과 상태로 대기 중이다.
- 투입물: (택1) AWS 자격증명 + state 버킷/lock 테이블, 또는 GCP 자격증명 + state 버킷,
  또는 기존 k8s 클러스터의 kubeconfig.

절차 A (k8s 직접 apply):

1. `kubectl config use-context <prod-context>; kubectl cluster-info`
2. `kubectl apply -f deploy/k8s/namespace.yaml; kubectl apply -f deploy/k8s/configmap.yaml`
3. `deploy/k8s/secrets.yaml`의 플레이스홀더를 실값 Secret으로 교체 후 apply
   (파일에 실값을 커밋하지 않는다.)
4. `kubectl apply -f deploy/k8s/deployment.yaml; kubectl apply -f deploy/k8s/service-and-ingress.yaml`
5. `kubectl rollout status deployment/bn-appmarket-api -n bn-appmarket`

절차 B (terraform AWS):

1. `Copy-Item deploy/terraform/aws/backend.s3.hcl deploy/terraform/aws/backend.s3.live.hcl`
   후 버킷명·리전·lock 테이블 기입 (`*.live.hcl`은 gitignore됨).
2. `terraform -chdir=deploy/terraform/aws init -backend-config=backend.s3.live.hcl`
3. `terraform -chdir=deploy/terraform/aws plan -out=tfplan` 검토 후 `apply tfplan`
4. outputs (`alb_dns_name`, `*_bucket`)를 차단 2·3종의 투입물로 사용.

절차 C (terraform GCP): `backend.gcs.live.hcl` 복사·기입 후 동일.
알려진 IaC 잔여 갭(apply 전 결정 필요): AWS HTTPS 리스너 없음(HTTP만),
OpenSearch in-cluster 가정, 인메모리 원장×3replica, migrate 잡 없음,
`S3_HEALTH_URL` 실값 교체. 상세는 `HANDOVER_MANIFESTO.md` IaC 절.

판정: rollout 성공 + ALB/ingress 경유 `/healthz/ready` 200 + 차단 3종 절차로 엣지 연결.

---

## 실행 순서 권장

차단 4종(인프라) → 차단 2종(실 S3) → 차단 3종(엣지 연결) → 차단 1종(실결제).
각 단계 판정을 만족해야 다음 단계로 간다. 전체 통과 후 `appmarket-status.md`의
차단 목록을 해제하고 본 런북 상단에 통과 일시를 기록한다.
