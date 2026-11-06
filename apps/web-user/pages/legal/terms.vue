<template>
  <main class="legal-page">
    <h1>이용약관</h1>
    <p class="legal-meta">시행일: 2026-11-06 · Booking Network App Market</p>

    <section>
      <h2>제1조 (목적)</h2>
      <p>이 약관은 Booking Network App Market(이하 "마켓")이 제공하는 애플리케이션 유통, 결제, 리뷰 및 고객지원 서비스의 이용 조건과 절차, 이용자와 마켓의 권리와 의무를 정합니다.</p>
    </section>

    <section>
      <h2>제2조 (정의)</h2>
      <ul>
        <li>"이용자"란 마켓에 접속하여 서비스를 이용하는 회원과 비회원을 말합니다.</li>
        <li>"개발자"란 마켓에 애플리케이션을 등록·배포하는 사업자를 말합니다.</li>
        <li>"유료 앱"이란 이용자가 대금을 지급하고 이용 권한(Entitlement)을 취득하는 애플리케이션을 말합니다.</li>
        <li>"서명 다운로드"란 위·변조 방지를 위해 만료 시간이 포함된 서명 URL로 제공되는 설치 파일을 말합니다.</li>
      </ul>
    </section>

    <section>
      <h2>제3조 (계정과 보안)</h2>
      <ul>
        <li>이용자는 이메일과 비밀번호로 계정을 만들 수 있으며, 비밀번호는 안전하게 보관해야 합니다.</li>
        <li>관리자 및 개발자 계정은 2단계 인증(MFA) 등록이 필수이며, 최초 로그인 시 인증 앱 등록을 안내합니다.</li>
        <li>타인 명의 도용, 부정 결제, 리뷰 조작 등 부정행위가 확인되면 계정 이용이 제한될 수 있습니다.</li>
      </ul>
    </section>

    <section>
      <h2>제4조 (유료 앱 구매와 청약철회)</h2>
      <ul>
        <li>유료 앱의 가격은 개발자가 국가별로 등록한 정가와 프로모션 할인이 적용된 금액이며, 주문 시 서버가 정가 일치 여부를 검증합니다.</li>
        <li>결제는 PG사를 통해 처리되며, 결제 완료 시 해당 앱의 영구 이용 권한이 이용자 보관함에 발급됩니다.</li>
        <li>전자상거래법 제17조에 따라 구매 후 14일 이내에 미다운로드 상태인 경우 전액 환불(청약철회)이 가능합니다.</li>
        <li>환불이 승인되면 해당 앱의 이용 권한이 즉시 회수됩니다.</li>
      </ul>
    </section>

    <section>
      <h2>제5조 (리뷰와 신고)</h2>
      <ul>
        <li>이용자는 구매·설치한 앱에 별점(1-5)과 리뷰를 작성할 수 있으며, 부정 리뷰 탐지 시스템이 어뷰징을 차단할 수 있습니다.</li>
        <li>악성코드, 지식재산권 침해, 결제 사기 등이 의심되는 앱은 신고하기로 접수할 수 있으며, 접수 시 티켓 ID가 발급됩니다.</li>
      </ul>
    </section>

    <section>
      <h2>제6조 (책임의 한계)</h2>
      <ul>
        <li>마켓은 개발자가 등록한 앱의 기능·품질에 대해 개발자를 대신하여 보증하지 않으며, 앱 자체의 하자는 개발자가 책임집니다.</li>
        <li>마켓은 천재지변, 기간통신사업자 장애 등 불가항력으로 인한 서비스 중단에 책임을 지지 않습니다.</li>
      </ul>
    </section>

    <section>
      <h2>제7조 (준거법과 분쟁 해결)</h2>
      <p>이 약관은 대한민국 법을 준거법으로 하며, 분쟁은 민사소송법상 관할 법원에서 해결합니다.</p>
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
