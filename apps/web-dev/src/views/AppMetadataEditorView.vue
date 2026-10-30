<template>
  <div class="bn-dev-page">
    <header class="bn-page-header">
      <h2>앱 메타데이터 및 다국어 편집기</h2>
      <p class="bn-text-muted">스토어에 노출될 앱 제목, 설명, 스크린샷 및 배포 카테고리를 설정합니다.</p>
    </header>

    <form class="bn-form-card" @submit.prevent="saveMetadata">
      <div class="bn-lang-tabs">
        <button
          v-for="loc in supportedLocales"
          :key="loc"
          type="button"
          class="bn-tab"
          :class="{ active: currentLocale === loc }"
          @click="currentLocale = loc"
        >
          {{ loc }}
        </button>
      </div>

      <div class="bn-form-group">
        <label>앱 이름 ({{ currentLocale }})</label>
        <input v-model="i18nTitles[currentLocale]" type="text" class="bn-input" required />
      </div>

      <div class="bn-form-group">
        <label>간략 소개 (한 줄 요약)</label>
        <input v-model="i18nShortDescs[currentLocale]" type="text" class="bn-input" />
      </div>

      <div class="bn-form-group">
        <label>상세 설명 (HTML 또는 텍스트)</label>
        <textarea v-model="i18nDescs[currentLocale]" rows="6" class="bn-textarea"></textarea>
      </div>

      <div class="bn-form-group">
        <label>스크린샷 업로드 (최소 2장, 16:9 또는 18:9 규격)</label>
        <div class="bn-screenshot-dropzone">
          <input type="file" multiple accept="image/png,image/jpeg,image/webp" @change="handleScreenshots" />
          <p class="bn-drop-text">권장 규격: 1080x1920 (세로) 또는 1920x1080 (가로)</p>
        </div>
      </div>

      <button type="submit" class="bn-btn bn-btn-primary" :disabled="saving">
        {{ saving ? '저장 중...' : '메타데이터 저장' }}
      </button>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const supportedLocales = ['ko-KR', 'en-US', 'ja-JP', 'ar-SA']
const currentLocale = ref('ko-KR')
const saving = ref(false)

const i18nTitles = ref<Record<string, string>>({
  'ko-KR': '시큐어 메신저',
  'en-US': 'Secure Messenger',
  'ja-JP': 'セキュアメッセンジャー',
  'ar-SA': 'المراسل الآمن'
})

const i18nShortDescs = ref<Record<string, string>>({
  'ko-KR': '안전한 개인정보 보호 메신저',
  'en-US': 'Private encrypted messaging',
  'ja-JP': '安全なプライバシー保護メッセンジャー',
  'ar-SA': 'رسائل مشفرة خاصة'
})

const i18nDescs = ref<Record<string, string>>({
  'ko-KR': '종단간 암호화와 추적 방지를 제공하는 오픈소스 메신저입니다.',
  'en-US': 'Open-source messenger with end-to-end encryption.',
  'ja-JP': 'エンドツーエンド暗号化とトラッキング防止を提供するオープンソースのメッセンジャーです。',
  'ar-SA': 'برنامج مراسلة مفتوح المصدر مع تشفير كامل من طرف إلى طرف.'
})

function handleScreenshots(event: Event) {
  // Screenshot validation logic
}

async function saveMetadata() {
  saving.value = true
  try {
    await new Promise(r => setTimeout(r, 500))
    alert('메타데이터가 성공적으로 저장되었습니다.')
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.bn-dev-page {
  padding: 24px;
  max-width: 900px;
}
.bn-lang-tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 20px;
}
.bn-tab {
  padding: 8px 16px;
  border-radius: 6px;
  border: 1px solid var(--bn-border-subtle, #E2E8F0);
  background-color: var(--bn-bg-elevated, #F1F5F9);
  cursor: pointer;
  font-weight: 500;
}
.bn-tab.active {
  background-color: var(--bn-primary, #2563EB);
  color: white;
  border-color: var(--bn-primary, #2563EB);
}
.bn-form-card {
  background-color: var(--bn-bg-surface, #FFFFFF);
  border: 1px solid var(--bn-border-subtle, #E2E8F0);
  border-radius: 12px;
  padding: 24px;
}
.bn-form-group {
  margin-bottom: 20px;
}
.bn-form-group label {
  display: block;
  font-weight: 600;
  margin-bottom: 6px;
  font-size: 14px;
}
.bn-input, .bn-textarea {
  width: 100%;
  padding: 10px 14px;
  border-radius: 6px;
  border: 1px solid var(--bn-border-subtle, #E2E8F0);
  font-size: 14px;
}
.bn-screenshot-dropzone {
  border: 2px dashed var(--bn-border-subtle, #E2E8F0);
  border-radius: 8px;
  padding: 24px;
  text-align: center;
}
.bn-drop-text {
  font-size: 12px;
  color: var(--bn-text-muted, #64748B);
  margin-top: 8px;
}
.bn-btn-primary {
  background-color: var(--bn-primary, #2563EB);
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 6px;
  font-weight: 600;
  cursor: pointer;
}
</style>

