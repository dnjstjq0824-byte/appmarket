import { createApp } from 'vue';
import { createRouter, createWebHistory } from 'vue-router';
import App from './App.vue';

import AdminOverviewDashboard from './views/AdminOverviewDashboard.vue';
import AdminRbacDualControlView from './views/AdminRbacDualControlView.vue';
import ApkSecurityReviewView from './views/ApkSecurityReviewView.vue';
import AppMetadataReviewView from './views/AppMetadataReviewView.vue';
import AuditLogChainExplorerView from './views/AuditLogChainExplorerView.vue';
import EmergencyTakedownView from './views/EmergencyTakedownView.vue';
import ExternalConnectorsAdminView from './views/ExternalConnectorsAdminView.vue';
import ModerationReportsQueueView from './views/ModerationReportsQueueView.vue';
import AdCampaignReviewView from './views/AdCampaignReviewView.vue';
import OrdersRefundsAdminView from './views/OrdersRefundsAdminView.vue';
import PlatformAnalyticsAdminView from './views/PlatformAnalyticsAdminView.vue';
import SettlementPayoutAdminView from './views/SettlementPayoutAdminView.vue';

const router = createRouter({
  history: createWebHistory('/admin/'),
  routes: [
    { path: '/', component: AdminOverviewDashboard },
    { path: '/rbac', component: AdminRbacDualControlView },
    { path: '/security/reviews', component: ApkSecurityReviewView },
    { path: '/apps/metadata', component: AppMetadataReviewView },
    { path: '/audit', component: AuditLogChainExplorerView },
    { path: '/security/kill-switch', component: EmergencyTakedownView },
    { path: '/connectors', component: ExternalConnectorsAdminView },
    { path: '/moderation', component: ModerationReportsQueueView },
    { path: '/ads', component: AdCampaignReviewView },
    { path: '/finance/orders', component: OrdersRefundsAdminView },
    { path: '/analytics', component: PlatformAnalyticsAdminView },
    { path: '/finance/settlements', component: SettlementPayoutAdminView },
  ],
});

const app = createApp(App);
app.use(router);
app.mount('#app');

