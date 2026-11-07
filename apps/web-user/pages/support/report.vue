<template>
  <div class="report-page-container">
    <div class="report-header">
      <h2>Report Abuse, Malware or Intellectual Property Violation (USR-20)</h2>
      <p class="subtitle">Submit security concerns, malicious behavior alerts, or legal DMCA takedown requests</p>
    </div>

    <div class="report-card">
      <form @submit.prevent="submitReport">
        <div class="form-group">
          <label>Target Application Package Name or URL</label>
          <input 
            type="text" 
            v-model="form.appIdentifier" 
            placeholder="e.g. com.example.app or app link"
            class="input-control" 
            required 
          />
        </div>

        <div class="form-group">
          <label>Violation Category</label>
          <select v-model="form.category" class="input-control" required>
            <option value="">-- Select Report Reason --</option>
            <option value="MALWARE_OR_PHISHING">Malware, Spyware, or Phishing Payload</option>
            <option value="DMCA_COPYRIGHT">DMCA / Copyright Infringement</option>
            <option value="TRADEMARK_IMPERSONATION">Brand / Trademark Impersonation</option>
            <option value="PRIVACY_VIOLATION">Undisclosed Privacy Violation / Sensitive Data Leak</option>
            <option value="FRAUDULENT_BILLING">Deceptive In-App Purchase / Billing Scam</option>
            <option value="HARMFUL_CONTENT">Inappropriate or Prohibited Content</option>
          </select>
        </div>

        <div class="form-group">
          <label>Your Contact Email (for legal correspondence)</label>
          <input 
            type="email" 
            v-model="form.reporterEmail" 
            placeholder="reporter@domain.com"
            class="input-control" 
            required 
          />
        </div>

        <div class="form-group">
          <label>Detailed Explanation & Evidence</label>
          <textarea 
            v-model="form.description" 
            rows="5" 
            placeholder="Provide specific details, reproduction steps, virus scan links, or copyrighted work registration numbers..."
            class="input-control" 
            required
          ></textarea>
        </div>

        <div class="checkbox-group">
          <label>
            <input type="checkbox" v-model="form.confirmAccuracy" required />
            <span>I certify under penalty of perjury that the information in this report is accurate and that I am authorized to submit it.</span>
          </label>
        </div>

        <div class="form-actions">
          <button
            type="submit"
            class="btn btn-primary"
            :disabled="!isFormValid || isSubmitting"
          >
            {{ isSubmitting ? 'Submitting Report...' : 'Submit Report for Triage' }}
          </button>
        </div>

        <div v-if="submitError" class="form-error">{{ submitError }}</div>
        <div v-if="ticketId" class="form-success">
          신고가 접수되었습니다. 티켓 ID: <strong>{{ ticketId }}</strong> (우선순위: {{ ticketPriority }})
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { getSharedClient, ApiClientError } from '@bookingnetwork/api-client';

const api = getSharedClient();

const form = ref({
  appIdentifier: '',
  category: '',
  reporterEmail: '',
  description: '',
  confirmAccuracy: false
});

const isSubmitting = ref(false);
const submitError = ref('');
const ticketId = ref('');
const ticketPriority = ref('');

const isFormValid = computed(() => {
  return form.value.appIdentifier.trim().length > 3 &&
         form.value.category !== '' &&
         form.value.reporterEmail.includes('@') &&
         form.value.description.trim().length >= 20 &&
         form.value.confirmAccuracy;
});

async function submitReport() {
  if (!isFormValid.value) return;
  isSubmitting.value = true;
  submitError.value = '';
  ticketId.value = '';

  try {
    const res = await api.fileReport({ ...form.value });
    ticketId.value = res.ticketId;
    ticketPriority.value = res.priority;
    form.value = {
      appIdentifier: '',
      category: '',
      reporterEmail: '',
      description: '',
      confirmAccuracy: false
    };
  } catch (err) {
    submitError.value = err instanceof ApiClientError ? err.message : 'Failed to submit report. Please try again.';
  } finally {
    isSubmitting.value = false;
  }
}
</script>

<style scoped>
.report-page-container {
  max-width: 760px;
  margin: 0 auto;
  padding: 32px 24px 60px;
}
.report-header {
  margin-bottom: 24px;
  padding-bottom: 20px;
  border-bottom: 2px solid var(--bn-ink);
}
.report-header h2 {
  font-family: var(--bn-font-display);
  font-size: 26px;
  font-weight: 900;
  letter-spacing: -0.02em;
  margin: 0 0 6px 0;
  line-height: 1.25;
}
.subtitle {
  color: var(--bn-text-secondary);
  font-size: 13px;
  margin: 0;
}
.report-card {
  background: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 28px;
}
.form-group {
  margin-bottom: 20px;
}
.form-group label {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--bn-text-muted);
  display: block;
  margin-bottom: 8px;
}
.input-control {
  width: 100%;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  background: var(--bn-bg-surface);
  color: var(--bn-text-primary);
  padding: 11px 12px;
  font-family: var(--bn-font-sans);
  font-size: 14px;
  box-sizing: border-box;
}
.input-control:focus {
  outline: none;
  border-color: var(--bn-ink);
}
textarea.input-control {
  resize: vertical;
  line-height: 1.6;
}
.checkbox-group {
  margin-bottom: 24px;
  font-size: 13px;
  color: var(--bn-text-secondary);
}
.checkbox-group label {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  cursor: pointer;
  line-height: 1.6;
}
.checkbox-group input {
  margin-top: 3px;
  accent-color: var(--bn-seal);
}
.form-actions {
  display: flex;
  justify-content: flex-end;
}
.btn {
  padding: 12px 28px;
  border-radius: var(--bn-radius-sm);
  font-size: 14px;
  font-weight: 700;
  cursor: pointer;
  border: none;
}
.btn-primary { background: var(--bn-action-bg); color: var(--bn-action-text); }
.btn-primary:hover:not(:disabled) { background: var(--bn-action-bg-hover); }
.btn-primary:disabled { opacity: 0.45; cursor: not-allowed; }
.form-error {
  color: var(--bn-danger-red);
  font-size: 13px;
  margin-top: 12px;
  padding: 10px 12px;
  background: var(--bn-danger-wash);
  border: 1px solid var(--bn-danger-red);
  border-radius: var(--bn-radius-sm);
}
.form-success {
  color: var(--bn-trust-verified);
  font-size: 13px;
  margin-top: 12px;
  padding: 10px 12px;
  background: var(--bn-trust-wash);
  border: 1px solid var(--bn-trust-verified);
  border-radius: var(--bn-radius-sm);
}
.form-success strong {
  font-family: var(--bn-font-mono);
}
</style>

