<template>
  <div class="login-page">
    <div class="login-card">
      <h2>{{ mode === 'login' ? '로그인' : '회원가입' }}</h2>
      <p class="login-desc">리뷰 작성·구매·보관함을 이용하려면 인증이 필요합니다.</p>

      <div v-if="providers.length > 0" class="oauth-row">
        <a v-for="p in providers" :key="p.id" class="btn btn-oauth" :href="oauthUrl(p.id)">{{ p.name }}로 계속하기</a>
      </div>

      <form v-if="!mfaStep" @submit.prevent="submit">
        <label v-if="mode === 'register'" class="field">표시 이름
          <input v-model="displayName" type="text" required autocomplete="name" />
        </label>
        <label class="field">이메일
          <input v-model="email" type="email" required autocomplete="email" />
        </label>
        <label class="field">비밀번호 (8자 이상)
          <input v-model="password" type="password" required minlength="8" autocomplete="current-password" />
        </label>
        <div v-if="error" class="login-error">{{ error }}</div>
        <button type="submit" class="btn btn-primary" :disabled="busy">
          {{ busy ? '처리 중…' : mode === 'login' ? '로그인' : '회원가입' }}
        </button>
      </form>

      <form v-else @submit.prevent="submitMfa">
        <p class="login-desc">MFA 인증 코드를 입력하십시오.</p>
        <div v-if="enrollment" class="enroll-box">
          <p class="login-desc">최초 로그인입니다. 인증 앱에 아래 시크릿을 등록하십시오.</p>
          <div class="enroll-secret">{{ enrollment.secret }}</div>
          <div class="enroll-uri">{{ enrollment.provisioning_uri }}</div>
          <p class="login-desc">복구 코드(최초 1회 표시):</p>
          <div class="enroll-codes">{{ enrollment.recovery_codes.join(' ') }}</div>
        </div>
        <label class="field">MFA 코드
          <input v-model="mfaCode" type="text" inputmode="numeric" required autocomplete="one-time-code" />
        </label>
        <div v-if="error" class="login-error">{{ error }}</div>
        <button type="submit" class="btn btn-primary" :disabled="busy">
          {{ busy ? '확인 중…' : enrollment ? '등록 확인' : 'MFA 확인' }}
        </button>
      </form>

      <button v-if="!mfaStep" class="link-btn" @click="toggleMode">
        {{ mode === 'login' ? '계정이 없으신가요? 회원가입' : '이미 계정이 있으신가요? 로그인' }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { getSharedClient, storeSession, ApiClientError, type MFAEnrollment } from '@bookingnetwork/api-client';

const api = getSharedClient();
const router = useRouter();
const route = useRoute();

const mode = ref<'login' | 'register'>('login');
const email = ref('');
const password = ref('');
const displayName = ref('');
const error = ref('');
const busy = ref(false);
const providers = ref<Array<{ id: string; name: string }>>([]);
const mfaStep = ref(false);
const mfaCode = ref('');
const mfaUserId = ref('');
const enrollment = ref<MFAEnrollment | null>(null);

onMounted(async () => {
  try {
    providers.value = await api.listOAuthProviders();
  } catch {
    providers.value = [];
  }
});

function oauthUrl(provider: string): string {
  return api.oauthLoginUrl(provider);
}

function toggleMode() {
  mode.value = mode.value === 'login' ? 'register' : 'login';
  error.value = '';
}

function redirectTarget(): string {
  return typeof route.query.redirect === 'string' ? route.query.redirect : '/';
}

async function submit() {
  busy.value = true;
  error.value = '';
  try {
    const res = mode.value === 'login'
      ? await api.login(email.value.trim(), password.value)
      : await api.register(email.value.trim(), password.value, displayName.value.trim());
    if (res.mfa_required) {
      mfaUserId.value = res.user_id || res.user?.id || '';
      if (res.mfa_enrollment_required) {
        const enrolled = await api.enrollMfa(email.value.trim(), password.value);
        enrollment.value = enrolled;
        mfaUserId.value = enrolled.user_id;
      }
      mfaStep.value = true;
      return;
    }
    if (!res.access_token) {
      error.value = '서버가 액세스 토큰을 반환하지 않았습니다.';
      return;
    }
    storeSession(res.access_token, res.refresh_token);
    await router.push(redirectTarget());
  } catch (err) {
    error.value = err instanceof ApiClientError ? err.message : '인증에 실패했습니다.';
  } finally {
    busy.value = false;
  }
}

async function submitMfa() {
  busy.value = true;
  error.value = '';
  try {
    const res = await api.verifyMfa(mfaUserId.value, mfaCode.value.trim());
    if (!res.access_token) {
      error.value = 'MFA 검증에 실패했습니다.';
      return;
    }
    storeSession(res.access_token, res.refresh_token);
    await router.push(redirectTarget());
  } catch (err) {
    error.value = err instanceof ApiClientError ? err.message : 'MFA 검증에 실패했습니다.';
  } finally {
    busy.value = false;
  }
}
</script>

<style scoped>
.login-page {
  max-width: 480px;
  margin: 0 auto;
  padding: 48px 24px 60px;
}
.login-card {
  background: var(--play-card-bg, #28292A);
  border: 1px solid var(--play-border, #3C4043);
  border-radius: 12px;
  padding: 32px;
}
.login-card h2 {
  margin: 0 0 4px;
  font-size: 20px;
}
.login-desc {
  color: var(--play-text-secondary, #9AA0A6);
  font-size: 13px;
  margin: 0 0 16px;
}
.oauth-row {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 16px;
}
.field {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 12px;
  font-weight: 600;
  margin-bottom: 12px;
}
.field input {
  padding: 10px 12px;
  border: 1px solid var(--play-border, #3C4043);
  border-radius: 6px;
  font-size: 14px;
  background: var(--play-subtle-bg, #131314);
  color: inherit;
}
.login-error {
  padding: 10px 12px;
  background: rgba(220, 38, 38, 0.12);
  border: 1px solid #7f1d1d;
  color: #fca5a5;
  border-radius: 6px;
  font-size: 13px;
  margin-bottom: 12px;
}
.btn {
  width: 100%;
  padding: 10px 16px;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  border: none;
  text-align: center;
  text-decoration: none;
  display: block;
}
.btn-primary {
  background: var(--play-accent, #00A870);
  color: white;
}
.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.btn-oauth {
  background: var(--play-subtle-bg, #131314);
  border: 1px solid var(--play-border, #3C4043);
  font-size: 13px;
}
.enroll-box {
  border: 1px solid #f59e0b;
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 12px;
}
.enroll-secret {
  font-family: monospace;
  font-size: 16px;
  font-weight: 700;
  letter-spacing: 1px;
  color: #fbbf24;
  word-break: break-all;
}
.enroll-uri {
  font-family: monospace;
  font-size: 11px;
  color: #94a3b8;
  word-break: break-all;
  margin-top: 6px;
}
.enroll-codes {
  font-family: monospace;
  font-size: 12px;
  word-break: break-all;
  line-height: 1.7;
}
.link-btn {
  margin-top: 16px;
  background: none;
  border: none;
  color: var(--play-accent, #00A870);
  font-size: 13px;
  cursor: pointer;
  width: 100%;
  text-align: center;
}
</style>
