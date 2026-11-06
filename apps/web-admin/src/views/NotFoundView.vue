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
  font-size: 64px;
  font-weight: 800;
  color: #b45309;
  line-height: 1;
}
.not-found h2 {
  font-size: 20px;
  margin: 16px 0 8px;
}
.not-found p {
  font-size: 14px;
  color: #64748b;
  margin: 0 0 24px;
}
.home-link {
  display: inline-block;
  padding: 10px 20px;
  background: #b45309;
  color: #fff;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 600;
  text-decoration: none;
}
</style>
