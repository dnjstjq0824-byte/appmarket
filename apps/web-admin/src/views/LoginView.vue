<template>
  <div class="admin-login-view">
    <div class="login-card">
      <div class="login-brand">
        <span class="seal-mark" aria-hidden="true">BN</span>
        <span class="logo-text">BookingNetwork</span>
        <span class="badge-admin">ADMIN CONSOLE</span>
      </div>
      <h2>관리자 보안 인증</h2>
      <p class="login-desc">관리자 콘솔 접근을 위해 승인된 직무 역할로 인증해야 합니다. (Zero-Trust IAM Enforcement)</p>

      <form v-if="!mfaStep" @submit.prevent="submitLogin">
        <label class="field">이메일
          <input v-model="email" type="email" required autocomplete="email" />
        </label>
        <label class="field">비밀번호
          <input v-model="password" type="password" required autocomplete="current-password" />
        </label>
        <div v-if="error" class="login-error">{{ error }}</div>
        <button type="submit" class="btn btn-primary" :disabled="busy">
          {{ busy ? '인증 중…' : '관리자 로그인' }}
        </button>
      </form>

      <form v-else-if="!enrollStep" @submit.prevent="submitMfa">
        <p class="login-desc">MFA 인증 코드를 입력하십시오.</p>
        <label class="field">MFA 코드
          <input v-model="mfaCode" type="text" inputmode="numeric" required autocomplete="one-time-code" />
        </label>
        <div v-if="error" class="login-error">{{ error }}</div>
        <button type="submit" class="btn btn-primary" :disabled="busy">
          {{ busy ? '확인 중…' : 'MFA 확인' }}
        </button>
      </form>

      <form v-else @submit.prevent="submitMfa">
        <p class="login-desc">최초 로그인입니다. 인증 앱에 아래 시크릿을 등록한 뒤 코드를 입력하십시오.</p>
        <div v-if="enrollment" class="enroll-box">
          <div class="enroll-secret">{{ enrollment.secret }}</div>
          <div class="enroll-uri">{{ enrollment.provisioning_uri }}</div>
          <p class="login-desc">복구 코드(최초 1회 표시, 안전한 곳에 보관):</p>
          <div class="enroll-codes">{{ enrollment.recovery_codes.join(' ') }}</div>
        </div>
        <label class="field">MFA 코드
          <input v-model="mfaCode" type="text" inputmode="numeric" required autocomplete="one-time-code" />
        </label>
        <div v-if="error" class="login-error">{{ error }}</div>
        <button type="submit" class="btn btn-primary" :disabled="busy">
          {{ busy ? '확인 중…' : '등록 확인' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { getSharedClient, storeSession, clearSession, ApiClientError, type MFAEnrollment } from '@bookingnetwork/api-client';

const api = getSharedClient();
const router = useRouter();
const route = useRoute();

const email = ref('');
const password = ref('');
const mfaCode = ref('');
const mfaUserId = ref('');
const mfaStep = ref(false);
const enrollStep = ref(false);
const enrollment = ref<MFAEnrollment | null>(null);
const error = ref('');
const busy = ref(false);

async function ensureAdminRole(): Promise<boolean> {
  const me = await api.getMe();
  const roles = me.roles || [];
  if (!roles.some(r => r === 'ADMIN_SUPER' || r.startsWith('ADMIN_'))) {
    clearSession();
    error.value = '관리자 권한이 없는 계정입니다.';
    return false;
  }
  return true;
}

function redirectTarget(): string {
  return typeof route.query.redirect === 'string' ? route.query.redirect : '/';
}

async function submitLogin() {
  busy.value = true;
  error.value = '';
  try {
    const res = await api.login(email.value.trim(), password.value);
    if (res.mfa_required) {
      mfaUserId.value = res.user_id || res.user?.id || '';
      if (res.mfa_enrollment_required) {
        const enrolled = await api.enrollMfa(email.value.trim(), password.value);
        enrollment.value = enrolled;
        mfaUserId.value = enrolled.user_id;
        enrollStep.value = true;
      }
      mfaStep.value = true;
      return;
    }
    if (!res.access_token) {
      error.value = '서버가 액세스 토큰을 반환하지 않았습니다.';
      return;
    }
    storeSession(res.access_token, res.refresh_token);
    if (!(await ensureAdminRole())) {
      return;
    }
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
    if (!(await ensureAdminRole())) {
      return;
    }
    await router.push(redirectTarget());
  } catch (err) {
    error.value = err instanceof ApiClientError ? err.message : 'MFA 검증에 실패했습니다.';
  } finally {
    busy.value = false;
  }
}
</script>

<style scoped>
.admin-login-view {
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
.badge-admin {
  font-family: var(--bn-font-mono);
  font-size: 9px;
  letter-spacing: 0.1em;
  color: var(--bn-seal);
  border: 1px solid var(--bn-seal);
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
.enroll-box {
  background: var(--bn-warning-wash);
  border: 1px solid var(--bn-warning-amber);
  border-radius: var(--bn-radius-md);
  padding: 14px;
  margin: 12px 0 16px;
  font-size: 13px;
}
.enroll-secret {
  font-family: var(--bn-font-mono);
  font-size: 16px;
  font-weight: 700;
  letter-spacing: 1px;
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
</style>
