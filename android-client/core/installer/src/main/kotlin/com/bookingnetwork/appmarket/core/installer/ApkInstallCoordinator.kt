package com.bookingnetwork.appmarket.core.installer

import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.content.pm.PackageInstaller
import android.content.pm.PackageManager
import android.os.Build
import android.os.storage.StorageManager
import com.bookingnetwork.appmarket.core.downloader.HttpRangeDownloader
import java.io.File

sealed class InstallResult {
    data class SessionCommitted(val sessionId: Int) : InstallResult()
    data class InsufficientStorage(val requiredBytes: Long, val allocatableBytes: Long) : InstallResult()
    data class ChecksumMismatch(val expected: String, val actual: String) : InstallResult()
    object SignerVerificationFailed : InstallResult()
    object RequiresUnknownSourcesPermission : InstallResult()
    data class InstallationFailed(val message: String, val cause: Throwable? = null) : InstallResult()
}

data class InstallArtifactParams(
    val packageName: String,
    val expectedSha256: String,
    val expectedSignerSha256: String,
    val fileSizeBytes: Long,
    val isSelfOwnedUpdate: Boolean = false
)

class ApkInstallCoordinator(
    private val context: Context,
    private val certVerifier: ApkCertVerifier = ApkCertVerifier(context)
) {
    fun installApk(
        stagedApkFile: File,
        artifact: InstallArtifactParams
    ): InstallResult {
        // 1. Storage Pre-check: 2.2x file size + 100MB buffer
        val storageManager = context.getSystemService(Context.STORAGE_SERVICE) as? StorageManager
        if (storageManager != null && Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val storageUuid = storageManager.getUuidForPath(stagedApkFile)
            val allocatable = storageManager.getAllocatableBytes(storageUuid)
            val requiredBytes = (artifact.fileSizeBytes * 22) / 10 + (100L * 1024 * 1024)
            if (allocatable < requiredBytes) {
                return InstallResult.InsufficientStorage(requiredBytes, allocatable)
            }
        }

        // 2. SHA-256 Hash Verification
        val actualSha256 = HttpRangeDownloader.computeFileSha256(stagedApkFile)
        if (!actualSha256.equals(artifact.expectedSha256.trim(), ignoreCase = true)) {
            stagedApkFile.delete()
            return InstallResult.ChecksumMismatch(expected = artifact.expectedSha256, actual = actualSha256)
        }

        // 3. Local APK Signer Certificate Verification
        val archiveCertSha256 = certVerifier.extractPrimarySignerSha256(stagedApkFile)
        if (archiveCertSha256 != null && !archiveCertSha256.equals(artifact.expectedSignerSha256.trim(), ignoreCase = true)) {
            stagedApkFile.delete()
            return InstallResult.SignerVerificationFailed
        }

        // 4. REQUEST_INSTALL_PACKAGES permission check (API 26+)
        val pm = context.packageManager
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O && !pm.canRequestPackageInstalls()) {
            return InstallResult.RequiresUnknownSourcesPermission
        }

        // 5. PackageInstaller.Session creation, stream copy, fsync, and commit
        return try {
            val packageInstaller = pm.packageInstaller
            val params = PackageInstaller.SessionParams(PackageInstaller.SessionParams.MODE_FULL_INSTALL).apply {
                setAppPackageName(artifact.packageName)
                setSize(stagedApkFile.length())
                if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S && artifact.isSelfOwnedUpdate) {
                    setRequireUserAction(PackageInstaller.SessionParams.USER_ACTION_NOT_REQUIRED)
                }
            }

            val sessionId = packageInstaller.createSession(params)
            packageInstaller.openSession(sessionId).use { session ->
                session.openWrite("base.apk", 0, stagedApkFile.length()).use { out ->
                    stagedApkFile.inputStream().use { input ->
                        input.copyTo(out, bufferSize = 64 * 1024)
                    }
                    session.fsync(out)
                }
                session.commit(createInstallIntentSender(sessionId, artifact.packageName))
            }

            InstallResult.SessionCommitted(sessionId)
        } catch (e: Exception) {
            InstallResult.InstallationFailed("PackageInstaller session commit failed: ${e.message}", e)
        }
    }

    private fun createInstallIntentSender(sessionId: Int, packageName: String): android.content.IntentSender {
        val intent = Intent("com.bookingnetwork.appmarket.ACTION_INSTALL_COMMIT_RESULT").apply {
            setPackage(context.packageName)
            putExtra(PackageInstaller.EXTRA_PACKAGE_NAME, packageName)
            putExtra(PackageInstaller.EXTRA_SESSION_ID, sessionId)
        }
        val flags = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_MUTABLE
        } else {
            PendingIntent.FLAG_UPDATE_CURRENT
        }
        return PendingIntent.getBroadcast(context, sessionId, intent, flags).intentSender
    }
}

