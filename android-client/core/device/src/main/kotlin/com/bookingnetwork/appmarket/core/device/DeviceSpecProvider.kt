package com.bookingnetwork.appmarket.core.device

import android.content.Context
import android.content.pm.PackageManager
import android.os.Build
import java.security.MessageDigest
import java.util.UUID

data class DeviceSpec(
    val sdkInt: Int,
    val supportedAbis: List<String>,
    val densityDpi: Int,
    val availableFeatures: Set<String>,
    val deviceHash: String
)

sealed class CompatibilityResult {
    object Compatible : CompatibilityResult()
    data class IncompatibleSdk(val requiredMinSdk: Int, val deviceSdk: Int) : CompatibilityResult()
    data class IncompatibleAbi(val requiredAbis: List<String>, val deviceAbis: List<String>) : CompatibilityResult()
    data class MissingFeatures(val missingFeatures: List<String>) : CompatibilityResult()
}

class DeviceSpecProvider(private val context: Context) {

    fun getDeviceSpec(): DeviceSpec {
        val pm = context.packageManager
        val features = pm.systemAvailableFeatures
            ?.mapNotNull { it.name }
            ?.toSet() ?: emptySet()

        val dpi = context.resources.displayMetrics.densityDpi
        val abis = Build.SUPPORTED_ABIS.toList()

        return DeviceSpec(
            sdkInt = Build.VERSION.SDK_INT,
            supportedAbis = abis,
            densityDpi = dpi,
            availableFeatures = features,
            deviceHash = getOrCreateDeviceHash()
        )
    }

    fun checkCompatibility(
        minSdk: Int,
        nativeAbis: List<String>,
        requiredFeatures: List<String>
    ): CompatibilityResult {
        val spec = getDeviceSpec()

        if (spec.sdkInt < minSdk) {
            return CompatibilityResult.IncompatibleSdk(minSdk, spec.sdkInt)
        }

        if (nativeAbis.isNotEmpty()) {
            val hasCommonAbi = nativeAbis.any { candAbi ->
                spec.supportedAbis.any { devAbi -> devAbi.equals(candAbi, ignoreCase = true) }
            }
            if (!hasCommonAbi) {
                return CompatibilityResult.IncompatibleAbi(nativeAbis, spec.supportedAbis)
            }
        }

        if (requiredFeatures.isNotEmpty()) {
            val missing = requiredFeatures.filter { feat -> !spec.availableFeatures.contains(feat) }
            if (missing.isNotEmpty()) {
                return CompatibilityResult.MissingFeatures(missing)
            }
        }

        return CompatibilityResult.Compatible
    }

    private fun getOrCreateDeviceHash(): String {
        val prefs = context.getSharedPreferences("bn_device_identity", Context.MODE_PRIVATE)
        var hash = prefs.getString("device_hash", null)
        if (hash == null) {
            val raw = "${Build.BRAND}:${Build.MODEL}:${UUID.randomUUID()}"
            val md = MessageDigest.getInstance("SHA-256")
            val digest = md.digest(raw.toByteArray(Charsets.UTF_8))
            hash = digest.joinToString("") { "%02x".format(it) }
            prefs.edit().putString("device_hash", hash).apply()
        }
        return hash
    }
}

