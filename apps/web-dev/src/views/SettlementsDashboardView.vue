<template>
  <div class="settlements-dashboard-view">
    <div class="view-header">
      <h2>Financial Settlements & Ledger Dashboard</h2>
      <p class="subtitle">15% fair store fee commission, transparent monthly ledger, and payout statements</p>
    </div>

    <div v-if="orgs.length > 0" class="org-row">
      <label for="orgSelect">Organization</label>
      <select id="orgSelect" v-model="activeOrgId" @change="loadAll">
        <option v-for="org in orgs" :key="org.id" :value="org.id">{{ org.org_name }}</option>
      </select>
    </div>
    <div v-if="isLoading" class="loading-row">Loading settlement ledger…</div>
    <div v-if="loadError" class="load-error">{{ loadError }}</div>

    <template v-if="!isLoading && !loadError">
      <!-- Settlement Balance Overview Cards -->
      <div class="balance-grid">
        <div class="balance-card">
          <label>Available for Payout</label>
          <div class="balance-amount font-mono">₩ {{ formatCurrency(summary.payoutTotal) }}</div>
          <span class="subtext">From {{ summary.paidCount }} paid orders</span>
        </div>
        <div class="balance-card">
          <label>Gross Revenue (All Time)</label>
          <div class="balance-amount font-mono">₩ {{ formatCurrency(summary.grossTotal) }}</div>
          <span class="subtext">{{ summary.totalCount }} total orders</span>
        </div>
        <div class="balance-card">
          <label>Store Commission (15%)</label>
          <div class="balance-amount font-mono text-danger">-₩ {{ formatCurrency(summary.commissionTotal) }}</div>
          <span class="subtext">Fair transparent platform fee</span>
        </div>
        <div class="balance-card">
          <label>Refunds</label>
          <div class="balance-amount font-mono text-danger">-₩ {{ formatCurrency(summary.refundTotal) }}</div>
          <span class="subtext">{{ summary.refundCount }} refunded orders</span>
        </div>
      </div>

      <!-- Monthly Statement Ledger -->
      <div class="ledger-card">
        <div class="card-header">
          <h3>Monthly Settlement Statements</h3>
          <button @click="downloadCsvLedger" class="btn btn-secondary btn-sm" :disabled="statements.length === 0">Export CSV Ledger</button>
        </div>

        <div v-if="statements.length === 0" class="empty-row">No settled orders yet.</div>
        <table v-else class="statement-table">
          <thead>
            <tr>
              <th>Settlement Period</th>
              <th>Gross Sales</th>
              <th>Refunds</th>
              <th>Platform Fee (15%)</th>
              <th>Net Payout</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="stmt in statements" :key="stmt.period">
              <td class="font-mono font-bold">{{ stmt.period }}</td>
              <td class="font-mono">₩ {{ formatCurrency(stmt.gross) }}</td>
              <td class="font-mono text-danger">-₩ {{ formatCurrency(stmt.refunds) }}</td>
              <td class="font-mono text-danger">-₩ {{ formatCurrency(stmt.fee) }}</td>
              <td class="font-mono text-success font-bold">₩ {{ formatCurrency(stmt.net) }}</td>
              <td>
                <span class="badge" :class="stmt.status === 'SETTLED' ? 'badge-success' : 'badge-pending'">
                  {{ stmt.status }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Payout Account Information -->
      <div class="payout-account-card">
        <h3>Registered Payout Bank Account</h3>
        <div v-if="accountError" class="load-error">{{ accountError }}</div>
        <div v-if="payoutAccount" class="bank-details">
          <div class="bank-item">
            <label>Bank Name</label>
            <div class="value">{{ payoutAccount.bank_name }}</div>
          </div>
          <div class="bank-item">
            <label>Account Number</label>
            <div class="value font-mono">****-****-{{ payoutAccount.account_last4 }}</div>
          </div>
          <div class="bank-item">
            <label>Account Holder</label>
            <div class="value">{{ payoutAccount.holder_name }}</div>
          </div>
          <div class="bank-item">
            <label>Verification</label>
            <div class="value font-bold" :class="payoutAccount.is_verified ? 'text-success' : 'text-amber'">
              {{ payoutAccount.is_verified ? 'Verified' : 'Pending verification' }}
            </div>
          </div>
        </div>
        <form class="payout-form" @submit.prevent="savePayoutAccount">
          <h4>{{ payoutAccount ? 'Update payout account' : 'Register payout account' }}</h4>
          <div class="form-grid">
            <label>Currency (ISO)
              <input v-model="payoutForm.currency" maxlength="3" placeholder="KRW" required />
            </label>
            <label>Bank name
              <input v-model="payoutForm.bank_name" placeholder="Shinhan Bank" required />
            </label>
            <label>Account holder
              <input v-model="payoutForm.holder_name" placeholder="Acme Inc" required />
            </label>
            <label>Account number
              <input v-model="payoutForm.account_number" inputmode="numeric" placeholder="110-123-892401" required />
            </label>
            <label>Routing / SWIFT code
              <input v-model="payoutForm.routing_code" placeholder="088" />
            </label>
          </div>
          <p class="form-note">Account numbers are AES-256-GCM encrypted at rest. Only the last 4 digits are ever displayed.</p>
          <button type="submit" class="btn btn-primary" :disabled="savingAccount">{{ savingAccount ? 'Saving…' : 'Save payout account' }}</button>
        </form>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { getSharedClient, ApiClientError, type OrgSummary } from '@bookingnetwork/api-client';

const api = getSharedClient();

interface SettlementOrder {
  id: string;
  userId: string;
  appId: string;
  currency: string;
  grossAmount: number;
  commissionAmount: number;
  devPayoutAmount: number;
  status: string;
  createdAt: string;
}

interface MonthlyStatement {
  period: string;
  gross: number;
  refunds: number;
  fee: number;
  net: number;
  status: string;
}

interface PayoutAccountView {
  org_id: string;
  currency: string;
  bank_name: string;
  holder_name: string;
  account_last4: string;
  routing_code: string;
  is_verified: boolean;
}

const orgs = ref<OrgSummary[]>([]);
const activeOrgId = ref('');
const orders = ref<SettlementOrder[]>([]);
const statements = ref<MonthlyStatement[]>([]);
const summary = ref({ grossTotal: 0, payoutTotal: 0, commissionTotal: 0, refundTotal: 0, totalCount: 0, paidCount: 0, refundCount: 0 });
const isLoading = ref(true);
const loadError = ref('');

const payoutAccount = ref<PayoutAccountView | null>(null);
const payoutForm = ref({ currency: 'KRW', bank_name: '', holder_name: '', account_number: '', routing_code: '' });
const accountError = ref('');
const savingAccount = ref(false);

onMounted(async () => {
  try {
    orgs.value = await api.listOrgs();
    if (orgs.value.length > 0) {
      activeOrgId.value = orgs.value[0].id;
      await loadAll();
    }
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load organizations';
  } finally {
    isLoading.value = false;
  }
});

async function loadAll() {
  if (!activeOrgId.value) return;
  isLoading.value = true;
  loadError.value = '';
  try {
    const settlement = await api.getOrgSettlements(activeOrgId.value) as { orders: SettlementOrder[]; grossTotal: number; payoutTotal: number };
    orders.value = settlement.orders || [];
    summary.value.grossTotal = settlement.grossTotal || 0;
    summary.value.payoutTotal = settlement.payoutTotal || 0;
    let commission = 0;
    let refunds = 0;
    let paid = 0;
    let refunded = 0;
    const byMonth = new Map<string, MonthlyStatement>();
    for (const o of orders.value) {
      const month = String(o.createdAt || '').slice(0, 7) || 'unknown';
      let stmt = byMonth.get(month);
      if (!stmt) {
        stmt = { period: month, gross: 0, refunds: 0, fee: 0, net: 0, status: 'ACCRUING' };
        byMonth.set(month, stmt);
      }
      if (o.status === 'REFUNDED') {
        refunded += 1;
        refunds += o.grossAmount;
        stmt.refunds += o.grossAmount;
      } else {
        if (o.status === 'PAID') {
          paid += 1;
        }
        commission += o.commissionAmount;
        stmt.gross += o.grossAmount;
        stmt.fee += o.commissionAmount;
        stmt.net += o.devPayoutAmount;
      }
    }
    summary.value.commissionTotal = commission;
    summary.value.refundTotal = refunds;
    summary.value.totalCount = orders.value.length;
    summary.value.paidCount = paid;
    summary.value.refundCount = refunded;
    const currentMonth = new Date().toISOString().slice(0, 7);
    statements.value = [...byMonth.values()]
      .map(s => ({ ...s, status: s.period < currentMonth ? 'SETTLED' : 'ACCRUING' }))
      .sort((a, b) => (a.period < b.period ? 1 : -1));
    await loadPayoutAccount();
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load settlements';
  } finally {
    isLoading.value = false;
  }
}

async function loadPayoutAccount() {
  payoutAccount.value = null;
  accountError.value = '';
  try {
    payoutAccount.value = await api.getPayoutAccount(activeOrgId.value) as PayoutAccountView;
  } catch (err) {
    if (err instanceof ApiClientError && err.status === 404) {
      return;
    }
    accountError.value = err instanceof ApiClientError ? err.message : 'Failed to load payout account';
  }
}

async function savePayoutAccount() {
  savingAccount.value = true;
  accountError.value = '';
  try {
    payoutAccount.value = await api.savePayoutAccount(activeOrgId.value, {
      currency: payoutForm.value.currency.trim().toUpperCase(),
      bank_name: payoutForm.value.bank_name.trim(),
      holder_name: payoutForm.value.holder_name.trim(),
      account_number: payoutForm.value.account_number.trim(),
      routing_code: payoutForm.value.routing_code.trim()
    }) as PayoutAccountView;
    payoutForm.value.account_number = '';
  } catch (err) {
    accountError.value = err instanceof ApiClientError ? err.message : 'Failed to save payout account';
  } finally {
    savingAccount.value = false;
  }
}

function formatCurrency(val: number): string {
  return Number(val || 0).toLocaleString('ko-KR');
}

function downloadCsvLedger() {
  const headers = 'Period,Gross,Refunds,Fee,Net,Status\n';
  const rows = statements.value.map(s => `${s.period},${s.gross},${s.refunds},${s.fee},${s.net},${s.status}`).join('\n');
  const blob = new Blob([headers + rows], { type: 'text/csv' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = 'settlement-ledger.csv';
  a.click();
  URL.revokeObjectURL(url);
}
</script>

<style scoped>
.settlements-dashboard-view {
  max-width: 1020px;
  margin: 0 auto;
}
.view-header {
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 2px solid var(--bn-ink);
}
.view-header h2 {
  font-family: var(--bn-font-display);
  font-size: 24px;
  font-weight: 900;
  letter-spacing: -0.02em;
  margin: 0 0 4px;
}
.subtitle {
  color: var(--bn-text-secondary);
  font-size: 13px;
  margin: 0;
}
.org-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
  font-size: 13px;
  color: var(--bn-text-secondary);
}
.org-row label {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--bn-text-muted);
}
.org-row select {
  padding: 8px 12px;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  font-size: 13px;
  font-family: var(--bn-font-sans);
  background: var(--bn-bg-surface);
  color: var(--bn-text-primary);
}
.loading-row {
  padding: 24px;
  color: var(--bn-text-muted);
  font-size: 14px;
  text-align: center;
}
.load-error {
  padding: 12px 16px;
  background: var(--bn-danger-wash);
  border: 1px solid var(--bn-danger-red);
  color: var(--bn-danger-red);
  border-radius: var(--bn-radius-sm);
  font-size: 13px;
  margin-bottom: 16px;
}
.empty-row {
  padding: 20px 16px;
  color: var(--bn-text-muted);
  font-size: 14px;
  text-align: center;
}
.balance-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  margin-bottom: 24px;
}
.balance-card {
  background: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 18px;
}
.balance-card label {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--bn-text-muted);
  display: block;
  margin-bottom: 8px;
}
.balance-amount {
  font-family: var(--bn-font-display);
  font-size: 24px;
  font-weight: 900;
  letter-spacing: -0.02em;
  margin-bottom: 6px;
}
.balance-amount.font-mono {
  font-family: var(--bn-font-mono);
  font-size: 18px;
  font-weight: 600;
  letter-spacing: 0;
}
.subtext {
  font-size: 11px;
  color: var(--bn-text-muted);
}
.font-bold { font-weight: 700; }
.font-mono { font-family: var(--bn-font-mono); }
.text-danger { color: var(--bn-danger-red); }
.text-success { color: var(--bn-trust-verified); }
.text-amber { color: var(--bn-warning-amber); }
.ledger-card, .payout-account-card {
  background: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 24px;
  margin-bottom: 24px;
}
.ledger-card h3, .payout-account-card h3 {
  font-family: var(--bn-font-display);
  font-size: 15px;
  font-weight: 800;
  margin: 0;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}
