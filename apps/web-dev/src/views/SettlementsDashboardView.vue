<template>
  <div class="settlements-dashboard-view">
    <div class="view-header">
      <h2>Financial Settlements & Ledger Dashboard</h2>
      <p class="subtitle">15% fair store fee commission, transparent monthly ledger, and payout statements</p>
    </div>

    <!-- Settlement Balance Overview Cards -->
    <div class="balance-grid">
      <div class="balance-card">
        <label>Available for Payout</label>
        <div class="balance-amount font-mono">₩ {{ formatCurrency(financialSummary.availableBalance) }}</div>
        <span class="subtext">Next payout cycle: 15th of next month</span>
      </div>
      <div class="balance-card">
        <label>Gross In-App Revenue (This Month)</label>
        <div class="balance-amount font-mono">₩ {{ formatCurrency(financialSummary.grossRevenue) }}</div>
        <span class="subtext">{{ financialSummary.totalTransactions }} completed transactions</span>
      </div>
      <div class="balance-card">
        <label>Store Commission (15%)</label>
        <div class="balance-amount font-mono text-danger">-₩ {{ formatCurrency(financialSummary.platformFee) }}</div>
        <span class="subtext">Fair transparent platform fee</span>
      </div>
      <div class="balance-card">
        <label>Refunds & Chargebacks</label>
        <div class="balance-amount font-mono text-danger">-₩ {{ formatCurrency(financialSummary.refunds) }}</div>
        <span class="subtext">{{ financialSummary.refundCount }} refund requests</span>
      </div>
    </div>

    <!-- Monthly Statement Ledger -->
    <div class="ledger-card">
      <div class="card-header">
        <h3>Monthly Settlement Statements</h3>
        <button @click="downloadCsvLedger" class="btn btn-secondary btn-sm">Export CSV Ledger</button>
      </div>

      <table class="statement-table">
        <thead>
          <tr>
            <th>Settlement Period</th>
            <th>Gross Sales</th>
            <th>Refunds</th>
            <th>Platform Fee (15%)</th>
            <th>Net Payout</th>
            <th>Status</th>
            <th>Action</th>
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
              <span class="badge" :class="stmt.status === 'PAID' ? 'badge-success' : 'badge-pending'">
                {{ stmt.status }}
              </span>
            </td>
            <td>
              <button class="btn btn-outline btn-xs">View Invoice</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Payout Account Information -->
    <div class="payout-account-card">
      <h3>Registered Payout Bank Account</h3>
      <div class="bank-details">
        <div class="bank-item">
          <label>Bank Name</label>
          <div class="value">Shinhan Bank (신한은행)</div>
        </div>
        <div class="bank-item">
          <label>Account Number</label>
          <div class="value font-mono">110-***-892401</div>
        </div>
        <div class="bank-item">
          <label>Account Holder</label>
          <div class="value">BookingNetwork Corp.</div>
        </div>
        <div class="bank-item">
          <label>KYB Verification</label>
          <div class="value text-success font-bold">✓ Verified</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';

const financialSummary = ref({
  availableBalance: 42500000,
  grossRevenue: 50000000,
  platformFee: 7500000, // Exactly 15%
  refunds: 0,
  totalTransactions: 12450,
  refundCount: 0
});

const statements = ref([
  {
    period: '2026-09',
    gross: 45000000,
    refunds: 500000,
    fee: 6675000,
    net: 37825000,
    status: 'PAID'
  },
  {
    period: '2026-08',
    gross: 38000000,
    refunds: 200000,
    fee: 5670000,
    net: 32130000,
    status: 'PAID'
  },
  {
    period: '2026-10',
    gross: 50000000,
    refunds: 0,
    fee: 7500000,
    net: 42500000,
    status: 'PROCESSING'
  }
]);

function formatCurrency(val: number): string {
  return val.toLocaleString('ko-KR');
}

function downloadCsvLedger() {
  const headers = 'Period,Gross,Refunds,Fee,Net,Status\n';
  const rows = statements.value.map(s => `${s.period},${s.gross},${s.refunds},${s.fee},${s.net},${s.status}`).join('\n');
  const blob = new Blob([headers + rows], { type: 'text/csv' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = 'settlement-ledger-2026.csv';
  a.click();
  URL.revokeObjectURL(url);
}
</script>

<style scoped>
.settlements-dashboard-view {
  padding: 24px;
  max-width: 1000px;
  margin: 0 auto;
}
.view-header {
  margin-bottom: 24px;
}
.subtitle {
  color: #64748b;
  font-size: 14px;
}
.balance-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}
.balance-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 20px;
}
.balance-card label {
  font-size: 12px;
  color: #64748b;
  text-transform: uppercase;
  font-weight: 600;
  display: block;
  margin-bottom: 8px;
}
.balance-amount {
  font-size: 22px;
  font-weight: 700;
  color: #0f172a;
  margin-bottom: 6px;
}
.subtext {
  font-size: 11px;
  color: #94a3b8;
}
.text-danger { color: #dc2626; }
.text-success { color: #16a34a; }
.ledger-card, .payout-account-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 24px;
  margin-bottom: 24px;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.statement-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}
.statement-table th {
  text-align: left;
  padding: 10px 12px;
  border-bottom: 2px solid #e2e8f0;
  color: #475569;
  font-weight: 600;
}
.statement-table td {
  padding: 12px;
  border-bottom: 1px solid #f1f5f9;
}
.badge {
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 600;
}
.badge-success { background: #dcfce7; color: #166534; }
.badge-pending { background: #fef3c7; color: #92400e; }
.bank-details {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-top: 16px;
  background: #f8fafc;
  padding: 16px;
  border-radius: 6px;
}
.bank-item label {
  font-size: 11px;
  color: #64748b;
  text-transform: uppercase;
  display: block;
  margin-bottom: 4px;
}
.bank-item .value {
  font-size: 14px;
}
.btn {
  padding: 6px 14px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border: none;
}
.btn-secondary { background: #f1f5f9; color: #334155; }
.btn-outline { border: 1px solid #cbd5e1; background: transparent; color: #334155; }
.btn-sm { font-size: 12px; padding: 4px 10px; }
.btn-xs { font-size: 11px; padding: 2px 8px; }
</style>

