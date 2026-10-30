package com.bookingnetwork.appmarket.core.installer

import android.content.Context
import android.content.pm.ApplicationInfo
import android.content.pm.PackageInfo
import android.content.pm.PackageManager
import android.os.Build
import androidx.core.content.pm.PackageInfoCompat

data class InstalledAppInfo(
    val packageName: String,
    val versionCode: Long,
    val versionName: String,
    val certSha256: String,
    val installingPackageName: String?
)

class InstalledPackagesScanner(private val context: Context) {

    fun scanInstalledApps(): List<InstalledAppInfo> {
        val pm = context.packageManager
        val flags = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.P) {
            PackageManager.GET_SIGNING_CERTIFICATES
        } else {
            @Suppress("DEPRECATION")
            PackageManager.GET_SIGNATURES
        }

        val packages = pm.getInstalledPackages(flags)
        val result = mutableListOf<InstalledAppInfo>()

        for (pkg in packages) {
            // Filter out pre-installed system apps unless updated by user
            val isSystemApp = (pkg.applicationInfo.flags and ApplicationInfo.FLAG_SYSTEM) != 0
            val isUpdatedSystemApp = (pkg.applicationInfo.flags and ApplicationInfo.FLAG_UPDATED_SYSTEM_APP) != 0
            if (isSystemApp && !isUpdatedSystemApp) {
                continue
            }

            val certSha256 = extractCertSha256(pkg) ?: continue
            val installer = getInstallerPackageName(pkg.packageName)
            val versionCode = PackageInfoCompat.getLongVersionCode(pkg)

            result.add(
                InstalledAppInfo(
                    packageName = pkg.packageName,
                    versionCode = versionCode,
                    versionName = pkg.versionName ?: "",
                    certSha256 = certSha256,
                    installingPackageName = installer
                )
            )
        }

        return result
    }

    private fun extractCertSha256(pkg: PackageInfo): String? {
        val certBytes = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.P) {
            val signingInfo = pkg.signingInfo ?: return null
            if (signingInfo.hasMultipleSigners()) {
                signingInfo.apkContentsSigners.firstOrNull()?.toByteArray()
            } else {
                signingInfo.signingCertificateHistory.lastOrNull()?.toByteArray()
            }
        } else {
            @Suppress("DEPRECATION")
            pkg.signatures?.firstOrNull()?.toByteArray()
        } ?: return null

        return ApkCertVerifier.computeCertSha256(certBytes)
    }

    private fun getInstallerPackageName(packageName: String): String? {
        return try {
            val pm = context.packageManager
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.R) {
                pm.getInstallSourceInfo(packageName).installingPackageName
            } else {
                @Suppress("DEPRECATION")
                pm.getInstallerPackageName(packageName)
            }
        } catch (e: Exception) {
            null
        }
    }
}