.statement-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}
.statement-table th {
  text-align: left;
  padding: 10px 12px;
  border-bottom: 2px solid var(--bn-ink);
  font-family: var(--bn-font-mono);
  font-weight: 400;
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--bn-text-muted);
  white-space: nowrap;
}
.statement-table td {
  padding: 12px;
  border-bottom: 1px solid var(--bn-border-subtle);
  vertical-align: middle;
}
.statement-table tr:last-child td {
  border-bottom: none;
}
.badge {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.04em;
  padding: 3px 9px;
  border-radius: var(--bn-radius-sm);
  border: 1px solid var(--bn-border-strong);
  color: var(--bn-text-secondary);
  white-space: nowrap;
}
.badge-success { border-color: var(--bn-trust-verified); color: var(--bn-trust-verified); }
.badge-pending { border-color: var(--bn-warning-amber); color: var(--bn-warning-amber); }
.bank-details {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-top: 16px;
  background: var(--bn-bg-well);
  border: 1px solid var(--bn-border-subtle);
  padding: 16px;
  border-radius: var(--bn-radius-sm);
}
.bank-item label {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--bn-text-muted);
  display: block;
  margin-bottom: 4px;
}
.bank-item .value {
  font-size: 14px;
}
.payout-form {
  margin-top: 20px;
  border-top: 1px solid var(--bn-border-subtle);
  padding-top: 16px;
}
.payout-form h4 {
  font-size: 14px;
  margin: 0 0 12px;
}
.form-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  margin-bottom: 12px;
}
.form-grid label {
  display: flex;
  flex-direction: column;
  gap: 6px;
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.06em;
  color: var(--bn-text-muted);
}
.form-grid input {
  padding: 9px 12px;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  background: var(--bn-bg-surface);
  color: var(--bn-text-primary);
  font-family: var(--bn-font-sans);
  font-size: 13px;
}
.form-grid input:focus {
  outline: none;
  border-color: var(--bn-ink);
}
.form-note {
  font-size: 12px;
  color: var(--bn-text-muted);
  margin: 0 0 14px;
}
.form-note code {
  font-family: var(--bn-font-mono);
}
.btn {
  padding: 9px 18px;
  border-radius: var(--bn-radius-sm);
  font-size: 13px;
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
.btn-primary:disabled { opacity: 0.45; cursor: not-allowed; }
.btn-secondary {
  background: transparent;
  border: 1px solid var(--bn-border-subtle);
  color: var(--bn-text-secondary);
}
.btn-secondary:hover:not(:disabled) {
  border-color: var(--bn-border-strong);
  color: var(--bn-text-primary);
}
.btn-secondary:disabled { opacity: 0.45; cursor: not-allowed; }
.btn-outline { border: 1px solid var(--bn-border-strong); background: transparent; color: var(--bn-text-primary); }
.btn-sm { font-size: 12px; padding: 6px 14px; }
.btn-xs { font-size: 11px; padding: 3px 10px; }
@media (max-width: 900px) {
  .balance-grid, .bank-details {
    grid-template-columns: repeat(2, 1fr);
  }
}
@media (max-width: 560px) {
  .balance-grid, .bank-details, .form-grid {
    grid-template-columns: 1fr;
  }
}
</style>
