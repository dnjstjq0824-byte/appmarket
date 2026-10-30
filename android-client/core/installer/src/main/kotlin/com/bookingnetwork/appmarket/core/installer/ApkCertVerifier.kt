package com.bookingnetwork.appmarket.core.installer

import android.content.Context
import android.content.pm.PackageManager
import android.os.Build
import java.io.File
import java.security.MessageDigest
import java.security.cert.CertificateFactory
import java.security.cert.X509Certificate

class ApkCertVerifier(private val context: Context) {

    fun extractPrimarySignerSha256(apkFile: File): String? {
        val pm = context.packageManager
        val flags = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.P) {
            PackageManager.GET_SIGNING_CERTIFICATES
        } else {
            @Suppress("DEPRECATION")
            PackageManager.GET_SIGNATURES
        }

        val pkgInfo = pm.getPackageArchiveInfo(apkFile.absolutePath, flags) ?: return null

        val certBytes = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.P) {
            val signingInfo = pkgInfo.signingInfo ?: return null
            if (signingInfo.hasMultipleSigners()) {
                signingInfo.apkContentsSigners.firstOrNull()?.toByteArray()
            } else {
                signingInfo.signingCertificateHistory.lastOrNull()?.toByteArray()
            }
        } else {
            @Suppress("DEPRECATION")
            pkgInfo.signatures?.firstOrNull()?.toByteArray()
        } ?: return null

        return computeCertSha256(certBytes)
    }

    companion object {
        fun computeCertSha256(certBytes: ByteArray): String {
            val md = MessageDigest.getInstance("SHA-256")
            val digest = md.digest(certBytes)
            return digest.joinToString("") { "%02x".format(it) }
        }
    }
}

