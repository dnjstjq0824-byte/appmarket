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
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';

const form = ref({
  appIdentifier: '',
  category: '',
  reporterEmail: '',
  description: '',
  confirmAccuracy: false
});

const isSubmitting = ref(false);

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

  try {
    const res = await fetch('/api/v1/moderation/reports', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(form.value)
    });

    if (res.ok) {
      alert('Report received. A ticket ID has been generated and sent to your email.');
      form.value = {
        appIdentifier: '',
        category: '',
        reporterEmail: '',
        description: '',
        confirmAccuracy: false
      };
    } else {
      alert('Failed to submit report. Please try again.');
    }
  } catch (err) {
    console.error(err);
  } finally {
    isSubmitting.value = false;
  }
}
</script>

<style scoped>
.report-page-container {
  max-width: 750px;
  margin: 0 auto;
  padding: 32px 16px;
}
.report-header {
  margin-bottom: 24px;
}
.report-header h2 {
  font-size: 22px;
  margin: 0 0 6px 0;
}
.subtitle {
  color: #64748b;
  font-size: 14px;
}
.report-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 24px;
}
.form-group {
  margin-bottom: 20px;
}
.form-group label {
  font-size: 12px;
  text-transform: uppercase;
  color: #64748b;
  font-weight: 600;
  display: block;
  margin-bottom: 6px;
}
.input-control {
  width: 100%;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  padding: 10px;
  font-family: inherit;
  font-size: 14px;
  box-sizing: border-box;
}
.checkbox-group {
  margin-bottom: 24px;
  font-size: 13px;
  color: #475569;
}
.checkbox-group label {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  cursor: pointer;
}
.checkbox-group input {
  margin-top: 3px;
}
.form-actions {
  display: flex;
  justify-content: flex-end;
}
.btn {
  padding: 10px 24px;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  border: none;
}
.btn-primary { background: #2563eb; color: white; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
</style>
