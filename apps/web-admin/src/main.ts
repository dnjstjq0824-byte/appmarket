import { createApp } from 'vue';
import { createRouter, createWebHistory } from 'vue-router';
import '@bookingnetwork/ui-kit/src/tokens.css';
import { isAuthenticated } from '@bookingnetwork/api-client';
import App from './App.vue';

import AdminOverviewDashboard from './views/AdminOverviewDashboard.vue';
import AdminRbacDualControlView from './views/AdminRbacDualControlView.vue';
import ApkSecurityReviewView from './views/ApkSecurityReviewView.vue';
import AppMetadataReviewView from './views/AppMetadataReviewView.vue';
import AuditLogChainExplorerView from './views/AuditLogChainExplorerView.vue';
import EmergencyTakedownView from './views/EmergencyTakedownView.vue';
import ExternalConnectorsAdminView from './views/ExternalConnectorsAdminView.vue';
import KybReviewQueueView from './views/KybReviewQueueView.vue';
import LoginView from './views/LoginView.vue';
import ModerationReportsQueueView from './views/ModerationReportsQueueView.vue';
import AdCampaignReviewView from './views/AdCampaignReviewView.vue';
import NotFoundView from './views/NotFoundView.vue';
import OrdersRefundsAdminView from './views/OrdersRefundsAdminView.vue';
import PlatformAnalyticsAdminView from './views/PlatformAnalyticsAdminView.vue';
import SettlementPayoutAdminView from './views/SettlementPayoutAdminView.vue';

const router = createRouter({
  history: createWebHistory('/admin/'),
  routes: [
    { path: '/login', component: LoginView },
    { path: '/', component: AdminOverviewDashboard },
    { path: '/rbac', component: AdminRbacDualControlView },
    { path: '/security/reviews', component: ApkSecurityReviewView },
    { path: '/apps/metadata', component: AppMetadataReviewView },
    { path: '/audit', component: AuditLogChainExplorerView },
    { path: '/security/kill-switch', component: EmergencyTakedownView },
    { path: '/connectors', component: ExternalConnectorsAdminView },
    { path: '/verifications', component: KybReviewQueueView },
    { path: '/moderation', component: ModerationReportsQueueView },
    { path: '/ads', component: AdCampaignReviewView },
    { path: '/finance/orders', component: OrdersRefundsAdminView },
    { path: '/analytics', component: PlatformAnalyticsAdminView },
    { path: '/finance/settlements', component: SettlementPayoutAdminView },
    { path: '/:pathMatch(.*)*', component: NotFoundView },
  ],
});

router.beforeEach((to) => {
  if (to.path === '/login') {
    return true;
  }
  if (!isAuthenticated()) {
    return { path: '/login', query: { redirect: to.fullPath } };
  }
  return true;
});

const app = createApp(App);
app.use(router);
app.mount('#app');
