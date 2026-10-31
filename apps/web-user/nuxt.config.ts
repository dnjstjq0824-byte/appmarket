// https://nuxt.com/docs/api/configuration/nuxt-config
import { defineNuxtConfig } from 'nuxt/config'

export default defineNuxtConfig({
  compatibilityDate: '2024-11-01',
  devtools: { enabled: false },
  css: ['@bookingnetwork/ui-kit/src/tokens.css'],
  nitro: {
    prerender: {
      crawlLinks: true,
      failOnError: false,
      ignore: ['/admin', '/admin/**', '/developer', '/developer/**']
    }
  },
  app: {
    head: {
      title: 'Booking Network App Market — Universal Independent App Store',
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'description', content: 'Global Independent App Store and Federated Application Discovery Engine' }
      ]
    }
  }
})
