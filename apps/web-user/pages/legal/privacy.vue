<template>
  <main class="legal-page">
    <h1>개인정보처리방침</h1>
    <p class="legal-meta">시행일: 2026-11-06 · Booking Network App Market</p>

    <section>
      <h2>1. 수집하는 개인정보 항목</h2>
      <ul>
        <li>회원가입: 이메일, 비밀번호(암호화 저장), 표시 이름</li>
        <li>본인 인증: 2단계 인증(TOTP) 등록 정보, 복구 코드(해시 저장)</li>
        <li>주문·결제: 주문 내역, 결제 금액, PG 결제 식별자 (카드 번호는 저장하지 않으며 PG사가 처리합니다)</li>
        <li>서비스 이용 기록: 앱 조회·다운로드·결제 이벤트(텔레메트리), 접속 로그, 기기 식별자</li>
      </ul>
    </section>

    <section>
      <h2>2. 이용 목적</h2>
      <ul>
        <li>회원 식별, 부정 이용 방지, 이중 승인 등 보안 통제</li>
        <li>주문 처리, 이용 권한 발급, 환불·고객지원 대응</li>
        <li>전환율·매출 통계 등 서비스 품질 개선 (집계 분석)</li>
      </ul>
    </section>

    <section>
      <h2>3. 보유와 파기</h2>
      <ul>
        <li>회원 정보는 탈퇴 시 지체 없이 파기합니다.</li>
        <li>주문·결제 기록은 전자상거래법(5년), 부가가치세법(5년) 등 관계 법령이 정한 기간 동안 보관합니다.</li>
        <li>감사 로그는 보안 사고 추적을 위해 해시 체인 형태로 보관되며 위·변조가 탐지됩니다.</li>
      </ul>
    </section>

    <section>
      <h2>4. 제3자 제공과 처리 위탁</h2>
      <ul>
        <li>결제 처리를 위해 PG사(토스페이먼츠 등)에 주문 번호와 결제 금액을 제공합니다.</li>
        <li>법령에 따른 수사·과세 기관의 적법한 요청이 있는 경우를 제외하고 제3자에게 제공하지 않습니다.</li>
      </ul>
    </section>

    <section>
      <h2>5. 이용자의 권리</h2>
      <p>이용자는 언제든 개인정보 열람·정정·삭제와 처리 정지를 요구할 수 있으며, 탈퇴 시 계정 정보가 파기됩니다. 문의: privacy@appmarket.local</p>
    </section>

    <section>
      <h2>6. 안전성 확보 조치</h2>
      <ul>
        <li>비밀번호 Argon2id 해시, MFA TOTP, JWT 단기 토큰과 리프레시 토큰 회전</li>
        <li>정산 계좌 AES-256-GCM 암호화 저장, 마스킹 조회</li>
        <li>관리자 작업 이중 승인(Dual-Control)과 감사 로그 해시 체인</li>
      </ul>
    </section>
  </main>
</template>

<style scoped>
.legal-page {
  max-width: 800px;
  margin: 0 auto;
  padding: 48px 24px 80px;
  line-height: 1.7;
}
.legal-page h1 {
  font-size: 26px;
  margin: 0 0 4px;
}
.legal-meta {
  color: var(--play-text-secondary, #9AA0A6);
  font-size: 13px;
  margin: 0 0 32px;
}
.legal-page section {
  margin-bottom: 28px;
}
.legal-page h2 {
  font-size: 17px;
  margin: 0 0 10px;
}
.legal-page p, .legal-page li {
  font-size: 14px;
  color: var(--play-text, #E8EAED);
}
.legal-page ul {
  padding-left: 20px;
  margin: 0;
}
.legal-page li + li {
  margin-top: 6px;
}
</style>
