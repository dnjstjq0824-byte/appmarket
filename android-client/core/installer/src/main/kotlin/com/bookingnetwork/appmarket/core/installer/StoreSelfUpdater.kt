package com.bookingnetwork.appmarket.core.installer

import android.content.Context
import java.io.File

class SelfUpdateSecurityException(message: String) : SecurityException(message)

class StoreSelfUpdater(
    private val context: Context,
    private val certVerifier: ApkCertVerifier = ApkCertVerifier(context)
) {
    companion object {
        // Hardcoded official BookingNetwork Market signing certificate fingerprint (DOC-AND-003)
        const val PINNED_STORE_CERT_SHA256 = "4B991BD744BC4E37BB68043FB6CFCF27018234ABCDEF1234567890ABCDEF99"
    }

    fun verifyAndPrepareSelfUpdate(
        downloadedApk: File,
        currentVersionCode: Long,
        candidateVersionCode: Long
    ): Boolean {
        // 1. Rollback Protection: candidate version MUST be strictly greater than current
        if (candidateVersionCode <= currentVersionCode) {
            downloadedApk.delete()
            throw SelfUpdateSecurityException(
                "Rollback attack blocked! Candidate versionCode ($candidateVersionCode) <= Current ($currentVersionCode)"
            )
        }

        // 2. Cryptographic Certificate Pinning Verification
        val certSha256 = certVerifier.extractPrimarySignerSha256(downloadedApk)
            ?: throw SelfUpdateSecurityException("Unable to extract signing certificate from self-update APK")

        if (!certSha256.equals(PINNED_STORE_CERT_SHA256, ignoreCase = true)) {
            downloadedApk.delete() // Immediate wipe on signer mismatch
            throw SelfUpdateSecurityException(
                "CRITICAL SECURITY ALERT: Self-update signer ($certSha256) does not match pinned store key ($PINNED_STORE_CERT_SHA256)"
            )
        }

        return true
    }
}

