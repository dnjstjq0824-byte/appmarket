<template>
  <div class="admin-login-view">
    <div class="login-card">
      <div class="login-brand">
        <span class="logo-icon">▲</span>
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
  min-height: calc(100vh - 64px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 32px 16px;
}
.login-card {
  width: 100%;
  max-width: 420px;
  background: #0f172a;
  border: 1px solid #1e293b;
  border-radius: 12px;
  padding: 32px;
}
.login-brand {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 20px;
}
.logo-icon {
  color: #f59e0b;
  font-size: 18px;
}
.logo-text {
  font-weight: 700;
  color: #f8fafc;
}
.badge-admin {
  background: #b45309;
  color: #fff;
  font-size: 10px;
  font-weight: 700;
  padding: 3px 8px;
  border-radius: 4px;
}
.login-card h2 {
  margin: 0 0 4px;
  color: #f8fafc;
  font-size: 20px;
}
.login-desc {
  color: #94a3b8;
  font-size: 13px;
  margin: 0 0 20px;
}
.field {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 12px;
  font-weight: 600;
  color: #94a3b8;
  margin-bottom: 12px;
}
.field input {
  padding: 10px 12px;
  border: 1px solid #334155;
  border-radius: 6px;
  font-size: 14px;
  background: #0b1120;
  color: #f8fafc;
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
}
.btn-primary {
  background: #b45309;
  color: white;
}
.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.enroll-box {
  background: #fffbeb;
  border: 1px solid #fcd34d;
  border-radius: 8px;
  padding: 12px;
  margin: 12px 0;
  font-size: 13px;
}
.enroll-secret {
  font-family: monospace;
  font-size: 16px;
  font-weight: 700;
  letter-spacing: 1px;
  word-break: break-all;
}
.enroll-uri {
  font-family: monospace;
  font-size: 11px;
  color: #92400e;
  word-break: break-all;
  margin-top: 6px;
}
.enroll-codes {
  font-family: monospace;
  font-size: 12px;
  word-break: break-all;
  line-height: 1.7;
}
</style>
