<template>
  <div class="not-found">
    <div class="code">404</div>
    <h2>페이지를 찾을 수 없습니다</h2>
    <p>요청하신 주소가 변경되었거나 삭제되었습니다.</p>
    <RouterLink to="/" class="home-link">대시보드로 돌아가기</RouterLink>
  </div>
</template>

<style scoped>
.not-found {
  max-width: 480px;
  margin: 0 auto;
  padding: 96px 24px;
  text-align: center;
}
.code {
  font-family: var(--bn-font-display);
  font-size: 72px;
  font-weight: 900;
  letter-spacing: -0.04em;
  line-height: 1;
}
.not-found h2 {
  font-family: var(--bn-font-display);
  font-size: 19px;
  font-weight: 800;
  margin: 16px 0 8px;
}
.not-found p {
  font-size: 14px;
  color: var(--bn-text-secondary);
  margin: 0 0 24px;
}
.home-link {
  display: inline-block;
  padding: 11px 24px;
  background: var(--bn-action-bg);
  color: var(--bn-action-text);
  border-radius: var(--bn-radius-sm);
  font-size: 14px;
  font-weight: 700;
  text-decoration: none;
}
.home-link:hover {
  background: var(--bn-action-bg-hover);
}
</style>
