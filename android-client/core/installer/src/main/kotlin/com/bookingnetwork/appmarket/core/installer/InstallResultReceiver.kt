package com.bookingnetwork.appmarket.core.installer

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.content.pm.PackageInstaller
import android.util.Log

class InstallResultReceiver : BroadcastReceiver() {

    override fun onReceive(context: Context, intent: Intent) {
        val status = intent.getIntExtra(PackageInstaller.EXTRA_STATUS, PackageInstaller.STATUS_FAILURE)
        val packageName = intent.getStringExtra(PackageInstaller.EXTRA_PACKAGE_NAME) ?: "unknown"
        val message = intent.getStringExtra(PackageInstaller.EXTRA_STATUS_MESSAGE) ?: ""

        when (status) {
            PackageInstaller.STATUS_PENDING_USER_ACTION -> {
                // User confirmation dialog needed (system prompt)
                val confirmIntent = intent.getParcelableExtra<Intent>(Intent.EXTRA_INTENT)
                if (confirmIntent != null) {
                    confirmIntent.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
                    context.startActivity(confirmIntent)
                }
            }

            PackageInstaller.STATUS_SUCCESS -> {
                Log.i("BNInstall", "Package $packageName installed successfully.")
                // Notify app state update
                val successIntent = Intent("com.bookingnetwork.appmarket.INSTALL_COMPLETED").apply {
                    putExtra("package_name", packageName)
                    putExtra("status", "SUCCESS")
                }
                context.sendBroadcast(successIntent)
            }

            PackageInstaller.STATUS_FAILURE_ABORTED -> {
                Log.w("BNInstall", "Installation cancelled by user for $packageName.")
            }

            PackageInstaller.STATUS_FAILURE_CONFLICT -> {
                // Signature conflict (INSTALL_FAILED_UPDATE_INCOMPATIBLE)
                Log.e("BNInstall", "Signature conflict for $packageName: $message")
                val conflictIntent = Intent("com.bookingnetwork.appmarket.INSTALL_SIGNATURE_CONFLICT").apply {
                    putExtra("package_name", packageName)
                    putExtra("message", message)
                }
                context.sendBroadcast(conflictIntent)
            }

            PackageInstaller.STATUS_FAILURE_STORAGE -> {
                Log.e("BNInstall", "Insufficient storage during installation for $packageName.")
            }

            PackageInstaller.STATUS_FAILURE_INVALID -> {
                Log.e("BNInstall", "Invalid APK or version downgrade blocked for $packageName.")
            }

            PackageInstaller.STATUS_FAILURE_INCOMPATIBLE -> {
                Log.e("BNInstall", "Incompatible device ABI or SDK for $packageName.")
            }

            else -> {
                Log.e("BNInstall", "Installation failed for $packageName: status=$status message=$message")
            }
        }
    }
}
