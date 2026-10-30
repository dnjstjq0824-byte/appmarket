import { createApp } from 'vue';
import { createRouter, createWebHistory } from 'vue-router';
import App from './App.vue';

import ApkUploaderView from './views/ApkUploaderView.vue';
import AppMetadataEditorView from './views/AppMetadataEditorView.vue';
import DeveloperApiKeysView from './views/DeveloperApiKeysView.vue';
import KybVerificationView from './views/KybVerificationView.vue';
import SettlementsDashboardView from './views/SettlementsDashboardView.vue';
import SigningLineageView from './views/SigningLineageView.vue';
import StagedRolloutView from './views/StagedRolloutView.vue';
import TeamMembersRbacView from './views/TeamMembersRbacView.vue';
import AdCampaignManagerView from './views/ads/AdCampaignManagerView.vue';
import AppPricingEditorView from './views/commerce/AppPricingEditorView.vue';
import PromoCodeManagerView from './views/commerce/PromoCodeManagerView.vue';

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: KybVerificationView },
    { path: '/releases/upload', component: ApkUploaderView },
    { path: '/apps/metadata', component: AppMetadataEditorView },
    { path: '/api-keys', component: DeveloperApiKeysView },
    { path: '/finance/settlements', component: SettlementsDashboardView },
    { path: '/releases/lineage', component: SigningLineageView },
    { path: '/releases/rollout', component: StagedRolloutView },
    { path: '/team', component: TeamMembersRbacView },
    { path: '/ads', component: AdCampaignManagerView },
    { path: '/commerce/pricing', component: AppPricingEditorView },
    { path: '/commerce/promos', component: PromoCodeManagerView },
  ],
});

const app = createApp(App);
app.use(router);
app.mount('#app');

