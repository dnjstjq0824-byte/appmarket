<template>
  <span class="bn-sponsored-badge" role="note" aria-label="Sponsored Advertising">
    광고 · Sponsored
  </span>
</template>

<script setup lang="ts">
// In accordance with Section 2 of DESIGN_SYSTEM.md:
// High-contrast (>= 7:1) badge for advertisement disclosure that cannot be hidden by user styling.
</script>

<style scoped>
.bn-sponsored-badge {
  display: inline-block !important;
  visibility: visible !important;
  opacity: 1 !important;
  font-size: 11px !important;
  font-weight: 700 !important;
  text-transform: uppercase !important;
  letter-spacing: 0.5px !important;
  padding: 2px 7px !important;
  border-radius: 4px !important;
  background-color: #0F172A !important;
  color: #F8FAFC !important;
  border: 1px solid #334155 !important;
  line-height: 1.4 !important;
}
</style>

