<template>
  <div class="login-view">
    <div class="login-card">
      <div class="login-brand">
        <span class="seal-mark" aria-hidden="true">BN</span>
        <span class="logo-text">BookingNetwork</span>
        <span class="badge-dev">DEVELOPER CONSOLE</span>
      </div>
      <h2>{{ mode === 'login' ? '개발자 로그인' : '개발자 회원가입' }}</h2>
      <p class="login-desc">조직 앱·릴리스·정산을 관리하려면 인증이 필요합니다.</p>

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

      <button class="link-btn" @click="toggleMode">
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
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/';
    await router.push(redirect);
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
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/';
    await router.push(redirect);
  } catch (err) {
    error.value = err instanceof ApiClientError ? err.message : 'MFA 검증에 실패했습니다.';
  } finally {
    busy.value = false;
  }
}
</script>

<style scoped>
.login-view {
  min-height: 60vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 32px 16px;
}
.login-card {
  width: 100%;
  max-width: 430px;
  background: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-top: 3px solid var(--bn-seal);
  border-radius: var(--bn-radius-md);
  padding: 32px;
}
.login-brand {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 22px;
}
.seal-mark {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  background-color: var(--bn-seal);
  color: #fff;
  font-family: var(--bn-font-display);
  font-weight: 900;
  font-size: 11px;
  border-radius: var(--bn-radius-sm);
}
.logo-text {
  font-family: var(--bn-font-display);
  font-weight: 800;
  font-size: 15px;
}
.badge-dev {
  font-family: var(--bn-font-mono);
  font-size: 9px;
  letter-spacing: 0.1em;
  color: var(--bn-text-muted);
  border: 1px solid var(--bn-border-subtle);
  padding: 3px 7px;
  border-radius: var(--bn-radius-sm);
}
.login-card h2 {
  margin: 0 0 4px;
  font-family: var(--bn-font-display);
  font-size: 22px;
  font-weight: 900;
  letter-spacing: -0.02em;
}
.login-desc {
  color: var(--bn-text-secondary);
  font-size: 13px;
  line-height: 1.6;
  margin: 0 0 18px;
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
  gap: 6px;
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--bn-text-muted);
  margin-bottom: 14px;
}
.field input {
  padding: 11px 12px;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  font-family: var(--bn-font-sans);
  font-size: 14px;
  background: var(--bn-bg-surface);
  color: var(--bn-text-primary);
}
.field input:focus {
  outline: none;
  border-color: var(--bn-ink);
}
.login-error {
  padding: 10px 12px;
  background: var(--bn-danger-wash);
  border: 1px solid var(--bn-danger-red);
  color: var(--bn-danger-red);
  border-radius: var(--bn-radius-sm);
  font-size: 13px;
  margin-bottom: 12px;
}
.btn {
  width: 100%;
  box-sizing: border-box;
  padding: 12px 16px;
  border-radius: var(--bn-radius-sm);
  font-size: 14px;
  font-weight: 700;
  cursor: pointer;
  text-align: center;
  text-decoration: none;
  display: block;
}
.btn-primary {
  background: var(--bn-action-bg);
  color: var(--bn-action-text);
  border: none;
}
.btn-primary:hover:not(:disabled) {
  background: var(--bn-action-bg-hover);
}
.btn-primary:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}
.btn-oauth {
  background: transparent;
  color: var(--bn-text-primary);
  border: 1px solid var(--bn-border-subtle);
  font-size: 13px;
}
.btn-oauth:hover {
  border-color: var(--bn-border-strong);
}
.enroll-box {
  background: var(--bn-warning-wash);
  border: 1px solid var(--bn-warning-amber);
  border-radius: var(--bn-radius-md);
  padding: 14px;
  margin-bottom: 14px;
}
.enroll-secret {
  font-family: var(--bn-font-mono);
  font-size: 16px;
  font-weight: 700;
  letter-spacing: 1px;
  color: var(--bn-text-primary);
  word-break: break-all;
}
.enroll-uri {
  font-family: var(--bn-font-mono);
  font-size: 11px;
  color: var(--bn-text-muted);
  word-break: break-all;
  margin-top: 6px;
}
.enroll-codes {
  font-family: var(--bn-font-mono);
  font-size: 12px;
  word-break: break-all;
  line-height: 1.7;
}
.link-btn {
  margin-top: 16px;
  background: none;
  border: none;
  color: var(--bn-text-primary);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  width: 100%;
  text-align: center;
}
.link-btn:hover {
  text-decoration: underline;
  text-underline-offset: 3px;
}
</style>
