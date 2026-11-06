<template>
  <main class="error-page">
    <div class="error-code">{{ error?.statusCode || 500 }}</div>
    <h1>{{ title }}</h1>
    <p class="error-desc">{{ description }}</p>
    <div class="error-actions">
      <NuxtLink to="/" class="error-btn">홈으로 돌아가기</NuxtLink>
      <button v-if="isServerError" class="error-btn error-btn-secondary" @click="handleError">다시 시도</button>
    </div>
  </main>
</template>

<script setup lang="ts">
import { computed } from 'vue';

const props = defineProps<{ error?: { statusCode?: number; message?: string } }>();

const isServerError = computed(() => (props.error?.statusCode || 500) >= 500);

const title = computed(() => {
  const code = props.error?.statusCode || 500;
  if (code === 404) return '페이지를 찾을 수 없습니다';
  if (code === 403) return '접근 권한이 없습니다';
  return '일시적인 오류가 발생했습니다';
});

const description = computed(() => {
  const code = props.error?.statusCode || 500;
  if (code === 404) return '요청하신 주소가 변경되었거나 삭제되었습니다. 주소를 확인해 주세요.';
  if (code === 403) return '로그인이 필요하거나 권한이 부족합니다.';
  return '서버 처리 중 문제가 발생했습니다. 잠시 후 다시 시도해 주세요.';
});

function handleError() {
  clearError({ redirect: '/' });
}
</script>

<style scoped>
.error-page {
  max-width: 560px;
  margin: 0 auto;
  padding: 96px 24px;
  text-align: center;
}
.error-code {
  font-size: 72px;
  font-weight: 800;
  color: var(--play-accent, #00A870);
  line-height: 1;
}
.error-page h1 {
  font-size: 22px;
  margin: 16px 0 8px;
}
.error-desc {
  font-size: 14px;
  color: var(--play-text-secondary, #9AA0A6);
  margin: 0 0 32px;
}
.error-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
}
.error-btn {
  padding: 10px 20px;
  background: var(--play-accent, #00A870);
  color: #fff;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 600;
  text-decoration: none;
  border: none;
  cursor: pointer;
}
.error-btn-secondary {
  background: transparent;
  border: 1px solid var(--play-border, #3C4043);
  color: inherit;
}
</style>
